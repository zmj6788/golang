package main

import "fmt"

func main() {
	// 集合使用示例

	siteMap := make(map[string]string)

	siteMap["google"] = "www.google.com"
	siteMap["baidu"] = "www.baidu.com"
	siteMap["github"] = "www.github.com"

	for k, v := range siteMap {
		fmt.Println(k, v)
	}
	// 查看元素是否在集合中存在
	site, ok := siteMap["baidu"]
	fmt.Println(site, ok)
	site2, ok2 := siteMap["jingdong"]
	fmt.Println(site2, ok2)
	if site3, ok3 := siteMap["github"]; ok3 {
		fmt.Println(site3, ok3)
	}
}
