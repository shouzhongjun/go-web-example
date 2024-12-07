package app

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"github.com/segmentio/kafka-go"
	"goWebExample/internal/pkg/db"
	"goWebExample/internal/repository/user"
	"goWebExample/internal/service/user_service"
)

// ProviderSet 定义应用程序的所有依赖集合
var ProviderSet = wire.NewSet(
	// 数据库模块
	db.DBSet,

	// Redis 配置
	NewRedisClient, // 新增 Redis 客户端的构造函数

	// Kafka 配置
	NewKafkaWriter, // 新增 Kafka Writer 的构造函数

	// User 模块
	user.NewUserRepository,
	user_service.NewUserService,

	// 创建 App 实例
	NewApp, // 引入 App 构造函数
)

// InitializeApp 使用 Wire 自动生成依赖绑定
func InitializeApp() (*App, error) {
	wire.Build(ProviderSet)
	return &App{}, nil
}

// NewRedisClient 创建 Redis 客户端实例
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // 根据需要修改配置
	})
}

// NewKafkaWriter 创建 Kafka Writer 实例
func NewKafkaWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"), // 根据需要修改配置
		Balancer: &kafka.LeastBytes{},
	}
}
