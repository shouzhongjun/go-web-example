package middleware

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// GzipMiddleware Gzip压缩中间件
func GzipMiddleware() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression)
}
