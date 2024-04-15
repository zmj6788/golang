package config

// 存储配置信息结构体
type Config struct {
	MySQL  MySQL  `yaml:"mysql"`
	System System `yaml:"system"`
	Logger Logger `yaml:"logger"`
}
