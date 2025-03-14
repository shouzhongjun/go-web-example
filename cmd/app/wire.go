//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"goWebExample/internal/app"
	"goWebExample/internal/configs"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/infra/providers"
	"goWebExample/pkg/zap"
)

// InfraSet 提供基础设施依赖
var InfraSet = wire.NewSet(
	zap.NewZap,
	providers.ProvideServiceFactory,
	wire.FieldsOf(new(*container.ServiceContainer), "DBConnector"),
	providers.ProvideGin,
	providers.ProvideHandlerRegistry,
)

// InitializeApp 初始化应用程序
func InitializeApp(config *configs.AllConfig) (*app.App, error) {
	wire.Build(
		InfraSet,
		app.NewApp,
	)
	return &app.App{}, nil
}
