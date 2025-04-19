package openapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/api/rest/response"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
)

func init() {
	// 注册模块
	module.GetRegistry().Register(module.NewBaseModule(
		"openapi",
		// 服务创建函数 - 这里不需要服务层，返回空
		func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
			return "", nil
		},
		// 处理器创建函数
		func(logger *zap.Logger) handlers.Handler {
			return NewOpenAPIHandler(logger)
		},
	))
}

// OpenAPIService 开放API处理器
type OpenAPIService struct {
	logger *zap.Logger
}

// NewOpenAPIHandler 创建一个新的OpenAPI处理器
func NewOpenAPIHandler(logger *zap.Logger) handlers.Handler {
	return &OpenAPIService{
		logger: logger,
	}
}

// GetRouteGroup 获取路由组
func (h *OpenAPIService) GetRouteGroup() handlers.RouteGroup {
	return handlers.OpenAPI
}

// GetStatus godoc
// @Summary      获取API状态
// @Description  获取API服务状态信息
// @Tags         openapi
// @Accept       JSON
// @Produce      JSON
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /status [get]
func (h *OpenAPIService) GetStatus(c *gin.Context) {
	h.logger.Info("获取API状态")
	response.SuccessWithData(c, gin.H{
		"status": "ok",
		"time":   http.TimeFormat,
	})
}

// GetData godoc
// @Summary      获取数据
// @Description  获取示例数据
// @Tags         openapi
// @Accept       JSON
// @Produce      JSON
// @Param        id query string false "数据ID"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /data [get]
func (h *OpenAPIService) GetData(c *gin.Context) {
	id := c.Query("id")
	h.logger.Info("获取数据", zap.String("id", id))

	data := map[string]interface{}{
		"id":      id,
		"message": "这是一个示例数据",
		"time":    http.TimeFormat,
	}

	response.SuccessWithData(c, data)
}

// RegisterRoutes 注册OpenAPI相关路由
func (h *OpenAPIService) RegisterRoutes(group *gin.RouterGroup) {
	if h == nil {
		panic("OpenAPIService is nil when registering routes")
	}

	// 注册路由
	group.GET("/status", h.GetStatus)
	group.GET("/data", h.GetData)
}
