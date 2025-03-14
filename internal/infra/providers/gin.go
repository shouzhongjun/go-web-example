package providers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goWebExample/internal/configs"

	"goWebExample/internal/app"
)

// ProvideGin 提供 Gin 引擎
func ProvideGin(config *configs.AllConfig, logger *zap.Logger) *gin.Engine {
	return app.NewGin(config, logger)
}
