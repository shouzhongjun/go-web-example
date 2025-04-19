package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 包装上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 更新请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 创建一个done通道，用于监听处理完成
		done := make(chan bool, 1)

		// 使用goroutine处理请求
		go func() {
			c.Next()
			done <- true
		}()

		// 等待请求完成或超时
		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 检查上下文是否已经超时
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				// 设置超时响应
				c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
					"code":    http.StatusRequestTimeout,
					"message": "请求处理超时",
				})
			}
			return
		}
	}
}
