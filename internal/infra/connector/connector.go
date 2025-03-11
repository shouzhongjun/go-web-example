package connector

import (
	"context"

	"go.uber.org/zap"
)

// BaseConnector 定义了所有连接器的基本接口
type BaseConnector interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	SetConnected(bool)
	Logger() *zap.Logger
	HealthCheck(ctx context.Context) (bool, error)
}

// Connector 提供了基础连接器实现
type Connector struct {
	name      string
	connected bool
	logger    *zap.Logger
}

// NewConnector 创建一个新的基础连接器
func NewConnector(name string, logger *zap.Logger) *Connector {
	return &Connector{
		name:   name,
		logger: logger,
	}
}

// Connect 连接
func (c *Connector) Connect(ctx context.Context) error {
	c.connected = true
	return nil
}

// Disconnect 断开连接
func (c *Connector) Disconnect(ctx context.Context) error {
	c.connected = false
	return nil
}

// IsConnected 返回连接状态
func (c *Connector) IsConnected() bool {
	return c.connected
}

// SetConnected 设置连接状态
func (c *Connector) SetConnected(connected bool) {
	c.connected = connected
}

// Logger 返回日志器
func (c *Connector) Logger() *zap.Logger {
	return c.logger
}

// HealthCheck 健康检查
func (c *Connector) HealthCheck(ctx context.Context) (bool, error) {
	return c.connected, nil
}
