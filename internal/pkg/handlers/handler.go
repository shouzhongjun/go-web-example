package handlers

import "github.com/gin-gonic/gin"

// Handler 处理器接口
type Handler interface {
	RegisterRoutes(group *gin.RouterGroup)
}
