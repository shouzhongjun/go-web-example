package module

import (
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/service"

	"go.uber.org/zap"
)

// Module 模块接口
type Module interface {
	// Name 模块名称
	Name() string
	// Init 初始化模块
	Init(logger *zap.Logger, container *container.ServiceContainer)
}

// BaseModule 基础模块实现
type BaseModule struct {
	name           string
	serviceCreator func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{})
	handlerCreator func(logger *zap.Logger) handlers.Handler
}

// NewBaseModule 创建基础模块
func NewBaseModule(
	name string,
	serviceCreator func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}),
	handlerCreator func(logger *zap.Logger) handlers.Handler,
) *BaseModule {
	return &BaseModule{
		name:           name,
		serviceCreator: serviceCreator,
		handlerCreator: handlerCreator,
	}
}

// Name 获取模块名称
func (m *BaseModule) Name() string {
	return m.name
}

// Init 初始化模块
func (m *BaseModule) Init(logger *zap.Logger, container *container.ServiceContainer) {
	// 初始化并注册服务
	if m.serviceCreator != nil {
		serviceName, svc := m.serviceCreator(logger, container)
		if svc != nil {
			service.GetRegistry().Register(serviceName, svc)
		}
	}

	// 初始化并注册处理器
	if m.handlerCreator != nil {
		handler := m.handlerCreator(logger)
		if handler != nil {
			handlers.GetRegistry().Register(handler)
		}
	}
}

// Registry 模块注册器
type Registry struct {
	modules []Module
}

var (
	registry *Registry
)

// GetRegistry 获取模块注册器
func GetRegistry() *Registry {
	if registry == nil {
		registry = &Registry{
			modules: make([]Module, 0),
		}
	}
	return registry
}

// Register 注册模块
func (r *Registry) Register(module Module) {
	r.modules = append(r.modules, module)
}

// InitAll 初始化所有模块
func (r *Registry) InitAll(logger *zap.Logger, container *container.ServiceContainer) {
	for _, module := range r.modules {
		logger.Info("初始化模块", zap.String("module", module.Name()))
		module.Init(logger, container)
	}
}
