package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
)

type CatBreed struct {
	Breed   string `json:"breed"`
	Origin  string `json:"origin"`
	Coat    string `json:"coat"`
	Pattern string `json:"pattern"`
	Country string `json:"country,omitempty"`
}
type Payload struct {
	Str string `json:"str"`
}
type CatBreedsByCountry map[string][]CatBreed

func main() {
	r := gin.Default()
	r.GET("/cat-breeds", Getcatbreeds)
	r.POST("/post-words", Countwords)
	r.Run(":8080")
}
func Countwords(c *gin.Context) {
	var payload Payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}
	wordCount := countWords(payload.Str)
	if wordCount >= 8 {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	} else {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "Not Acceptable"})
	}
}
func countWords(str string) int {
	wordsRegex := regexp.MustCompile(`\b\w+\b`)
	words := wordsRegex.FindAllString(str, -1)
	return len(words)
}
func Getcatbreeds(c *gin.Context) {
	catBreeds, err := getAllCatBreeds()
	if err != nil {
		log.Println("Failed to retrieve cat breeds:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cat breeds"})
		return
	}
	//Return cat breeds grouped by Country
	catBreedsByCountry := groupCatBreedsByCountry(catBreeds)
	c.JSON(http.StatusOK, catBreedsByCountry)
}

func logResponseToFile(responseText string) {
	file, err := os.OpenFile("response.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println("Failed to create response.txt:", err)
	}
	defer file.Close()
	_, err = file.WriteString(responseText + "\n")
	if err != nil {
		log.Println("Failed to write to response.txt:", err)
	}
}

func getAllCatBreeds() ([]CatBreed, error) {
	pages := 1
	var allCatBreeds []CatBreed
	for i := 1; i <= pages; i++ {
		pageCatBreeds, pageno, err := getCatBreedsByPage(i)
		if err != nil {
			return nil, err
		}
		if i == 1 {
			pages = pageno
			log.Println("total pages: ", pageno)
		}
		allCatBreeds = append(allCatBreeds, pageCatBreeds...)
	}
	return allCatBreeds, nil
}
func getCatBreedsByPage(page int) ([]CatBreed, int, error) {
	url := fmt.Sprintf("https://catfact.ninja/breeds?page=%d", page)
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	var data struct {
		Data []CatBreed `json:"data"`
		Page int        `json:"last_page"`
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	logResponseToFile(string(body))
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, 0, err
	}
	return data.Data, data.Page, nil
}
func groupCatBreedsByCountry(catBreeds []CatBreed) CatBreedsByCountry {
	catBreedsByCountry := make(CatBreedsByCountry)
	for _, breed := range catBreeds {
		country := breed.Country
		breed.Country = ""
		catBreedsByCountry[country] = append(catBreedsByCountry[country], breed)
	}
	return catBreedsByCountry
}
