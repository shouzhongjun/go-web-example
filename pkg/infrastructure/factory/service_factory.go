package factory

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/pkg/infrastructure/connector"
)

// ServiceFactory 管理所有外部服务连接
type ServiceFactory struct {
	config        *configs.AllConfig
	logger        *zap.Logger
	connectors    map[string]connector.ServiceConnector
	mu            sync.RWMutex
	shutdownOrder []string // 服务关闭顺序
}

// NewServiceFactory 创建服务工厂
func NewServiceFactory(config *configs.AllConfig, logger *zap.Logger) *ServiceFactory {
	return &ServiceFactory{
		config:        config,
		logger:        logger,
		connectors:    make(map[string]connector.ServiceConnector),
		shutdownOrder: []string{}, // 将按注册的反序关闭
	}
}

// RegisterConnector 注册服务连接器
func (f *ServiceFactory) RegisterConnector(name string, connector connector.ServiceConnector) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.connectors[name] = connector
	// 添加到关闭顺序列表的开头，这样后注册的会先关闭
	f.shutdownOrder = append([]string{name}, f.shutdownOrder...)
}

// GetConnector 获取服务连接器
func (f *ServiceFactory) GetConnector(name string) (connector.ServiceConnector, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if conn, ok := f.connectors[name]; ok {
		return conn, nil
	}
	return nil, fmt.Errorf("服务连接器 %s 未注册", name)
}

// InitializeServices 初始化指定的服务
func (f *ServiceFactory) InitializeServices(ctx context.Context, serviceNames ...string) error {
	for _, name := range serviceNames {
		if err := f.initializeService(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

// InitializeAllServices 初始化所有已注册的服务
func (f *ServiceFactory) InitializeAllServices(ctx context.Context) error {
	f.mu.RLock()
	services := make([]string, 0, len(f.connectors))
	for name := range f.connectors {
		services = append(services, name)
	}
	f.mu.RUnlock()

	return f.InitializeServices(ctx, services...)
}

// initializeService 初始化单个服务
func (f *ServiceFactory) initializeService(ctx context.Context, name string) error {
	connector, err := f.GetConnector(name)
	if err != nil {
		return err
	}

	if connector.IsConnected() {
		f.logger.Info("服务已连接，跳过初始化", zap.String("service", name))
		return nil
	}

	f.logger.Info("正在初始化服务连接", zap.String("service", name))
	return connector.Connect(ctx)
}

// ShutdownAll 关闭所有服务
func (f *ServiceFactory) ShutdownAll(ctx context.Context) {
	f.mu.RLock()
	shutdownOrder := make([]string, len(f.shutdownOrder))
	copy(shutdownOrder, f.shutdownOrder)
	f.mu.RUnlock()

	for _, name := range shutdownOrder {
		f.mu.RLock()
		connector, ok := f.connectors[name]
		f.mu.RUnlock()

		if !ok {
			continue
		}

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
func (f *ServiceFactory) HealthCheckAll(ctx context.Context) map[string]bool {
	f.mu.RLock()
	connectors := make(map[string]connector.ServiceConnector, len(f.connectors))
	for name, conn := range f.connectors {
		connectors[name] = conn
	}
	f.mu.RUnlock()

	results := make(map[string]bool, len(connectors))
	for name, conn := range connectors {
		healthy, _ := conn.HealthCheck(ctx)
		results[name] = healthy
	}

	return results
}
