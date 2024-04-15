package models

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

type WeatherData struct {
	Reason      string `json:"reason"`
	Weatherdata []struct {
		Result struct {
			City     string `json:"city"`
			Realtime struct {
				Temperature string `json:"temperature"`
				Humidity    string `json:"humidity"`
				Info        string `json:"info"`
				Wid         string `json:"wid"`
				Direct      string `json:"direct"`
				Power       string `json:"power"`
				Aqi         string `json:"aqi"`
			} `json:"realtime"`
			Future []struct {
				Date        string `json:"date"`
				Temperature string `json:"temperature"`
				Weather     string `json:"weather"`
				Wid         struct {
					Day   string `json:"day"`
					Night string `json:"night"`
				} `json:"wid"`
				Direct string `json:"direct"`
			} `json:"future"`
		} `json:"result"`
		ErrorCode int    `json:"error_code"`
		Reason    string `json:"reason,omitempty"`
	} `json:"weatherdata"`
}

type Weather struct {
	City        string
	Temperature string
	Humidity    string
	Info        string
	Wid         string
	Direct      string
	Power       string
	Aqi         string
	WeatherWeek WeatherWeek
}
type WeatherDay struct {
	Date        string
	Temperature string
	Weather     string
	WidDay      string
	WidNight    string
	Direct      string
}
type WeatherWeek [5]WeatherDay

var Weather_Data WeatherData
var WeatherMap = make(map[string]Weather)

const weatherfile = "./data/weather.json"

func IninWeatherData() {
	w := &WeatherData{}

	byteData, err := os.ReadFile(weatherfile)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = json.Unmarshal(byteData, &w)
	if err != nil {
		logrus.Error(err)
		return
	}
	Weather_Data = *w

	for _, v := range Weather_Data.Weatherdata {
		var weekWeather WeatherWeek
		for i, future := range v.Result.Future {
			weekWeather[i] = WeatherDay{
				Date:        future.Date,
				Temperature: future.Temperature,
				Weather:     future.Weather,
				WidDay:      future.Wid.Day,
				WidNight:    future.Wid.Night,
				Direct:      future.Direct,
			}
		}
		WeatherMap[v.Result.City] = Weather{
			City:        v.Result.City,
			Temperature: v.Result.Realtime.Temperature,
			Humidity:    v.Result.Realtime.Humidity,
			Info:        v.Result.Realtime.Info,
			Wid:         v.Result.Realtime.Wid,
			Direct:      v.Result.Realtime.Direct,
			Power:       v.Result.Realtime.Power,
			Aqi:         v.Result.Realtime.Aqi,
			WeatherWeek: weekWeather,
		}
	}

	logrus.Info("weather data init success")

}
