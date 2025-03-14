package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/internal/infra/di/container"
)

var (
	// GlobalGroups 全局路由组
	GlobalGroups = &RouterGroups{}
)

// RouterGroups 预定义的路由组
type RouterGroups struct {
	API        *gin.RouterGroup // API 通用路由组
	Admin      *gin.RouterGroup // 管理后台路由组
	Public     *gin.RouterGroup // 公开路由组
	V1         *gin.RouterGroup // API v1 版本路由组
	DataCenter *gin.RouterGroup // 数据中心路由组
}

// InitGroups 初始化全局路由组
func InitGroups(engine *gin.Engine, logger *zap.Logger, container *container.ServiceContainer) {
	GlobalGroups = &RouterGroups{
		API:        engine.Group("/api"),
		Admin:      engine.Group("/admin"),
		Public:     engine.Group("/public"),
		DataCenter: engine.Group("/api/datacenter"),
		V1:         engine.Group("/api/v1"),
	}
	// 注册健康检查路由
	engine.GET("/health", func(c *gin.Context) {
		// 检查所有服务的健康状态
		healthStatus := container.Factory.HealthCheckAll(c.Request.Context())

		// 检查是否所有服务都健康
		allHealthy := true
		for _, healthy := range healthStatus {
			if !healthy {
				allHealthy = false
				break
			}
		}

		if allHealthy {
			c.JSON(200, gin.H{
				"status":   "ok",
				"services": healthStatus,
			})
		} else {
			c.JSON(503, gin.H{
				"status":   "degraded",
				"services": healthStatus,
			})
		}
	})

	logger.Info("路由组初始化完成")
}
