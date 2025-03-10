package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"goWebExample/internal/configs"
)

// RedisService Redis服务
type RedisService struct {
	client *redis.Client
	logger *zap.Logger
	config *configs.RedisConfig
}

// NewRedisService 创建Redis服务
func NewRedisService(config *configs.AllConfig, logger *zap.Logger) *RedisService {
	return &RedisService{
		logger: logger,
		config: config.Redis,
	}
}

// Name 返回服务名称
func (s *RedisService) Name() string {
	return "redis"
}

// Initialize 初始化Redis连接
func (s *RedisService) Initialize(ctx context.Context) error {
	s.client = redis.NewClient(&redis.Options{
		Addr:     s.config.Addr,
		Password: s.config.Password,
		DB:       s.config.DB,
	})

	// 测试连接
	if err := s.client.Ping(ctx).Err(); err != nil {
		return err
	}

	s.logger.Info("Redis连接成功", zap.String("addr", s.config.Addr))
	return nil
}

// Close 关闭Redis连接
func (s *RedisService) Close(ctx context.Context) error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Client 获取Redis客户端
func (s *RedisService) Client() *redis.Client {
	return s.client
}
