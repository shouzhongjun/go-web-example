package service

import (
	"context"
	"fmt"
	"sync"

	"goWebExample/pkg/infrastructure/connector"
)

// Factory 服务工厂，用于管理各种连接器
type Factory struct {
	connectors sync.Map
}

// NewFactory 创建新的服务工厂
func NewFactory() *Factory {
	return &Factory{}
}

// RegisterConnector 注册连接器
func (f *Factory) RegisterConnector(name string, conn connector.Connector) {
	f.connectors.Store(name, conn)
}

// GetConnector 获取连接器
func (f *Factory) GetConnector(name string) (connector.Connector, error) {
	if conn, ok := f.connectors.Load(name); ok {
		return conn.(connector.Connector), nil
	}
	return nil, fmt.Errorf("连接器 %s 未注册", name)
}

// GetConnectorAs 获取类型化的连接器
func (f *Factory) GetConnectorAs(name string, target interface{}) error {
	conn, err := f.GetConnector(name)
	if err != nil {
		return err
	}

	if typed, ok := conn.(interface{}); ok {
		if ptr, ok := target.(*interface{}); ok {
			*ptr = typed
			return nil
		}
	}

	return fmt.Errorf("连接器 %s 类型不匹配", name)
}

// InitializeAllServices 初始化所有服务
func (f *Factory) InitializeAllServices(ctx context.Context) error {
	var errs []error
	f.connectors.Range(func(key, value interface{}) bool {
		if conn, ok := value.(connector.Connector); ok {
			if err := conn.Connect(ctx); err != nil {
				errs = append(errs, fmt.Errorf("初始化服务 %s 失败: %w", key.(string), err))
			}
		}
		return true
	})

	if len(errs) > 0 {
		return fmt.Errorf("初始化服务时发生错误: %v", errs)
	}
	return nil
}

// ShutdownAllServices 关闭所有服务
func (f *Factory) ShutdownAllServices(ctx context.Context) {
	f.connectors.Range(func(key, value interface{}) bool {
		if conn, ok := value.(connector.Connector); ok {
			if err := conn.Disconnect(ctx); err != nil {
				// 只记录错误，继续关闭其他服务
				fmt.Printf("关闭服务 %s 时发生错误: %v\n", key.(string), err)
			}
		}
		return true
	})
}
