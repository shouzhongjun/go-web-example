package stream

import (
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/api/rest/response"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
	"goWebExample/internal/service"
	streamSvc "goWebExample/internal/service/stream"
)

func init() {
	// 注册模块
	module.GetRegistry().Register(module.NewBaseModule(
		"stream",
		// 服务创建函数
		func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
			streamService := streamSvc.NewStreamService(logger)
			return streamSvc.ServiceName, streamService
		},
		// 处理器创建函数
		func(logger *zap.Logger) handlers.Handler {
			return NewStreamHandler(logger)
		},
	))
}

// StreamHandler 处理流式输出相关的HTTP请求
type StreamHandler struct {
	logger *zap.Logger
}

// NewStreamHandler 创建一个新的流式处理器
func NewStreamHandler(logger *zap.Logger) *StreamHandler {
	return &StreamHandler{
		logger: logger,
	}
}

// GetRouteGroup 获取路由组
func (h *StreamHandler) GetRouteGroup() handlers.RouteGroup {
	return handlers.API
}

// HandleStream 处理流式请求
func (h *StreamHandler) HandleStream(c *gin.Context) {
	// 从服务注册器获取服务
	srv, ok := service.GetRegistry().Get(streamSvc.ServiceName).(*streamSvc.StreamService)
	if !ok || srv == nil {
		h.logger.Error("stream service not initialized")
		c.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "流式服务未初始化"))
		return
	}

	// 设置响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 示例消息
	messages := []string{
		"Hello",
		"Stream",
		"Output",
		"Example",
	}

	// 获取流式通道
	stream := srv.GenerateStream(messages)

	// 清除之前的body
	c.Writer.Flush()

	// 发送数据
	for msg := range stream {
		data := map[string]string{"message": msg}
		jsonData, err := sonic.Marshal(data)
		if err != nil {
			h.logger.Error("failed to marshal message", zap.Error(err))
			continue
		}

		// 发送SSE格式的数据
		_, err = c.Writer.Write([]byte("data: " + string(jsonData) + "\n\n"))
		if err != nil {
			return
		}
		c.Writer.Flush()
	}
}

// RegisterRoutes 注册流式输出相关路由
func (h *StreamHandler) RegisterRoutes(apiGroup *gin.RouterGroup) {
	if h == nil {
		panic("StreamHandler is nil when registering routes")
	}

	h.logger.Info("registering stream routes")
	streamGroup := apiGroup.Group("/stream")
	{
		streamGroup.POST("/msg", h.HandleStream)
	}
}
