package server

import (
	"goWebExample/internal/configs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// InitSwagger 初始化 Swagger 路由
func InitSwagger(config *configs.AllConfig, logger *zap.Logger) {
	if !config.Swagger.Enable {
		logger.Info("Swagger 已禁用")
		return
	}

	if !config.IsDev() {
		logger.Info("非开发环境，Swagger 已禁用")
		return
	}

	logger.Info("正在初始化 Swagger")
	GlobalGroups.API.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api/swagger/doc.json")))
}
