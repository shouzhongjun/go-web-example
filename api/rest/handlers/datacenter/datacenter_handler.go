package datacenter

import (
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
	"goWebExample/internal/service"
	"goWebExample/internal/service/datacenter"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	// 注册模块
	module.GetRegistry().Register(module.NewBaseModule(
		"datacenter",
		// 服务创建函数
		func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
			svc := datacenter.NewMockDataCenter()
			return datacenter.ServiceName, svc
		},
		// 处理器创建函数
		func(logger *zap.Logger) handlers.Handler {
			return NewDataCenterHandler(logger)
		},
	))
}

// DataCenterHandler 数据中心处理器
type DataCenterHandler struct {
	logger *zap.Logger
}

// Data 请求参数
type Data struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

// NewDataCenterHandler 创建数据中心处理器
func NewDataCenterHandler(logger *zap.Logger) *DataCenterHandler {
	return &DataCenterHandler{logger: logger}
}

// GetRouteGroup 获取路由组
func (h *DataCenterHandler) GetRouteGroup() handlers.RouteGroup {
	return handlers.DataCenter
}

// PostDataCenter 获取数据中心数据
func (h *DataCenterHandler) PostDataCenter(c *gin.Context) {
	// 从服务注册器获取服务
	srv, ok := service.GetRegistry().Get(datacenter.ServiceName).(*datacenter.MockDataCenter)
	if !ok || srv == nil {
		h.logger.Error("datacenter service not initialized")
		c.JSON(500, gin.H{"error": "datacenter service not initialized"})
		return
	}

	var data Data
	err := c.ShouldBindJSON(&data)
	if err != nil {
		h.logger.Error("failed to bind JSON", zap.Error(err))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	mockData, num, err := srv.GetMockData(data.PageNo, data.PageSize)
	if err != nil {
		h.logger.Error("failed to get mock data", zap.Error(err))
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"data": mockData,
		"num":  num,
	})
}

// RegisterRoutes registers the routes for the DataCenterHandler
func (h *DataCenterHandler) RegisterRoutes(apiGroup *gin.RouterGroup) {
	if h == nil {
		panic("DataCenterHandler is nil when registering routes")
	}

	dataCenterGroup := apiGroup.Group("/datacenter")
	{
		dataCenterGroup.POST("", h.PostDataCenter)
	}
}
