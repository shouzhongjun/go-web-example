//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"

	"goWebExample/api/rest/handlers"
	"goWebExample/api/rest/handlers/user"
	"goWebExample/internal/app"
	"goWebExample/internal/configs"
	"goWebExample/internal/infrastructure/cache"
	"goWebExample/internal/infrastructure/db/mysql"
	"goWebExample/internal/infrastructure/di/container"
	"goWebExample/internal/infrastructure/di/factory"
	"goWebExample/internal/infrastructure/discovery"
	"goWebExample/internal/pkg/server"
	internalwire "goWebExample/internal/wire"
	zaplogger "goWebExample/pkg/zap"
)

// InfraSet 提供基础设施依赖
var InfraSet = wire.NewSet(
	// Logger
	zaplogger.NewZap,

	// ServiceFactory
	ProvideServiceFactory,
	wire.FieldsOf(new(*container.ServiceContainer), "DBConnector"),

	// Gin
	ProvideGin,

	// Handlers and Router
	ProvideHandlers,
	ProvideRouter,
)

// ServiceContainer 包含所有服务依赖
type ServiceContainer struct {
	Factory         *factory.Factory
	DBConnector     *mysql.DBConnector
	EtcdConnector   *discovery.EtcdConnector
	ServiceRegistry discovery.ServiceRegistry
}

// ProvideServiceFactory 创建服务工厂，统一管理所有连接器
func ProvideServiceFactory(config *configs.AllConfig, logger *zap.Logger) (*container.ServiceContainer, error) {
	// 创建服务工厂
	factory := factory.NewFactory(config, logger)

	container := container.NewServiceContainer(logger)
	container.Factory = factory

	// 创建数据库连接器
	dbConnector := mysql.NewDBConnector(&config.Database, logger)
	factory.RegisterConnector("db", dbConnector)
	container.DBConnector = dbConnector

	// 创建ETCD连接器
	if config.Etcd != nil && config.Etcd.Enable {
		etcdConnector := discovery.NewEtcdConnector(config.Etcd, logger)
		factory.RegisterConnector("etcd", etcdConnector)
		container.EtcdConnector = etcdConnector
		container.ServiceRegistry = discovery.NewServiceRegistry(config, logger, etcdConnector)
	}

	// 创建Redis连接器
	if config.Redis.Enable {
		redisConnector := cache.NewRedisConnector(&config.Redis, logger)
		factory.RegisterConnector("redis", redisConnector)
	}

	return container, nil
}

// ProvideGin 提供 Gin 引擎
func ProvideGin(logger *zap.Logger) *gin.Engine {
	return app.NewGin(logger)
}

// ProvideHandlers 提供处理器
func ProvideHandlers(
	logger *zap.Logger,
	userHandler *user.UserHandler,
	dataCenter handlers.Handler,
) *server.Handlers {
	return &server.Handlers{
		User:       userHandler,
		DataCenter: dataCenter,
	}
}

// ProvideRouter 提供路由管理器
func ProvideRouter(
	engine *gin.Engine,
	logger *zap.Logger,
	handlers *server.Handlers,
) *server.Router {
	return server.NewRouter(engine, logger, handlers.User, handlers.DataCenter)
}

// InitializeApp 初始化应用程序
func InitializeApp(config *configs.AllConfig) (*app.App, error) {
	wire.Build(
		InfraSet,
		internalwire.BusinessSet,
		app.NewApp,
	)
	return &app.App{}, nil
}
