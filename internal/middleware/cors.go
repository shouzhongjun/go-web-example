package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CORSConfig CORS配置结构体
type CORSConfig struct {
	// 允许的源
	AllowedOrigins map[string]bool
	// 允许的HTTP方法
	AllowedMethods string
	// 允许的请求头
	AllowedHeaders []string
	// 暴露的响应头
	ExposedHeaders []string
	// 是否允许携带凭证
	AllowCredentials bool
	// 预检请求缓存时间（秒）
	MaxAge string
	// 是否启用调试模式
	Debug bool
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: map[string]bool{
			"http://localhost:3000": true,
			"http://example.com":    true,
		},
		AllowedMethods: "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS",
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Custom-Header",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Access-Control-Allow-Origin",
		},
		AllowCredentials: true,
		MaxAge:           "43200", // 12小时
		Debug:            gin.Mode() != gin.ReleaseMode,
	}
}

// Cors 跨域中间件
func Cors(logger *zap.Logger) gin.HandlerFunc {
	config := DefaultCORSConfig()

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 调试模式下记录CORS请求信息
		if config.Debug {
			logger.Debug("CORS请求",
				zap.String("origin", origin),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("headers", c.Request.Header.Get("Access-Control-Request-Headers")),
			)
		}

		// 检查是否是允许的源
		if config.AllowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else if origin != "" {
			if config.Debug {
				logger.Debug("拒绝未允许的源",
					zap.String("origin", origin),
					zap.Strings("allowed_origins", getKeys(config.AllowedOrigins)),
				)
			}
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// 设置基本的 CORS 头
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", config.AllowedMethods)

		// 设置允许的头部
		if len(config.AllowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ","))
		}

		// 设置暴露的头部
		if len(config.ExposedHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ","))
		}

		// 设置缓存时间
		c.Header("Access-Control-Max-Age", config.MaxAge)

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			if config.Debug {
				logger.Debug("预检请求响应",
					zap.String("origin", origin),
					zap.String("allowed_headers", strings.Join(config.AllowedHeaders, ",")),
					zap.String("allowed_methods", config.AllowedMethods),
				)
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getKeys 获取map的所有键
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
