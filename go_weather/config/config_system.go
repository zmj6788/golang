package config

import "fmt"

type System struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Env  string `yaml:"env"`
}

// 拼接服务器启动地址
func (s *System) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
