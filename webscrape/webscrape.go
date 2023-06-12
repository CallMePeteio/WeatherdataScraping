package webscrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

var print = fmt.Println

type item struct {
	Date    string `json:"date"`
	Wind    string `json:"windMS"`
	Rain    string `json:"rainMM"`
	TempMax string `json:"tempMax"`
	TempMin string `json:"tempMin"`
}

func convertDateInt(date string) (int, error) {

	dateSlice := strings.Split(date, ".")

	dayVal, dayErr := strconv.Atoi(dateSlice[0])
	monthVal, monthErr := strconv.Atoi(dateSlice[1])
	yearVal, yearErr := strconv.Atoi(dateSlice[2])

	if dayErr != nil || monthErr != nil || yearErr != nil {
		fmt.Println("Error with converting string to int convertDateInt func: ", dayErr, monthErr, yearVal)
		time.Sleep(1 * time.Second)
		return 0, errors.New("There was an error converting string to integer")
	}
	return dayVal + monthVal*31 + yearVal*365, nil
}

func GetJsonData(path string) ([]item, error) {
	jsonData, err := os.Open(path) // OPENS THE JSON FILE

	if err != nil { // CHECKS FOR ERRORS
		fmt.Println("Error Reading json file: ", err)
		time.Sleep(1 * time.Second)
		return nil, errors.New("There Was an error reading the previous weather data")
	}

	defer jsonData.Close()               // CLOSES THE JSON FILE
	bytes, _ := ioutil.ReadAll(jsonData) // RETURNS THE BYTE ARRAY FOR, USED FOR MAKING THE SLICE

	var prevData []item                                      // INITIALIZES THE OLD DATA SLICE
	if err := json.Unmarshal(bytes, &prevData); err != nil { // UNMARSHALS THE BYTE ARRAY (CONVERTS THE BYTEARRAY TO A SLIC)
		fmt.Println("Error Unmasheling: ", err)
		time.Sleep(1 * time.Second)
		return nil, errors.New("There was an error unmashaling the byte array form the json file")
	}
	return prevData, nil
}

func convertMonthInt(dateText string) (string, error) {

	monthMap := map[string]string{
		"januar":    "1",
		"februar":   "2",
		"mars":      "3",
		"april":     "4",
		"mai":       "5",
		"juni":      "6",
		"juli":      "7",
		"august":    "8",
		"september": "9",
		"oktober":   "10",
		"november":  "11",
		"desember":  "12",
	}

	dateText = strings.ToLower(dateText) // CONVERTS THE TEXT TO LOWERCASE

	for key, value := range monthMap { // LOOPS OVER ALL OF THE VALUES/KEYS IN THE MONTHMAP TABLE
		if strings.Index(dateText, key) != -1 { // IF THE KEY EXISTS IN THE STRING
			return value, nil // RETURN THE VALUE
		}
	}
	return "error", errors.New("There Was an error converting the month to integer")
}

func correctData(items []item) []item {

	for i := 0; i < len(items); i++ {
		re := regexp.MustCompile(`([-+]?[0-9]*,[0-9]+)|\d+`) // DOES SOME MAGIC

		day := re.FindAllString(items[i].Date, -1)       // FINDS THE INT IN THE DATE ITEM
		wind := re.FindAllString(items[i].Wind, -1)      // FINDS THE INT IN THE WIND ITEM
		temp := re.FindAllString(items[i].TempMax, -1)   // FINDS THE INT IN THE TEMP ITEM
		rain := re.FindAllString(items[i].Rain, -1)      // FINDS THE FLOAT IN THE RAIN ITEM
		rain[0] = strings.Replace(rain[0], ",", ".", -1) // CONVERTS THE "," TO "."

		monthInt, errMonthInt := convertMonthInt(items[i].Date) // CONVERTS THE STRING MONTH TO AN INTEGER THAT IS STILL A  STRING
		year := strconv.Itoa(time.Now().Year())                 // GETS THE CURRENT YEAR, AND CONVERTS IT TO A STRING

		if errMonthInt == nil && len(temp) >= 2 {
			items[i].Date = day[0] + "." + monthInt + "." + year
			items[i].Wind = wind[0]
			items[i].Rain = rain[0]
			items[i].TempMax = temp[0]
			items[i].TempMin = temp[1]

		} else {
			errMsg := errors.New("Problem converting Monrth to int or the length of the temperatures")
			fmt.Printf("%+, \n Month: %v Temp Length: %v \n \n", errMsg, errMonthInt, len(temp))
			time.Sleep(1 * time.Second)
		}

	}
	return items
}

func GetItems() error {

	var items []item
	c := colly.NewCollector() // ININTIALIZES THE WEBSCRAPER

	c.OnHTML("li.daily-weather-list-item", func(h *colly.HTMLElement) {
		item := item{
			Date:    h.ChildText("div.daily-weather-list-item__date-and-warnings"),
			Wind:    h.ChildText("div.daily-weather-list-item__wind"),
			Rain:    h.ChildText("div.daily-weather-list-item__precipitation"),
			TempMax: h.ChildText("div.daily-weather-list-item__temperature"),
			TempMin: "Nil", // GETS ADDED LATER IN THE "correctDate" FUNCTION (GETS IT FROM TempMax)
		}
		items = append(items, item)

	})
	c.Visit("https://www.yr.no/nb/v%C3%A6rvarsel/daglig-tabell/1-269359/Norge/Nordland/Bod%C3%B8/Bod%C3%B8")

	items = correctData(items)
	prevData, _ := GetJsonData("weatherData.json") // GETS THE PREVIOUS DATA IN THE JSON FILE

	var iCounter int
	var currentData []item
	for _, data := range prevData { // LOOPS OVER THE PREVIOUS DATA
		prevDateInt, _ := convertDateInt(data.Date)        // CONVERTS THE DATE TO INTEGER FORM
		dateInt, _ := convertDateInt(items[iCounter].Date) // CONVERTS THE DATE TO INTEGER FORM

		if dateInt <= prevDateInt { // IF THE PREVIOUS DATE IS GREATER THAN THE CURRENT DATE
			currentData = append(currentData, items[iCounter]) // APPENDS THE DATA TO THE MAINDATA SLICE
			iCounter += 1                                      // ADDS ONE TO THE INDEX COUNTER

			if iCounter >= len(items) { // IF WE ARE DONE ITERATING OVER ALL OF THE DIFFRENT ITEMS GATHERED
				break
			}
		} else {
			currentData = append(currentData, data)
		}
	}

	content, err := json.Marshal(currentData)

	if err != nil {
		print("There was an error writing to the JSON File")
		return errors.New("There was an error writing to the JSON File")
	}

	os.WriteFile("weatherData.json", content, 0644)

	return nil

}
