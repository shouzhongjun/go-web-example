package container

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"goWebExample/internal/infrastructure/db/mysql"
	"goWebExample/internal/infrastructure/di/factory"
	"goWebExample/internal/infrastructure/discovery"
)

// ServiceContainer 包含所有服务依赖
type ServiceContainer struct {
	Factory         *factory.Factory
	DBConnector     *mysql.DBConnector
	EtcdConnector   *discovery.EtcdConnector
	ServiceRegistry discovery.ServiceRegistry
	logger          *zap.Logger
}

// NewServiceContainer 创建服务容器
func NewServiceContainer(logger *zap.Logger) *ServiceContainer {
	return &ServiceContainer{
		logger: logger,
	}
}

// Initialize 初始化所有服务
func (c *ServiceContainer) Initialize(ctx context.Context) error {
	// 初始化服务工厂
	if c.Factory != nil {
		if err := c.Factory.InitializeAllServices(ctx); err != nil {
			c.logger.Error("初始化服务工厂失败", zap.Error(err))
			return fmt.Errorf("初始化服务工厂失败: %w", err)
		}
		c.logger.Info("服务工厂初始化成功")
	}

	// 初始化数据库连接
	if c.DBConnector != nil {
		if err := c.DBConnector.Connect(ctx); err != nil {
			c.logger.Error("连接数据库失败", zap.Error(err))
			return fmt.Errorf("连接数据库失败: %w", err)
		}
		c.logger.Info("数据库连接成功")
	}

	// 初始化服务注册
	if c.ServiceRegistry != nil {
		if err := c.ServiceRegistry.Register(ctx); err != nil {
			c.logger.Error("注册服务失败", zap.Error(err))
			return fmt.Errorf("注册服务失败: %w", err)
		}
		c.logger.Info("服务注册成功")
	}

	return nil
}

// Shutdown 关闭所有服务
func (c *ServiceContainer) Shutdown(ctx context.Context) {
	// 先注销服务
	if c.ServiceRegistry != nil {
		if err := c.ServiceRegistry.Deregister(ctx); err != nil {
			c.logger.Error("注销服务失败", zap.Error(err))
		} else {
			c.logger.Info("服务已注销")
		}
	}

	// 关闭数据库连接
	if c.DBConnector != nil && c.DBConnector.IsConnected() {
		if err := c.DBConnector.Disconnect(ctx); err != nil {
			c.logger.Error("关闭数据库连接失败", zap.Error(err))
		} else {
			c.logger.Info("数据库连接已关闭")
		}
	}

	// 最后关闭服务工厂（包括所有连接器）
	if c.Factory != nil {
		c.Factory.ShutdownAll(ctx)
		c.logger.Info("服务工厂已关闭")
	}
}

// GetFactory 获取服务工厂
func (c *ServiceContainer) GetFactory() *factory.Factory {
	return c.Factory
}

// GetDBConnector 获取数据库连接器
func (c *ServiceContainer) GetDBConnector() *mysql.DBConnector {
	return c.DBConnector
}

// GetEtcdConnector 获取ETCD连接器
func (c *ServiceContainer) GetEtcdConnector() *discovery.EtcdConnector {
	return c.EtcdConnector
}

// GetServiceRegistry 获取服务注册器
func (c *ServiceContainer) GetServiceRegistry() discovery.ServiceRegistry {
	return c.ServiceRegistry
}
