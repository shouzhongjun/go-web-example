package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"

	"goWebExample/api/rest/response"
	"goWebExample/internal/configs"
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
		// 验证签名
		if err := apiKeySvc.VerifySign(apiKeyStr, sign, timestamp); err != nil {
			c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, err.Error()))
			c.Abort()
			return
		}
		// 验证通过，继续处理请求
		logger.Info("OpenAPI认证成功", zap.String("apiKey", apiKeyStr))
		c.Next()
	}
}
