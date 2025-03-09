//go:build wireinject
// +build wireinject

package main

import (
	"goWebExample/internal/app"
	"goWebExample/internal/configs"
	"goWebExample/internal/pkg/etcd"
	"goWebExample/internal/pkg/server"

	"github.com/google/wire"
)

// InitializeApp 是 wire 的注入函数
func InitializeApp(config *configs.AllConfig) *server.HTTPServer {
	wire.Build(
		// 使用 app 包中定义的 provider 集合
		app.ProviderSet,
		app.LoggerSet,
		app.DatabaseSet,

		// 添加 etcd 的 provider
		etcd.NewServiceRegistry,
		// 使用 app.NewGin 而不是 gin.Default
		app.NewGin,

		app.RouterSet,

		// 使用 NewHTTPServer
		server.NewHTTPServer,
	)
	return &server.HTTPServer{}
}
