package handlers

import (
	"github.com/gin-gonic/gin"
)

// Handler 定义处理器接口
// 所有的HTTP处理器都应实现此接口以统一路由注册方式
type Handler interface {
	// RegisterRoutes 注册路由到指定的路由组
	RegisterRoutes(group *gin.RouterGroup)
}
