package ly_stop

import (
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
	"goWebExample/internal/service"
	"goWebExample/internal/service/ly_stop"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	// 注册模块
	module.GetRegistry().Register(module.NewBaseModule(
		"ly_stop",
		// 服务创建函数
		func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
			svc := ly_stop.NewService(logger)
			return ly_stop.ServiceName, svc
		},
		// 处理器创建函数
		func(logger *zap.Logger) handlers.Handler {
			return NewHandler(logger)
		},
	))
}

// Handler 停诊处理器
type Handler struct {
	logger *zap.Logger
}

// NewHandler 创建停诊处理器
func NewHandler(logger *zap.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// GetRouteGroup 获取路由组
func (h *Handler) GetRouteGroup() handlers.RouteGroup {
	return handlers.V1
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	v1 := group.Group("/stop")
	{
		v1.GET("/list", h.GetStopList) // GET /api/v1/stop/list
	}
}

// GetStopList 获取停诊列表
// @Summary 获取停诊列表
// @Description 获取所有科室的停诊信息
// @Tags 停诊服务
// @Accept json
// @Produce json
// @Success 200 {array} ly_stop.DataMock
// @Router /v1/stop/list [get]
func (h *Handler) GetStopList(ctx *gin.Context) {
	// 从服务注册器获取服务
	svc, ok := service.GetRegistry().Get(ly_stop.ServiceName).(*ly_stop.Service)
	h.logger.Info("GetStopList", zap.Bool("ok", ok), zap.Any("svc", svc))
	if !ok || svc == nil {
		ctx.JSON(500, gin.H{"error": "stop service not initialized"})
		return
	}

	mockData := svc.GetData()
	ctx.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": mockData,
	})
}
