package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	scrape "webScraper/webscrape"

	"github.com/gin-gonic/gin"
)

var print = fmt.Println

type logStruct struct {
	Action  string `json:"action"`
	Date    string `json:"date"`
	Sucsess bool   `json:"sucsess"`
}

func getJsonData_(path string) ([]logStruct, error) {
	jsonData, err := os.Open(path) // OPENS THE JSON FILE

	if err != nil { // CHECKS FOR ERRORS
		fmt.Println("Error Reading json file: ", err)
		time.Sleep(1 * time.Second)
		return nil, errors.New("There Was an error reading the previous weather data")
	}

	defer jsonData.Close()               // CLOSES THE JSON FILE
	bytes, _ := ioutil.ReadAll(jsonData) // RETURNS THE BYTE ARRAY FOR, USED FOR MAKING THE SLICE

	var prevData []logStruct                                 // INITIALIZES THE OLD DATA SLICE
	if err := json.Unmarshal(bytes, &prevData); err != nil { // UNMARSHALS THE BYTE ARRAY (CONVERTS THE BYTEARRAY TO A SLIC)
		fmt.Println("Error Unmasheling: ", err)
		time.Sleep(1 * time.Second)
		return nil, errors.New("There was an error unmashaling the byte array form the json file")
	}
	return prevData, nil
}

func WebScrapeNewData() {
	var log logStruct

	for {
		scrapeErr := scrape.GetItems() // WEBSCRAPES THE DATA

		if scrapeErr != nil { // IF THERE WAS AN ERROR WEBSCRAPING THE DATA
			print("There Was an error getting new data: ", scrapeErr)
			log = logStruct{Action: "Webscraping", Date: time.Now().Format("02-01-2006 15:04:05"), Sucsess: false} // LOGS THE ERROR IN THE "log.json" FILE
		} else {
			print("Sucsessfully webscraped new Data", time.Now().Format("02-01-2006 15:04:05"))
			log = logStruct{Action: "Webscraping", Date: time.Now().Format("02-01-2006 15:04:05"), Sucsess: true} // LOGS THE SUCSESS MESSAGE TO THE "log.json" FILE
		}

		prevLogs, readJsonErr := getJsonData_("log.json")
		prevLogs = append(prevLogs, log)
		content, logErr := json.Marshal(log) // RETURNS A BYTE ARRAY FROM THE LOG SLICE

		if logErr == nil && readJsonErr == nil { // IF THERE WASNT ANY ERRORS
			os.WriteFile("log.json", content, 0644) // WRITES THE CONTENT TO THE JOSN FILE
		} else {
			print("There was an error writing to the log json file!", logErr, readJsonErr) // PRINTS THE ERROR
		}

		time.Sleep(10 * time.Second)
		//time.Sleep(86400 * time.Second) // SLEEPS FOR A DAY
	}
}

func getItems(c *gin.Context) {
	gatheredData, err := scrape.GetJsonData("weatherData.json") // GETS THE DATA GATRGERED IN THE JSON FILE

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Failed Getting Data Gathered In The Json File"})
		return
	}
	c.IndentedJSON(http.StatusOK, gatheredData)
}

func main() {

	go WebScrapeNewData() // WEBSCRAPES THE DATA ONCE A DAY, ON A SEPERATE THREAD

	router := gin.Default()

	router.GET("/getItems", getItems) // ADDS THE PATH TO GET THE ITEMS SCRAPED
	router.Run("localhost:8080")      // RUNS THE WEBSERVER

}
