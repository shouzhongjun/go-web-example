//go:build wireinject
// +build wireinject

package main

import (
	"goWebExample/internal/app"
	"goWebExample/internal/configs"
	"goWebExample/internal/pkg/etcd"
	"goWebExample/internal/pkg/httpServer"

	"github.com/google/wire"
)

// WireApp 是 wire 的注入函数
func WireApp(config *configs.AllConfig) *httpServer.HttpServer {
	wire.Build(
		// 使用 app 包中定义的 provider 集合
		app.ProviderSet,
		app.LoggerSet,
		app.DatabaseSet,

		// 使用 app.NewGin 而不是 gin.Default
		app.NewGin,

		// 添加 Router 的 provider
		httpServer.NewRouter,

		// 添加 etcd 的 provider
		etcd.NewServiceRegistry,

		// 使用 NewHttpServer
		httpServer.NewHttpServer,
	)
	return &httpServer.HttpServer{}
}
