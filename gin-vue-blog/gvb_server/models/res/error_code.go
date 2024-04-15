package res

type ErrorCode int

// 错误码
// 通常是从JSON文件中读取
const (
	SettingsError ErrorCode = 1001 //系统错误
)

// 错误码对应错误信息
var ErrorMap = map[ErrorCode]string{
	SettingsError: "系统错误",
}
