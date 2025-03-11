package providers

import (
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/infra/cache"
	"goWebExample/internal/infra/db/mysql"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/infra/di/factory"
	"goWebExample/internal/infra/discovery"
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
