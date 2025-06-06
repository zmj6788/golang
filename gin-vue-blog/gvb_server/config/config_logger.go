package config

type Logger struct {
	Level        string `yaml:"level"`
	Prefix       string `yaml:"prefix"` // 日志前缀
	Director     string `yaml:"director"`
	ShowLine     bool   `yaml:"show_line"`      //是否显示行号
	LogInConsole bool   `yaml:"log_in_console"` // 是否显示打印的路径
}
