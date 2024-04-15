package api

import (
	"gvb_server/api/city_api"
	"gvb_server/api/user_api"
	"gvb_server/api/weather_api"
)

type ApiGroup struct {
	WeatherApi weather_api.WeatherApi
	CityApi    city_api.CityApi
	UserApi    user_api.UserApi
}

var ApiGroupApp = new(ApiGroup)
