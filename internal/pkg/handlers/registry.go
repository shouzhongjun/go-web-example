package handlers

import (
	"sync"

	"go.uber.org/zap"
)

// Registry 处理器注册器
type Registry struct {
	handlers []Handler
	logger   *zap.Logger
	mu       sync.RWMutex
}

var (
	registry *Registry
	once     sync.Once
)

// GetRegistry 获取处理器注册器单例
func GetRegistry() *Registry {
	once.Do(func() {
		registry = &Registry{
			handlers: make([]Handler, 0),
		}
	})
	return registry
}

// Init 初始化注册器
func (r *Registry) Init(logger *zap.Logger) {
	r.logger = logger
}

// Register 注册处理器
func (r *Registry) Register(handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = append(r.handlers, handler)
}

// GetHandlers 获取所有处理器
func (r *Registry) GetHandlers() []Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.handlers
}
