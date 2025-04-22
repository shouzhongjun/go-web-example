package info

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/api/rest/response"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
	"goWebExample/internal/version"
)

func init() {
	// 注册模块
	module.GetRegistry().Register(module.NewBaseModule(
		"info",
		// 服务创建函数 - 这里不需要服务层，返回空
		func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
			return "", nil
		},
		// 处理器创建函数
		func(logger *zap.Logger) handlers.Handler {
			return NewInfoHandler(logger)
		},
	))
}

// InfoService 信息处理器
type InfoService struct {
	logger *zap.Logger
}

// NewInfoHandler 创建一个新的信息处理器
func NewInfoHandler(logger *zap.Logger) handlers.Handler {
	return &InfoService{
		logger: logger,
	}
}

// GetRouteGroup 获取路由组
func (h *InfoService) GetRouteGroup() handlers.RouteGroup {
	return handlers.Public
}

// GetInfo godoc
// @Summary      获取服务信息
// @Description  获取服务的基本信息，包括名称、版本和状态
// @Tags         info
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /info [get]
func (h *InfoService) GetInfo(c *gin.Context) {
	h.logger.Info("获取服务信息")
	response.SuccessWithData(c, gin.H{
		"name":      "go-server-rest-api",
		"version":   version.GetVersion(),
		"buildTime": version.GetBuildTime(),
		"commitSHA": version.CommitSHA,
		"status":    "running",
	})
}

// RegisterRoutes 注册信息相关路由
func (h *InfoService) RegisterRoutes(group *gin.RouterGroup) {
	if h == nil {
		panic("InfoService is nil when registering routes")
	}

	// 注册路由
	group.GET("/info", h.GetInfo)
}
