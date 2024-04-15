package res

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
}

const (
	Success = 200
	Error   = 7
)

// 返回全部响应格式信息,Result被用于封装调用
func Result(code int, data any, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

// 封装调用Result,返回特定格式信息
func Ok(data any, msg string, c *gin.Context) {
	Result(Success, data, msg, c)
}

// 第二个参数用map[string]any{}原因
// map[string]any{}可以接收任意类型的值，便于返回数据时不出错误
func OkWithData(data any, c *gin.Context) {
	Result(Success, map[string]any{}, "获取数据成功", c)
}
func OkWithMessage(msg string, c *gin.Context) {
	Result(Success, map[string]any{}, msg, c)
}

func Fail(data any, msg string, c *gin.Context) {
	Result(Error, data, msg, c)
}

func FailWithMessage(msg string, c *gin.Context) {
	Result(Error, map[string]any{}, msg, c)
}
func FailWithCode(code ErrorCode, c *gin.Context) {
	//根据错误码获取错误信息
	msg, ok := ErrorMap[code]
	if ok {
		Result(int(code), map[string]any{}, msg, c)
		return
		//return在这里作用，不让后续代码执行
	}
	Result(Error, map[string]any{}, "未知错误", c)
}
