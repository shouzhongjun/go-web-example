package providers

import (
	"github.com/gin-gonic/gin"
	"goWebExample/internal/configs"

	"goWebExample/internal/app"
)

// ProvideGin 提供 Gin 引擎
func ProvideGin(config *configs.AllConfig) *gin.Engine {
	return app.NewGin(config)
}
