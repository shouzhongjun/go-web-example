package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/infra/connector"
)

// RedisConnector Redis连接器
type RedisConnector struct {
	connector.Connector
	config *configs.Redis
	client *redis.Client
}

// NewRedisConnector 创建Redis连接器
func NewRedisConnector(config *configs.Redis, logger *zap.Logger) *RedisConnector {
	return &RedisConnector{
		Connector: *connector.NewConnector("redis", logger),
		config:    config,
	}
}

// Connect 连接Redis
func (c *RedisConnector) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return nil
	}

	c.Logger().Info("正在连接Redis",
		zap.String("addr", c.config.RedisAddr()),
		zap.Int("db", c.config.Db))

	client := redis.NewClient(&redis.Options{
		Addr:         c.config.RedisAddr(),
		Password:     c.config.Password,
		DB:           c.config.Db,
		PoolSize:     c.config.MaxActiveConns,
		MinIdleConns: c.config.MaxIdleConns,
		// 添加更多连接池配置
		MaxRetries:      3,                      // 最大重试次数
		MinRetryBackoff: 8 * time.Millisecond,   // 最小重试间隔
		MaxRetryBackoff: 512 * time.Millisecond, // 最大重试间隔
		DialTimeout:     5 * time.Second,        // 连接超时
		ReadTimeout:     3 * time.Second,        // 读取超时
		WriteTimeout:    3 * time.Second,        // 写入超时
		PoolTimeout:     4 * time.Second,        // 连接池超时
		IdleTimeout:     5 * time.Minute,        // 空闲连接超时
		MaxConnAge:      30 * time.Minute,       // 连接最大存活时间
	})

	// 使用带超时的上下文进行连接测试
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}

	c.client = client
	c.SetConnected(true)
	c.Logger().Info("Redis连接成功",
		zap.Int("最大连接数", c.config.MaxActiveConns),
		zap.Int("最小空闲连接", c.config.MaxIdleConns))

	return nil
}

// Disconnect 断开Redis连接
func (c *RedisConnector) Disconnect(ctx context.Context) error {
	if !c.IsConnected() {
		return nil
	}

	if err := c.client.Close(); err != nil {
		return fmt.Errorf("关闭Redis连接失败: %w", err)
	}

	c.client = nil
	c.SetConnected(false)
	c.Logger().Info("Redis连接已关闭")

	return nil
}

// GetClient 获取Redis客户端
func (c *RedisConnector) GetClient() *redis.Client {
	return c.client
}

// HealthCheck 健康检查
func (c *RedisConnector) HealthCheck(ctx context.Context) (bool, error) {
	if !c.IsConnected() || c.client == nil {
		return false, fmt.Errorf("Redis未连接")
	}

	// 使用带超时的上下文进行健康检查
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := c.client.Ping(pingCtx).Err()
	return err == nil, err
}

// FlushDB 清空当前数据库
func (c *RedisConnector) FlushDB(ctx context.Context) error {
	if !c.IsConnected() || c.client == nil {
		return fmt.Errorf("Redis未连接")
	}

	if err := c.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("清空Redis数据库失败: %w", err)
	}

	c.Logger().Info("Redis数据库已清空", zap.Int("db", c.config.Db))
	return nil
}

// Info 获取Redis服务器信息
func (c *RedisConnector) Info(ctx context.Context, section string) (string, error) {
	if !c.IsConnected() || c.client == nil {
		return "", fmt.Errorf("Redis未连接")
	}

	info, err := c.client.Info(ctx, section).Result()
	if err != nil {
		return "", fmt.Errorf("获取Redis信息失败: %w", err)
	}

	return info, nil
}

// Stats 获取Redis统计信息
func (c *RedisConnector) Stats(ctx context.Context) (*redis.PoolStats, error) {
	if !c.IsConnected() || c.client == nil {
		return nil, fmt.Errorf("Redis未连接")
	}

	stats := c.client.PoolStats()
	return stats, nil
}

// GetKeyTTL 获取键的过期时间
func (c *RedisConnector) GetKeyTTL(ctx context.Context, key string) (time.Duration, error) {
	if !c.IsConnected() || c.client == nil {
		return 0, fmt.Errorf("Redis未连接")
	}

	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("获取键[%s]的过期时间失败: %w", key, err)
	}

	return ttl, nil
}

// SetKeyTTL 设置键的过期时间
func (c *RedisConnector) SetKeyTTL(ctx context.Context, key string, expiration time.Duration) error {
	if !c.IsConnected() || c.client == nil {
		return fmt.Errorf("Redis未连接")
	}

	if err := c.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("设置键[%s]的过期时间失败: %w", key, err)
	}

	return nil
}

// DeleteKeys 删除指定的键
func (c *RedisConnector) DeleteKeys(ctx context.Context, keys ...string) error {
	if !c.IsConnected() || c.client == nil {
		return fmt.Errorf("Redis未连接")
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("删除键失败: %w", err)
	}

	c.Logger().Debug("删除Redis键成功",
		zap.Strings("keys", keys))
	return nil
}
