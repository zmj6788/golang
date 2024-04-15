package models

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

type CityList struct {
	Reason string `json:"reason"`
	Result []struct {
		ID       string `json:"id"`
		Province string `json:"province"`
		City     string `json:"city"`
		District string `json:"district"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}

var Citys CityList

const citylistfile = "./data/citylist.json"

func InitCityListJSon() {
	c := &CityList{}
	byteData, err := os.ReadFile(citylistfile)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = json.Unmarshal(byteData, &c)
	if err != nil {
		logrus.Error(err)
		return
	}
	Citys = *c
	logrus.Info("citylist init success")
}
