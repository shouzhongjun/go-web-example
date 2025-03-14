package providers

import (
	"go.uber.org/zap"

	"goWebExample/internal/pkg/handlers"
)

// ProvideHandlerRegistry 提供处理器注册器
func ProvideHandlerRegistry(logger *zap.Logger) *handlers.Registry {
	registry := handlers.GetRegistry()
	registry.Init(logger)
	return registry
}
