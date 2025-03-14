package service

import (
	"sync"

	"go.uber.org/zap"
)

// Registry 服务注册器
type Registry struct {
	services map[string]interface{}
	logger   *zap.Logger
	mu       sync.RWMutex
}

var (
	registry *Registry
	once     sync.Once
)

// GetRegistry 获取服务注册器单例
func GetRegistry() *Registry {
	once.Do(func() {
		registry = &Registry{
			services: make(map[string]interface{}),
		}
	})
	return registry
}

// Init 初始化注册器
func (r *Registry) Init(logger *zap.Logger) {
	r.logger = logger
}

// Register 注册服务
func (r *Registry) Register(name string, service interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[name] = service
}

// Get 获取服务
func (r *Registry) Get(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.services[name]
}

// GetLogger 获取日志器
func (r *Registry) GetLogger() *zap.Logger {
	return r.logger
}
