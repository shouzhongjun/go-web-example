package connector

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Connector 定义基础连接器接口
type Connector interface {
	// Connect 连接到服务
	Connect(ctx context.Context) error
	// Disconnect 断开连接
	Disconnect(ctx context.Context) error
	// IsConnected 检查是否已连接
	IsConnected() bool
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) (bool, error)
}

// ServiceConnector 定义服务连接器接口
type ServiceConnector interface {
	// Connect 连接到服务
	Connect(ctx context.Context) error

	// Disconnect 断开服务连接
	Disconnect(ctx context.Context) error

	// IsConnected 检查是否已连接
	IsConnected() bool

	// GetClient 获取底层客户端
	GetClient() interface{}

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) (bool, error)

	// Name 获取服务名称
	Name() string
}

// BaseConnector 基础连接器实现
type BaseConnector struct {
	name      string
	logger    *zap.Logger
	connected bool
	client    interface{}
}

// NewBaseConnector 创建基础连接器
func NewBaseConnector(name string, logger *zap.Logger) *BaseConnector {
	return &BaseConnector{
		name:   name,
		logger: logger,
	}
}

// IsConnected 检查是否已连接
func (c *BaseConnector) IsConnected() bool {
	return c.connected
}

// SetConnected 设置连接状态
func (c *BaseConnector) SetConnected(connected bool) {
	c.connected = connected
}

// Logger 获取日志记录器
func (c *BaseConnector) Logger() *zap.Logger {
	return c.logger
}

// SetClient 设置客户端
func (c *BaseConnector) SetClient(client interface{}) {
	c.client = client
}

// GetClient 获取客户端
func (c *BaseConnector) GetClient() interface{} {
	return c.client
}

// Name 获取连接器名称
func (c *BaseConnector) Name() string {
	return c.name
}

// ConnectWithRetry 带重试的连接实现
func ConnectWithRetry(ctx context.Context, connector ServiceConnector, maxRetries int, initialDelay time.Duration) error {
	var err error
	delay := initialDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err = connector.Connect(ctx)
		if err == nil {
			return nil
		}

		if attempt == maxRetries {
			break
		}

		select {
		case <-time.After(delay):
			delay = delay * 2 // 指数退避
		case <-ctx.Done():
			return fmt.Errorf("连接上下文已取消: %w", ctx.Err())
		}
	}

	return fmt.Errorf("连接失败，已重试 %d 次: %w", maxRetries, err)
}

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	MaxIdleConns    int
	MaxActiveConns  int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}
