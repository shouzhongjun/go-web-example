package providers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/internal/app"
)

// ProvideGin 提供 Gin 引擎
func ProvideGin(logger *zap.Logger) *gin.Engine {
	return app.NewGin(logger)
}
