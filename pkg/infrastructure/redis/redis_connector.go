package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/pkg/infrastructure/connector"
)

// RedisConnector Redis连接器实现
type RedisConnector struct {
	*connector.BaseConnector
	config *configs.Redis
	client *redis.Client
}

// NewRedisConnector 创建Redis连接器
func NewRedisConnector(config *configs.Redis, logger *zap.Logger) *RedisConnector {
	if config == nil || config.Enable == false {
		return nil
	}
	base := connector.NewBaseConnector("redis", logger)
	return &RedisConnector{
		BaseConnector: base,
		config:        config,
	}
}

// Connect 连接到Redis
func (c *RedisConnector) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return nil
	}

	c.Logger().Info("正在连接Redis",
		zap.String("addr", c.config.RedisAddr()),
		zap.Int("db", c.config.Db))

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:         c.config.RedisAddr(),
		Password:     c.config.Password,
		DB:           c.config.Db,
		PoolSize:     c.config.MaxActiveConns,
		MinIdleConns: c.config.MaxIdleConns,
		// 其他连接池配置
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接测试失败: %w", err)
	}

	c.client = client
	c.SetConnected(true)
	c.SetClient(client)
	c.Logger().Info("Redis连接成功")

	return nil
}

// Disconnect 断开Redis连接
func (c *RedisConnector) Disconnect(ctx context.Context) error {
	if !c.IsConnected() || c.client == nil {
		return nil
	}

	err := c.client.Close()
	if err != nil {
		return fmt.Errorf("关闭Redis连接失败: %w", err)
	}

	c.SetConnected(false)
	c.SetClient(nil)
	c.client = nil
	c.Logger().Info("Redis连接已关闭")

	return nil
}

// GetTypedClient 获取类型化的Redis客户端
func (c *RedisConnector) GetTypedClient() *redis.Client {
	return c.client
}

// HealthCheck 检查Redis健康状态
func (c *RedisConnector) HealthCheck(ctx context.Context) (bool, error) {
	if !c.IsConnected() || c.client == nil {
		return false, fmt.Errorf("Redis未连接")
	}

	err := c.client.Ping(ctx).Err()
	if err != nil {
		return false, fmt.Errorf("Redis健康检查失败: %w", err)
	}

	return true, nil
}
