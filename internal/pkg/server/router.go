package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/api/rest/handlers"
	"goWebExample/api/rest/handlers/user"
)

// Router 路由管理器
type Router struct {
	engine   *gin.Engine
	handlers *Handlers
	logger   *zap.Logger
}

// Handlers 包含所有HTTP处理器
type Handlers struct {
	User       *user.UserHandler
	DataCenter handlers.Handler
}

// NewRouter 创建路由管理器
func NewRouter(
	engine *gin.Engine,
	logger *zap.Logger,
	userHandler *user.UserHandler,
	dataCenter handlers.Handler,
) *Router {
	// 初始化全局路由组
	InitGroups(engine, logger)

	return &Router{
		engine: engine,
		handlers: &Handlers{
			User:       userHandler,
			DataCenter: dataCenter,
		},
		logger: logger,
	}
}

// Register 注册所有路由
func (r *Router) Register() {
	// 注册健康检查路由
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 注册各个模块的路由
	r.registerHandlers()

	r.logger.Info("路由注册完成")
}

// registerHandlers 注册所有处理器的路由
func (r *Router) registerHandlers() {
	// 用户相关路由注册到 V1 组
	if r.handlers.User != nil {
		r.handlers.User.RegisterRoutes(GlobalGroups.V1)
	}

	// 数据中心路由注册到 DataCenter 组
	if r.handlers.DataCenter != nil {
		r.handlers.DataCenter.RegisterRoutes(GlobalGroups.DataCenter)
	}

	// 可以在这里添加更多处理器的路由注册
	// 例如:
	// if r.handlers.Auth != nil {
	//     r.handlers.Auth.RegisterRoutes(GlobalGroups.Public)
	// }
}
