package httpServer

import (
	"goWebExample/api/rest"

	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	engine  *gin.Engine
	userApi *rest.UserApi
	// 其他 API 控制器
}

// NewRouter 创建路由管理器
func NewRouter(engine *gin.Engine, userApi *rest.UserApi) *Router {
	return &Router{
		engine:  engine,
		userApi: userApi,
	}
}

// Register 注册所有路由
func (r *Router) Register() {
	// API 路由组
	apiGroup := r.engine.Group("/api")

	// 注册用户相关路由
	r.registerUserRoutes(apiGroup)

	// 可以添加其他路由组的注册方法
}

// registerUserRoutes 注册用户相关路由
func (r *Router) registerUserRoutes(apiGroup *gin.RouterGroup) {
	userGroup := apiGroup.Group("/users")
	{
		userGroup.GET("/:userId", r.userApi.GetUserDetail)
		// 其他用户相关路由
	}
}
