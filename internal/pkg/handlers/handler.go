package handlers

import "github.com/gin-gonic/gin"

// RouteGroup 路由组类型
type RouteGroup string

const (
	API        RouteGroup = "api"        // API 通用路由组
	Admin      RouteGroup = "admin"      // 管理后台路由组
	Public     RouteGroup = "public"     // 公开路由组
	V1         RouteGroup = "v1"         // API v1 版本路由组
	DataCenter RouteGroup = "datacenter" // 数据中心路由组
	OpenAPI    RouteGroup = "openapi"
)

// Handler 处理器接口
type Handler interface {
	// RegisterRoutes 注册路由到指定的路由组
	RegisterRoutes(group *gin.RouterGroup)
	// GetRouteGroup 获取处理器使用的路由组
	GetRouteGroup() RouteGroup
}
