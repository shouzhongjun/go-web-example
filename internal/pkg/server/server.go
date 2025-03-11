package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/infra/di/container"
	"goWebExample/pkg/utils"
)

// HTTPServer 封装HTTP服务器及其依赖
type HTTPServer struct {
	AllConfig *configs.AllConfig
	Logger    *zap.Logger
	Engine    *gin.Engine
	Router    *Router
	container *container.ServiceContainer
}

// NewHTTPServer 创建一个新的HttpServer实例
func NewHTTPServer(
	config *configs.AllConfig,
	logger *zap.Logger,
	engine *gin.Engine,
	router *Router,
	container *container.ServiceContainer,
) *HTTPServer {
	server := &HTTPServer{
		AllConfig: config,
		Logger:    logger,
		Engine:    engine,
		Router:    router,
		container: container,
	}

	// 注册路由
	server.Router.Register()

	return server
}

// RunServer 启动HTTP服务器
func (s *HTTPServer) RunServer() {
	// 初始化全局验证器
	if err := utils.InitGlobalValidator(); err != nil {
		s.Logger.Error("初始化全局验证器失败", zap.Error(err))
		return
	}

	// 初始化所有服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.container.Initialize(ctx); err != nil {
		s.Logger.Error("初始化服务失败", zap.Error(err))
		return
	}

	s.startServer()
}

// startServer 配置并启动HTTP服务器
func (s *HTTPServer) startServer() {
	// 创建HTTP服务器
	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", s.AllConfig.Server.Port),
		Handler:        s.Engine,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// 在goroutine中启动服务器
	go func() {
		s.Logger.Info(fmt.Sprintf("服务器启动在 :%d", s.AllConfig.Server.Port))

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Logger.Fatal("HTTP服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.Logger.Info("正在关闭服务器...")

	// 创建一个5秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试优雅关闭服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		s.Logger.Error("服务器强制关闭", zap.Error(err))
	}

	// 关闭所有服务
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	s.container.Shutdown(shutdownCtx)

	s.Logger.Info(s.AllConfig.Server.ServerName + " 已退出")
}
