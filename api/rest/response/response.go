package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义错误码常量
const (
	CodeSuccess          = 200  // 成功
	CodeBadRequest       = 400  // 请求参数错误
	CodeUnauthorized     = 401  // 未授权
	CodeForbidden        = 403  // 禁止访问
	CodeNotFound         = 404  // 资源不存在
	CodeMethodNotAllowed = 405  // 方法不允许
	CodeServerError      = 500  // 服务器内部错误
	CodeValidationError  = 1001 // 数据验证错误
	CodeDBError          = 1002 // 数据库错误
)

// Response 通用API响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id,omitempty"` // 追踪ID，用于日志追踪
}

type ResWithData struct {
	Response
	Data interface{} `json:"data"`
}

// ResponseWithData 带数据的响应结构
type ResponseWithData struct {
	Response
	Data interface{} `json:"data,omitempty"` // 数据，可选
}

// ResponseWithPagination 带分页的响应结构
type ResponseWithPagination struct {
	Response
	Data       interface{} `json:"data,omitempty"`       // 数据列表
	Pagination *Pagination `json:"pagination,omitempty"` // 分页信息
}

// Pagination 分页信息
type Pagination struct {
	Current  int   `json:"current"`   // 当前页码
	PageSize int   `json:"page_size"` // 每页条数
	Total    int64 `json:"total"`     // 总条数
}

// JSON 发送JSON响应
func JSON(c *gin.Context, resp interface{}) {
	c.JSON(http.StatusOK, resp)
}

// Success 返回成功响应
func Success(c *gin.Context) {
	JSON(c, Response{
		Code:    CodeSuccess,
		Message: "success",
	})
}

// Error 返回错误响应
func Error(c *gin.Context, code int, message string) {
	JSON(c, Response{
		Code:    code,
		Message: message,
	})
}

// Fail 返回失败响应
func Fail(code int, message string) ResWithData {
	return ResWithData{
		Response: Response{
			Code:    code,
			Message: message,
		},
		Data: nil,
	}
}

// SuccessWithMessage 返回带自定义消息的成功响应
func SuccessWithMessage(message string, data interface{}) ResWithData {
	return ResWithData{
		Response: Response{
			Code:    200,
			Message: message,
		},
		Data: data,
	}
}

func SuccessWithData(c *gin.Context, data interface{}) {
	JSON(c, ResWithData{
		Response: Response{
			Code:    CodeSuccess,
			Message: "success",
		},
		Data: data,
	})

}

// WithData 返回带数据的成功响应
func WithData(c *gin.Context, data interface{}) {
	JSON(c, ResponseWithData{
		Response: Response{
			Code:    CodeSuccess,
			Message: "success",
		},
		Data: data,
	})
}

// WithMessage 返回带自定义消息的成功响应
func WithMessage(c *gin.Context, message string) {
	JSON(c, Response{
		Code:    CodeSuccess,
		Message: message,
	})
}

// WithDataAndMessage 返回带数据和自定义消息的成功响应
func WithDataAndMessage(c *gin.Context, data interface{}, message string) {
	JSON(c, ResponseWithData{
		Response: Response{
			Code:    CodeSuccess,
			Message: message,
		},
		Data: data,
	})
}

// WithPagination 返回带分页的响应
func WithPagination(c *gin.Context, data interface{}, current, pageSize int, total int64) {
	JSON(c, ResponseWithPagination{
		Response: Response{
			Code:    CodeSuccess,
			Message: "success",
		},
		Data: data,
		Pagination: &Pagination{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// BadRequest 返回400错误响应
func BadRequest(c *gin.Context, message string) {
	Error(c, CodeBadRequest, message)
}

// Unauthorized 返回401错误响应
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权访问"
	}
	Error(c, CodeUnauthorized, message)
}

// Forbidden 返回403错误响应
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "禁止访问"
	}
	Error(c, CodeForbidden, message)
}

// NotFound 返回404错误响应
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源不存在"
	}
	Error(c, CodeNotFound, message)
}

// ServerError 返回500错误响应
func ServerError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	Error(c, CodeServerError, message)
}

// WithTrace 为响应添加追踪ID并返回
func WithTrace(c *gin.Context, resp interface{}, traceID string) {
	switch r := resp.(type) {
	case Response:
		r.TraceID = traceID
		JSON(c, r)
	case ResponseWithData:
		r.TraceID = traceID
		JSON(c, r)
	case ResponseWithPagination:
		r.TraceID = traceID
		JSON(c, r)
	default:
		JSON(c, resp)
	}
}
