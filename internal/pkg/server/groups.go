package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
func InitGroups(engine *gin.Engine, logger *zap.Logger) {
	GlobalGroups = &RouterGroups{
		API:        engine.Group("/api"),
		Admin:      engine.Group("/admin"),
		Public:     engine.Group("/public"),
		DataCenter: engine.Group("/api/datacenter"),
		V1:         engine.Group("/api/v1"),
	}

	logger.Info("路由组初始化完成")
}
