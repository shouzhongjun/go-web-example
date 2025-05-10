package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/infra/di/container"
)

// ShutdownHandler 定义关闭接口
type ShutdownHandler interface {
	Shutdown(ctx context.Context) error
}

// HTTPTimeouts 服务器超时配置
type HTTPTimeouts struct {
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	InitTimeout        time.Duration
	ShutdownTimeout    time.Duration
	ContainerTimeout   time.Duration
	ApplicationTimeout time.Duration
}

// DefaultTimeouts 返回默认的超时配置
func DefaultTimeouts() HTTPTimeouts {
	return HTTPTimeouts{
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		InitTimeout:        30 * time.Second,
		ShutdownTimeout:    10 * time.Second,
		ContainerTimeout:   10 * time.Second,
		ApplicationTimeout: 5 * time.Second,
	}
}

// HTTPServer HTTP服务器
type HTTPServer struct {
	config    *configs.AllConfig
	logger    *zap.Logger
	engine    *gin.Engine
	container *container.ServiceContainer
	srv       *http.Server
	app       ShutdownHandler
	timeouts  HTTPTimeouts
}

// NewHTTPServer 创建新的HTTP服务器
func NewHTTPServer(
	config *configs.AllConfig,
	logger *zap.Logger,
	engine *gin.Engine,
	container *container.ServiceContainer,
) *HTTPServer {
	timeouts := DefaultTimeouts()

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Server.Port),
		Handler:        engine,
		ReadTimeout:    timeouts.ReadTimeout,
		WriteTimeout:   timeouts.WriteTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	server := &HTTPServer{
		config:    config,
		logger:    logger,
		engine:    engine,
		container: container,
		srv:       srv,
		timeouts:  timeouts,
	}

	return server
}

// SetShutdownHandler 设置应用程序实例
func (s *HTTPServer) SetShutdownHandler(app ShutdownHandler) {
	s.app = app
}

// SetTimeouts 自定义服务器超时设置
func (s *HTTPServer) SetTimeouts(timeouts HTTPTimeouts) {
	s.timeouts = timeouts
	s.srv.ReadTimeout = timeouts.ReadTimeout
	s.srv.WriteTimeout = timeouts.WriteTimeout
}

// RunServer 运行服务器
func (s *HTTPServer) RunServer() error {
	// 初始化所有服务
	ctx, cancel := context.WithTimeout(context.Background(), s.timeouts.InitTimeout)
	defer cancel()

	if err := s.container.Initialize(ctx); err != nil {
		s.logger.Error("初始化服务失败", zap.Error(err))
		return fmt.Errorf("初始化服务失败: %w", err)
	}

	// 检查InitSwagger函数是否可用
	swaggerInitialized := false
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("初始化Swagger失败", zap.Any("error", r))
		}
	}()

	// 初始化Swagger
	InitSwagger(s.config, s.logger)
	swaggerInitialized = true

	// 恢复正常流程
	if !swaggerInitialized {
		s.logger.Warn("Swagger未初始化，继续启动服务")
	}

	// 创建错误通道和完成通道，用于监听HTTP服务器状态
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})
	defer close(doneChan) // 确保在函数结束时关闭通道

	// 在goroutine中启动服务器
	go func() {
		s.logger.Info("HTTP服务器启动",
			zap.String("地址", s.srv.Addr),
			zap.String("服务名称", s.config.Server.ServerName),
		)

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("HTTP服务器运行失败", zap.Error(err))
			select {
			case errChan <- err:
				// 错误已发送
			case <-doneChan:
				// RunServer已退出，不需要发送错误
			}
		}
	}()

	// 实现简单的健康检查
	go func() {
		// 给服务器一点时间启动
		time.Sleep(500 * time.Millisecond)

		healthCheckURL := fmt.Sprintf("http://localhost:%d/health", s.config.Server.Port)
		client := http.Client{Timeout: 2 * time.Second}

		// 尝试5次健康检查
		for i := 0; i < 5; i++ {
			resp, err := client.Get(healthCheckURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				s.logger.Info("服务器健康检查通过")
				err := resp.Body.Close()
				if err != nil {
					s.logger.Error("关闭响应体失败", zap.Error(err))
					return
				}
				break
			}

			if i == 4 {
				s.logger.Warn("服务器健康检查失败，但继续运行",
					zap.Int("attempts", i+1))
			}

			if resp != nil {
				err := resp.Body.Close()
				if err != nil {
					s.logger.Error("关闭响应体失败", zap.Error(err))
					return
				}
			}

			time.Sleep(time.Second)
		}
	}()

	// 等待中断信号或HTTP服务器错误
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待中断信号或HTTP服务器错误
	select {
	case <-quit:
		s.logger.Info("收到中断信号，正在关闭服务器...")
	case err := <-errChan:
		s.logger.Error("HTTP服务器启动失败，正在关闭服务...", zap.Error(err))
		return fmt.Errorf("HTTP服务器启动失败: %w", err)
	}

	s.logger.Info("正在关闭服务器...")

	// 关闭服务器
	ctx, cancel = context.WithTimeout(context.Background(), s.timeouts.ShutdownTimeout)
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
	var wg sync.WaitGroup
	var shutdownErr error
	var mu sync.Mutex

	// 设置错误处理函数
	setError := func(err error, message string) {
		if err != nil {
			mu.Lock()
			if shutdownErr == nil {
				shutdownErr = fmt.Errorf("%s: %w", message, err)
			}
			mu.Unlock()
			s.logger.Error(message, zap.Error(err))
		}
	}

	// 先关闭 HTTP 服务器
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.srv.Shutdown(ctx); err != nil {
			setError(err, "关闭 HTTP 服务器失败")
		} else {
			s.logger.Info("HTTP 服务器已关闭")
		}
	}()

	// 等待HTTP服务器关闭完成
	wgDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		// HTTP服务器已成功关闭
	case <-ctx.Done():
		return fmt.Errorf("HTTP服务器关闭超时: %w", ctx.Err())
	}

	// 关闭应用程序（如果存在）
	if s.app != nil {
		appCtx, appCancel := context.WithTimeout(ctx, s.timeouts.ApplicationTimeout)
		defer appCancel()

		if err := s.app.Shutdown(appCtx); err != nil {
			setError(err, "关闭应用程序失败")
		} else {
			s.logger.Info("应用程序已关闭")
		}
	}

	// 最后关闭所有服务
	containerCtx, containerCancel := context.WithTimeout(ctx, s.timeouts.ContainerTimeout)
	defer containerCancel()

	s.container.Shutdown(containerCtx)
	s.logger.Info("所有服务已关闭")

	return shutdownErr
}
