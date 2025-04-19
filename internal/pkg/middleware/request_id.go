package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware 生成请求ID的中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取请求ID
		requestID := c.Request.Header.Get("X-Request-ID")

		// 如果请求头中没有请求ID，则生成一个
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 将请求ID设置到上下文和响应头中
		c.Set("X-Request-ID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}
