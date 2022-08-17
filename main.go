package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	datestrings, indexes := getDates()
	fmt.Println(indexes)

	aachen := Standort{lat: 50.77, long: 6.1, name: "Aachen"}
	leipzig := Standort{lat: 51.34, long: 12.37, name: "Leipzig"}

	var standorte []Standort
	standorte = append(standorte, aachen, leipzig)

	for _, standort := range standorte {
		for _, datestring := range datestrings {

			url := fmt.Sprintf("https://api.brightsky.dev/weather?lat=%.2f&lon=%.2f&date=%s", standort.lat, standort.long, datestring)
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
			fmt.Printf("Max Temperature in %s on the %s was %.2f \n", standort.name, datestring, maxTemp)
		}
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

func getDates() ([]string, []string) {
	f, err := excelize.OpenFile("Book1.xlsx")
	if err != nil {
		fmt.Println(err)

	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err := f.GetRows("Tabelle1")
	if err != nil {
		fmt.Println(err)

	}

	var tempIndex int
	var dateIndex int

	for indexRowOne, rowOne := range rows[0] {
		if rowOne == "Temperature" {
			tempIndex = indexRowOne
		}
		if rowOne == "Reported Date" {
			dateIndex = indexRowOne
		}
	}

	rows = rows[1:]
	var datestrings []string
	var indexes []string
	for index, row := range rows {

		if len(row) > dateIndex && len(row[dateIndex]) > 0 {
			if len(row) > tempIndex && len(row[tempIndex]) == 0 {

				datestrings = append(datestrings, row[dateIndex])
				coord, err := excelize.CoordinatesToCellName(dateIndex+1, index+1)
				if err != nil {
					fmt.Println(err)
				}
				indexes = append(indexes, coord)
			}
			if len(row) <= tempIndex {

				datestrings = append(datestrings, row[dateIndex])
				coord, err := excelize.CoordinatesToCellName(dateIndex+1, index+1)
				if err != nil {
					fmt.Println(err)
				}
				indexes = append(indexes, coord)
			}

		}

	}

	return datestrings, indexes
}
