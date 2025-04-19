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
)

// Shutdowner 定义关闭接口
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

// HTTPServer HTTP服务器
type HTTPServer struct {
	config    *configs.AllConfig
	logger    *zap.Logger
	engine    *gin.Engine
	container *container.ServiceContainer
	srv       *http.Server
	app       Shutdowner
}

// NewHTTPServer 创建新的HTTP服务器
func NewHTTPServer(
	config *configs.AllConfig,
	logger *zap.Logger,
	engine *gin.Engine,
	container *container.ServiceContainer,
) *HTTPServer {
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Server.Port),
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	server := &HTTPServer{
		config:    config,
		logger:    logger,
		engine:    engine,
		container: container,
		srv:       srv,
	}

	return server
}

// SetShutdowner 设置应用程序实例
func (s *HTTPServer) SetShutdowner(app Shutdowner) {
	s.app = app
}

// RunServer 运行服务器
func (s *HTTPServer) RunServer() error {
	// 初始化所有服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.container.Initialize(ctx); err != nil {
		s.logger.Error("初始化服务失败", zap.Error(err))
		return fmt.Errorf("初始化服务失败: %w", err)
	}

	// 初始化 Swagger
	InitSwagger(s.config, s.logger)

	// 在 goroutine 中启动服务器
	go func() {
		s.logger.Info("HTTP服务器启动",
			zap.String("地址", s.srv.Addr),
			zap.String("服务名称", s.config.Server.ServerName),
		)

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("HTTP服务器运行失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("正在关闭服务器...")

	// 关闭服务器
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		s.logger.Error("服务器强制关闭", zap.Error(err))
		return fmt.Errorf("服务器强制关闭: %w", err)
	}

	return nil
}

// Shutdown 优雅关闭服务器
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("正在关闭 HTTP 服务器...")

	// 先关闭 HTTP 服务器
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error("关闭 HTTP 服务器失败", zap.Error(err))
		return err
	}

	// 关闭所有服务
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	s.container.Shutdown(shutdownCtx)

	// 关闭应用程序（包括追踪器）
	if s.app != nil {
		appCtx, appCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer appCancel()
		if err := s.app.Shutdown(appCtx); err != nil {
			s.logger.Error("关闭应用程序失败", zap.Error(err))
		}
	}

	s.logger.Info("HTTP 服务器已关闭")
	return nil
}
