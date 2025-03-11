package factory

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/infra/connector"
)

// Factory 服务工厂，用于管理所有连接器
type Factory struct {
	config        *configs.AllConfig
	logger        *zap.Logger
	connectors    map[string]connector.BaseConnector
	mu            sync.RWMutex
	shutdownOrder []string // 服务关闭顺序
}

// NewFactory 创建一个新的服务工厂
func NewFactory(config *configs.AllConfig, logger *zap.Logger) *Factory {
	return &Factory{
		config:        config,
		logger:        logger,
		connectors:    make(map[string]connector.BaseConnector),
		shutdownOrder: []string{}, // 将按注册的反序关闭
	}
}

// RegisterConnector 注册一个连接器
func (f *Factory) RegisterConnector(name string, connector connector.BaseConnector) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.connectors[name] = connector
	// 添加到关闭顺序列表的开头，这样后注册的会先关闭
	f.shutdownOrder = append([]string{name}, f.shutdownOrder...)
}

// GetConnector 获取一个连接器
func (f *Factory) GetConnector(name string) connector.BaseConnector {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.connectors[name]
}

// Initialize 初始化所有连接器
func (f *Factory) Initialize(ctx context.Context) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for name, c := range f.connectors {
		if err := c.Connect(ctx); err != nil {
			f.logger.Error("初始化连接器失败",
				zap.String("name", name),
				zap.Error(err))
			return err
		}
	}
	return nil
}

// Shutdown 关闭所有连接器
func (f *Factory) Shutdown(ctx context.Context) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for name, c := range f.connectors {
		if err := c.Disconnect(ctx); err != nil {
			f.logger.Error("关闭连接器失败",
				zap.String("name", name),
				zap.Error(err))
		}
	}
}

// InitializeServices 初始化指定的服务
func (f *Factory) InitializeServices(ctx context.Context, serviceNames ...string) error {
	for _, name := range serviceNames {
		if err := f.initializeService(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

// InitializeAllServices 初始化所有已注册的服务
func (f *Factory) InitializeAllServices(ctx context.Context) error {
	f.mu.RLock()
	services := make([]string, 0, len(f.connectors))
	for name := range f.connectors {
		services = append(services, name)
	}
	f.mu.RUnlock()

	return f.InitializeServices(ctx, services...)
}

// initializeService 初始化单个服务
func (f *Factory) initializeService(ctx context.Context, name string) error {
	connector := f.GetConnector(name)

	if connector.IsConnected() {
		f.logger.Info("服务已连接，跳过初始化", zap.String("service", name))
		return nil
	}

	f.logger.Info("正在初始化服务连接", zap.String("service", name))
	return connector.Connect(ctx)
}

// ShutdownAll 关闭所有服务
func (f *Factory) ShutdownAll(ctx context.Context) {
	f.mu.RLock()
	shutdownOrder := make([]string, len(f.shutdownOrder))
	copy(shutdownOrder, f.shutdownOrder)
	f.mu.RUnlock()

	for _, name := range shutdownOrder {
		connector := f.GetConnector(name)

		if !connector.IsConnected() {
			continue
		}

		f.logger.Info("正在关闭服务", zap.String("service", name))
		if err := connector.Disconnect(ctx); err != nil {
			f.logger.Error("关闭服务失败", zap.String("service", name), zap.Error(err))
		} else {
			f.logger.Info("服务已关闭", zap.String("service", name))
		}
	}
}

// HealthCheckAll 检查所有服务的健康状态
func (f *Factory) HealthCheckAll(ctx context.Context) map[string]bool {
	results := make(map[string]bool, len(f.connectors))
	for name, conn := range f.connectors {
		healthy, _ := conn.HealthCheck(ctx)
		results[name] = healthy
	}

	return results
}
