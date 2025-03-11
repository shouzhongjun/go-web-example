package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Router 路由管理器
type Router struct {
	engine *gin.Engine
	logger *zap.Logger
}

// NewRouter 创建路由管理器
func NewRouter(
	engine *gin.Engine,
	logger *zap.Logger,
) *Router {
	// 初始化全局路由组
	InitGroups(engine, logger)

	return &Router{
		engine: engine,
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

	r.logger.Info("路由注册完成")
}
