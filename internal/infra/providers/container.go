package providers

import (
	"go.uber.org/zap"
	"goWebExample/internal/infra/mq"
	"goWebExample/internal/pkg/jwt"

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
	JWTManager      *jwt.JwtManager
}

// ProvideServiceFactory 创建服务工厂，统一管理所有连接器
func ProvideServiceFactory(config *configs.AllConfig, logger *zap.Logger) (*container.ServiceContainer, error) {
	// 创建服务工厂
	newFactory := factory.NewFactory(config, logger)

	serviceContainer := container.NewServiceContainer(logger)
	serviceContainer.Factory = newFactory

	// 创建数据库连接器
	dbConnector := mysql.NewDBConnector(&config.Database, logger)
	newFactory.RegisterConnector("db", dbConnector)
	serviceContainer.DBConnector = dbConnector

	// 创建 JWT 管理器
	jwtManager := jwt.NewJWTManager(jwt.Config{
		SecretKey: config.JWT.SecretKey,
		Issuer:    config.JWT.Issuer,
		Duration:  config.JWT.Duration,
	})
	serviceContainer.JWTManager = jwtManager

	// 创建ETCD连接器
	if config.Etcd != nil && config.Etcd.Enable {
		etcdConnector := discovery.NewEtcdConnector(config.Etcd, logger)
		newFactory.RegisterConnector("etcd", etcdConnector)
		serviceContainer.EtcdConnector = etcdConnector
		serviceContainer.ServiceRegistry = discovery.NewServiceRegistry(config, logger, etcdConnector)
	}

	// 创建Redis连接器
	if config.Redis.Enable {
		redisConnector := cache.NewRedisConnector(&config.Redis, logger)
		newFactory.RegisterConnector("redis", redisConnector)
	}

	if config.Kafka.Enable || config.Kafka.Host != "" {
		kafkaConnector := mq.NewKafkaConnector(&config.Kafka, logger)
		newFactory.RegisterConnector("kafka", kafkaConnector)
	}

	return serviceContainer, nil
}
