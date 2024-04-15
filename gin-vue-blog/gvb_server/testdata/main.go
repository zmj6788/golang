package main

import (
	"encoding/json"
	"fmt"
	"gvb_server/models/res"
	"os"

	"github.com/sirupsen/logrus"
)

const errorcodeFile = "errorcode.json"

// 读取json文件中的错误码示例
func main() {

	byteData, err := os.ReadFile(errorcodeFile)
	if err != nil {
		logrus.Error(err)
		return
	}
	var errormap map[res.ErrorCode]string
	err = json.Unmarshal(byteData, &errormap)
	if err != nil {
		logrus.Error(err)
		return
	}
	fmt.Println(errormap)
	fmt.Println(errormap[1001])
}
