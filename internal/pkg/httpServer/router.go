package httpServer

import (
	"goWebExample/api/rest/handlers"

	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	engine            *gin.Engine
	userHandler       *handlers.UserHandler
	dataCenterHandler *handlers.DataCenterHandler
	// 其他 Handler
}

// NewRouter 创建路由管理器
func NewRouter(engine *gin.Engine, userHandler *handlers.UserHandler, dataCenterHandler *handlers.DataCenterHandler) *Router {
	return &Router{
		engine:            engine,
		userHandler:       userHandler,
		dataCenterHandler: dataCenterHandler,
	}
}

// Register 注册所有路由
func (r *Router) Register() {
	// API 路由组
	apiGroup := r.engine.Group("/api")

	// 让各个Handler注册自己的路由
	r.userHandler.RegisterRoutes(apiGroup)

	// 可以添加其他Handler的路由注册
	// r.otherHandler.RegisterRoutes(apiGroup)
}

// 移除了registerUserRoutes方法，因为路由注册逻辑已经移至UserHandler
