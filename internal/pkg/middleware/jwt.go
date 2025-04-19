package middleware

import (
	"goWebExample/internal/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/api/rest/response"
)

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware(jwtManager *jwt.JwtManager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("JWT认证中间件", zap.String("Method", c.Request.Method), zap.String("Path", c.Request.URL.Path))
		// 从 Header 中获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("请求未携带token")
			c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "请先登录"))
			c.Abort()
			return
		}

		// 检查 token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("token格式错误")
			c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "token格式错误"))
			c.Abort()
			return
		}

		// 解析 token
		claims, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			logger.Warn("token无效", zap.Error(err))
			c.JSON(http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, err.Error()))
			c.Abort()
			return
		}

		// 将用户信息保存到上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("isAdmin", claims.IsAdmin)

		c.Next()
	}
}
