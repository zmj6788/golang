package models

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

type WeatherCity struct {
	Reason string `json:"reason"`
	Result []struct {
		ID       string `json:"id"`
		Province string `json:"province"`
		City     string `json:"city"`
	} `json:"result"`
}

const weathercityfile = "./data/weathercity.json"

var WeatherCityData WeatherCity

func IninWeatherCityData() {
	w := &WeatherCity{}
	byteData, err := os.ReadFile(weathercityfile)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = json.Unmarshal(byteData, &w)
	if err != nil {
		logrus.Error(err)
		return
	}
	WeatherCityData = *w
	logrus.Info("citylist init success")
}
