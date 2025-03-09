package server

import (
	"context"
	"errors"
	"fmt"
	"goWebExample/internal/pkg/service"
	"goWebExample/pkg/infrastructure/etcd"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"goWebExample/internal/configs"
	"goWebExample/pkg/utils"
)

// HTTPServer 封装HTTP服务器及其依赖
type HTTPServer struct {
	AllConfig *configs.AllConfig
	Logger    *zap.Logger
	DB        *gorm.DB
	Router    *Router
	registry  etcd.ServiceRegistry
	Engine    *gin.Engine
	services  map[string]service.Service // 服务注册表
}

// NewHTTPServer 创建一个新的HttpServer实例
func NewHTTPServer(
	config *configs.AllConfig,
	logger *zap.Logger,
	db *gorm.DB,
	engine *gin.Engine,
	router *Router,
	registry etcd.ServiceRegistry,
) *HTTPServer {
	server := &HTTPServer{
		AllConfig: config,
		Logger:    logger,
		DB:        db,
		Router:    router,
		registry:  registry,
		Engine:    engine,
		services:  make(map[string]service.Service),
	}

	// 注册路由
	server.Router.Register()

	return server
}

// RegisterService 注册服务
func (s *HTTPServer) RegisterService(srv service.Service) {
	s.services[srv.Name()] = srv
	s.Logger.Info("服务已注册", zap.String("service", srv.Name()))
}

// RunServer 启动HTTP服务器
func (s *HTTPServer) RunServer() {
	// 初始化全局验证器
	if err := utils.InitGlobalValidator(); err != nil {
		s.Logger.Error("初始化全局验证器失败", zap.Error(err))
		return
	}

	// 验证数据库连接
	if s.DB != nil {
		sqlDB, err := s.DB.DB()
		if err != nil {
			s.Logger.Error("获取数据库连接失败", zap.Error(err))
			return
		}

		if err := sqlDB.Ping(); err != nil {
			s.Logger.Error("数据库连接测试失败", zap.Error(err))
			return
		}

		s.Logger.Info("数据库连接成功")
	} else {
		s.Logger.Warn("未配置数据库连接")
	}

	// 验证 etcd 服务注册
	if s.registry != nil {
		if err := s.registry.Register(context.Background()); err != nil {
			s.Logger.Error("注册服务到Etcd失败", zap.Error(err))
			return
		}
		s.Logger.Info("服务已成功注册到Etcd")
	} else {
		s.Logger.Warn("未配置Etcd服务注册")
	}

	// 初始化所有注册的服务
	if err := s.initializeServices(); err != nil {
		s.Logger.Error("初始化服务失败", zap.Error(err))
		return
	}

	s.startServer()
}

// initializeServices 初始化所有注册的服务
func (s *HTTPServer) initializeServices() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for name, srv := range s.services {
		s.Logger.Info("正在初始化服务", zap.String("service", name))
		if err := srv.Initialize(ctx); err != nil {
			s.Logger.Error("服务初始化失败", zap.String("service", name), zap.Error(err))
			return fmt.Errorf("初始化服务 %s 失败: %w", name, err)
		}
		s.Logger.Info("服务初始化成功", zap.String("service", name))
	}
	return nil
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

	// 首先尝试注册服务到Etcd
	if s.registry != nil {
		if err := s.registry.Register(context.Background()); err != nil {
			s.Logger.Error("注册服务到Etcd失败，服务器将不会启动", zap.Error(err))
			return // 直接返回，不启动HTTP服务
		}
		//s.Logger.Info("服务已成功注册到Etcd")
	} else {
		s.Logger.Info("未配置Etcd服务注册，跳过服务注册步骤")
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

	// 从Etcd注销服务
	if s.registry != nil {
		if err := s.registry.Deregister(context.Background()); err != nil {
			s.Logger.Error("从Etcd注销服务失败", zap.Error(err))
		} else {
			s.Logger.Info("服务已从Etcd成功注销")
		}
	}

	// 关闭所有服务
	s.closeServices()

	// 创建一个5秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试优雅关闭服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		s.Logger.Error("服务器强制关闭", zap.Error(err))
	}

	s.Logger.Info(s.AllConfig.Server.ServerName + " 已退出")
}

// closeServices 关闭所有服务
func (s *HTTPServer) closeServices() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for name, srv := range s.services {
		s.Logger.Info("正在关闭服务", zap.String("service", name))
		if err := srv.Close(ctx); err != nil {
			s.Logger.Error("关闭服务失败", zap.String("service", name), zap.Error(err))
		} else {
			s.Logger.Info("服务已关闭", zap.String("service", name))
		}
	}
}
