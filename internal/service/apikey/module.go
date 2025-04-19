package apikey

import (
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
	"goWebExample/internal/repository/apikey"

	"go.uber.org/zap"
)

func init() {
	// 注册模块
	module.GetRegistry().Register(module.NewBaseModule(
		"apikey",
		// 服务创建函数
		func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
			if container != nil && container.DBConnector != nil {
				apiKeyRepo := apikey.NewAPIKeyRepository(container.DBConnector)
				apiKeySvc := NewAPIKeyService(apiKeyRepo, logger)
				return ServiceName, apiKeySvc
			}
			logger.Error("无法初始化API密钥服务：数据库连接器未初始化")
			return "", nil
		},
		// 处理器创建函数 - 这里不需要处理器，返回nil
		func(logger *zap.Logger) handlers.Handler {
			return nil
		},
	))
}
