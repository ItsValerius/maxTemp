package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type Response struct {
	Weather []struct {
		Timestamp         time.Time   `json:"timestamp"`
		SourceID          int         `json:"source_id"`
		Precipitation     float64     `json:"precipitation"`
		PressureMsl       float64     `json:"pressure_msl"`
		Sunshine          interface{} `json:"sunshine"`
		Temperature       float64     `json:"temperature"`
		WindDirection     int         `json:"wind_direction"`
		WindSpeed         float64     `json:"wind_speed"`
		CloudCover        int         `json:"cloud_cover"`
		DewPoint          float64     `json:"dew_point"`
		RelativeHumidity  int         `json:"relative_humidity"`
		Visibility        int         `json:"visibility"`
		WindGustDirection int         `json:"wind_gust_direction"`
		WindGustSpeed     float64     `json:"wind_gust_speed"`
		Condition         string      `json:"condition"`
		Icon              string      `json:"icon"`
	} `json:"weather"`
	Sources []struct {
		ID              int       `json:"id"`
		DwdStationID    string    `json:"dwd_station_id"`
		ObservationType string    `json:"observation_type"`
		Lat             float64   `json:"lat"`
		Lon             float64   `json:"lon"`
		Height          float64   `json:"height"`
		StationName     string    `json:"station_name"`
		WmoStationID    string    `json:"wmo_station_id"`
		FirstRecord     time.Time `json:"first_record"`
		LastRecord      time.Time `json:"last_record"`
		Distance        float64   `json:"distance"`
	} `json:"sources"`
}

type Standort struct {
	lat  float64
	long float64
	name string
}

func main() {
	f, err := excelize.OpenFile("weekly incident report.xlsx")
	if err != nil {
		fmt.Println(err)

	}

	datestrings, indexes := getDates()

	aachen := Standort{lat: 50.77, long: 6.1, name: "Aachen"}
	leipzig := Standort{lat: 51.34, long: 12.37, name: "Leipzig"}

	counter := 0

	for index, datestring := range datestrings {
		var url string
		url = fmt.Sprintf("https://api.brightsky.dev/weather?lat=%.2f&lon=%.2f&date=%s", leipzig.lat, leipzig.long, datestring)

		if indexes[index].Number >= 13 && indexes[index].Number <= 16 {
			url = fmt.Sprintf("https://api.brightsky.dev/weather?lat=%.2f&lon=%.2f&date=%s", aachen.lat, aachen.long, datestring)
		}

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		var result Response
		if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
			fmt.Println("Can not unmarshal JSON")
		}
		maxTemp := findMax(result)
		fmt.Printf("On %s the max temperature was: %.2f°C for number %d \n", datestring, maxTemp, indexes[index].Number)
		currYearSheet, err := findCurrYear(f)
		if err != nil {
			fmt.Println(err)
			return
		}
		f.SetCellFloat(currYearSheet, indexes[index].Axis, maxTemp, 2, 64)
		counter++
		if counter >= 5 {
			<-time.After(time.Second * 5)
			counter = 0
		}
	}
	f.Save()
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}
}

func findMax(response Response) float64 {
	weather := response.Weather
	maxTemp := weather[0].Temperature
	for i := 1; i < len(weather); i++ {
		if weather[i].Temperature > maxTemp {
			maxTemp = weather[i].Temperature
		}
	}

	return maxTemp
}

func findCurrYear(f *excelize.File) (string, error) {
	sheets := f.GetSheetList()
	t := time.Now()
	currYearStr := strconv.Itoa(t.Year())
	for _, sheet := range sheets {
		if sheet == currYearStr {
			return sheet, nil
		}
	}
	return "", errors.New("no sheet with year found")
}

type rowValue struct {
	Axis   string
	Number int
}

func getDates() ([]string, []rowValue) {
	f, err := excelize.OpenFile("weekly incident report.xlsx")
	if err != nil {
		fmt.Println(err)

	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	currYearSheet, err := findCurrYear(f)
	if err != nil {
		fmt.Println(err)

	}
	rows, err := f.GetRows(currYearSheet)
	if err != nil {
		fmt.Println(err)

	}

	var tempIndex int
	var dateIndex int
	var numberIndex int

	for indexRowOne, rowOne := range rows[0] {
		if rowOne == "Temperature" {
			tempIndex = indexRowOne
		}
		if rowOne == "Reported Date" {
			dateIndex = indexRowOne
		}
		if rowOne == "Epsilon Number" {
			numberIndex = indexRowOne
		}
	}

	var datestrings []string
	var rowValues []rowValue

	for index, row := range rows {

		if len(row) > dateIndex && len(row[dateIndex]) > 0 {
			if len(row) > tempIndex && len(row[tempIndex]) == 0 {

				datestrings = append(datestrings, row[dateIndex])
				axis, err := excelize.CoordinatesToCellName(tempIndex+1, index+1)
				if err != nil {
					fmt.Println(err)
				}
				number, err := strconv.Atoi(row[numberIndex])
				if err != nil {
					fmt.Println(err)
				}
				rowValue := rowValue{Axis: axis, Number: number}
				rowValues = append(rowValues, rowValue)
			}
			if len(row) <= tempIndex {

				datestrings = append(datestrings, row[dateIndex])
				axis, err := excelize.CoordinatesToCellName(tempIndex+1, index+1)
				if err != nil {
					fmt.Println(err)
				}

				number, err := strconv.Atoi(row[numberIndex])
				if err != nil {
					fmt.Println(err)
				}
				rowValue := rowValue{Axis: axis, Number: number}
				rowValues = append(rowValues, rowValue)
			}

		}

	}

	return datestrings, rowValues
}
