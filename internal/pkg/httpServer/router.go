package httpServer

import (
	"goWebExample/api/rest/handlers"

	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	Engine            *gin.Engine // 首字母大写以便 Wire 可以注入
	UserHandler       *handlers.UserHandler
	DataCenterHandler *handlers.DataCenterHandler
	// 其他 Handler
}

// Register 注册所有路由
func (r *Router) Register() {
	// API 路由组
	apiGroup := r.Engine.Group("/api")

	// 让各个Handler注册自己的路由
	if r.UserHandler != nil {
		r.UserHandler.RegisterRoutes(apiGroup)
	}

	// 数据中心路由注册
	if r.DataCenterHandler != nil {
		r.DataCenterHandler.RegisterRoutes(apiGroup)
	}

	// 可以添加其他Handler的路由注册
	// r.otherHandler.RegisterRoutes(apiGroup)
}
