package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
)

// Cors 跨域中间件
func Cors(config configs.Cors, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Enable {
			c.Next()
			return
		}

		origin := c.Request.Header.Get("Origin")

		// 调试模式下记录CORS请求信息
		if gin.Mode() != gin.ReleaseMode {
			logger.Debug("CORS请求",
				zap.String("origin", origin),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("headers", c.Request.Header.Get("Access-Control-Request-Headers")),
			)
		}

		// 设置CORS响应头
		if len(config.AllowedOrigins) == 0 || contains(config.AllowedOrigins, "*") {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && contains(config.AllowedOrigins, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else if origin != "" {
			if gin.Mode() != gin.ReleaseMode {
				logger.Debug("拒绝未允许的源",
					zap.String("origin", origin),
					zap.Strings("allowed_origins", config.AllowedOrigins),
				)
			}
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// 设置允许的方法
		if len(config.AllowedMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		}

		// 设置允许的请求头
		if len(config.AllowedHeaders) > 0 {
			if contains(config.AllowedHeaders, "*") {
				c.Header("Access-Control-Allow-Headers", "*")
			} else {
				c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			}
		}

		// 设置暴露的响应头
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// 设置是否允许携带凭证
		if config.AllowCredentials && !contains(config.AllowedOrigins, "*") {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求缓存时间
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", string(config.MaxAge))
		}

		// 设置是否允许私有网络访问
		if config.AllowPrivateNetwork {
			c.Header("Access-Control-Allow-Private-Network", "true")
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			if gin.Mode() != gin.ReleaseMode {
				logger.Debug("预检请求响应",
					zap.String("origin", origin),
					zap.String("allowed_headers", c.Writer.Header().Get("Access-Control-Allow-Headers")),
					zap.String("allowed_methods", c.Writer.Header().Get("Access-Control-Allow-Methods")),
				)
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// contains 检查字符串是否在切片中
func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
