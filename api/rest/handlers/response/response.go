package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`               // 业务状态码
	Message string      `json:"message"`            // 状态描述
	Data    interface{} `json:"data,omitempty"`     // 数据负载
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
		TraceID: c.GetString("trace_id"),
	})
}

// Error 返回错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		TraceID: c.GetString("trace_id"),
	})
}

// ValidationError 返回参数验证错误
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    400,
		Message: message,
		TraceID: c.GetString("trace_id"),
	})
}

// ServerError 返回服务器错误
func ServerError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    500,
		Message: err.Error(),
		TraceID: c.GetString("trace_id"),
	})
}

// Unauthorized 返回未授权错误
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "unauthorized"
	}
	c.JSON(http.StatusUnauthorized, Response{
		Code:    401,
		Message: message,
		TraceID: c.GetString("trace_id"),
	})
}

// Forbidden 返回禁止访问错误
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "forbidden"
	}
	c.JSON(http.StatusForbidden, Response{
		Code:    403,
		Message: message,
		TraceID: c.GetString("trace_id"),
	})
}
