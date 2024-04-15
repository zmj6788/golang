package core

import (
	"fmt"
	"gvb_server/config"
	"gvb_server/global"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// 读取配置操作
// 读取yaml文件的配置
func InitConf() {
	const configFile = "settings.yaml"
	c := &config.Config{}
	yamlConfig, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Errorf("read config file error, %v", err))
	}
	err = yaml.Unmarshal(yamlConfig, c)
	if err != nil {
		log.Fatal("unmarshal config file error, %v", err)
	}
	log.Println("config yamlFile load Init success.")
	global.Config = c
}
