package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/api/rest/response"
	"goWebExample/internal/configs"
	"goWebExample/internal/repository/apikey"
	"goWebExample/internal/service"
	apikeysvc "goWebExample/internal/service/apikey"
)

// OpenAPIAuthMiddleware 创建OpenAPI认证中间件
func OpenAPIAuthMiddleware(config *configs.OpenAPIConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("OpenAPI认证中间件", zap.String("Method", c.Request.Method), zap.String("Path", c.Request.URL.Path))

		// 如果OpenAPI未启用，直接返回错误
		if !config.Enable {
			logger.Warn("OpenAPI未启用")
			c.JSON(http.StatusForbidden, response.Fail(http.StatusForbidden, "OpenAPI未启用"))
			c.Abort()
			return
		}

		// 从服务注册表获取API密钥服务
		apiKeySvc, ok := service.GetRegistry().Get(apikeysvc.ServiceName).(apikeysvc.ServiceAPIKey)
		if !ok || apiKeySvc == nil {
			logger.Error("API密钥服务未初始化")
			c.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "API密钥服务未初始化"))
			c.Abort()
			return
		}

		// 从请求中获取apikey和sign
		apiKeyStr := c.GetHeader("X-API-Key")
		sign := c.GetHeader("X-API-Sign")
		timestamp := c.GetHeader("X-API-Timestamp")

		// 验证必要参数
		if apiKeyStr == "" || sign == "" || timestamp == "" {
			logger.Warn("缺少必要的认证参数",
				zap.String("apiKey", apiKeyStr),
				zap.String("sign", sign),
				zap.String("timestamp", timestamp))
			c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "缺少必要的认证参数"))
			c.Abort()
			return
		}

		// 验证时间戳是否在有效期内（例如5分钟）
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			logger.Warn("无效的时间戳", zap.String("timestamp", timestamp), zap.Error(err))
			c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "无效的时间戳"))
			c.Abort()
			return
		}

		// 检查时间戳是否在有效期内（5分钟）
		now := time.Now().Unix()
		if now-ts > 300 || ts-now > 300 {
			logger.Warn("时间戳过期", zap.Int64("timestamp", ts), zap.Int64("now", now))
			c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "时间戳过期"))
			c.Abort()
			return
		}

		// 验证签名
		valid, err := apiKeySvc.VerifySign(apiKeyStr, sign, timestamp)
		if err != nil {
			switch err {
			case apikey.ErrAPIKeyNotFound:
				logger.Warn("无效的API Key", zap.String("apiKey", apiKeyStr))
				c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "无效的API Key"))
			case apikey.ErrAPIKeyDisabled:
				logger.Warn("API Key已禁用", zap.String("apiKey", apiKeyStr))
				c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "API Key已禁用"))
			case apikey.ErrAPIKeyExpired:
				logger.Warn("API Key已过期", zap.String("apiKey", apiKeyStr))
				c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "API Key已过期"))
			default:
				logger.Error("验证签名失败", zap.String("apiKey", apiKeyStr), zap.Error(err))
				c.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "服务器内部错误"))
			}
			c.Abort()
			return
		}

		if !valid {
			logger.Warn("签名验证失败", zap.String("apiKey", apiKeyStr))
			c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "签名验证失败"))
			c.Abort()
			return
		}

		// 验证通过，继续处理请求
		logger.Info("OpenAPI认证成功", zap.String("apiKey", apiKeyStr))
		c.Next()
	}
}
