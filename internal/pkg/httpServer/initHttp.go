package httpServer

import (
	"context"
	"errors"
	"fmt"
	"goWebExample/api/rest"
	"goWebExample/internal/configs"
	"goWebExample/pkg/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HttpServer 封装HTTP服务器及其依赖
type HttpServer struct {
	AllConfig *configs.AllConfig
	Engine    *gin.Engine
	Logger    *zap.Logger
	DB        *gorm.DB
	Router    *Router
	UserApi   *rest.UserApi
}

// NewHttpServer 创建一个新的HttpServer实例
func NewHttpServer(config *configs.AllConfig, logger *zap.Logger, db *gorm.DB, engine *gin.Engine, router *Router, userApi *rest.UserApi) *HttpServer {
	server := &HttpServer{AllConfig: config, Engine: engine, Logger: logger, DB: db, Router: router, UserApi: userApi}
	// 注册路由
	server.Router.Register()

	return server
}

// RunServer 启动HTTP服务器
func (h *HttpServer) RunServer() {
	// 初始化全局验证器
	if err := utils.InitGlobalValidator(); err != nil {
		h.Logger.Error("初始化全局验证器失败", zap.Error(err))
		return
	}

	// 验证数据库连接
	if h.DB != nil {
		sqlDB, err := h.DB.DB()
		if err != nil {
			h.Logger.Error("获取数据库连接失败", zap.Error(err))
			return
		}

		if err := sqlDB.Ping(); err != nil {
			h.Logger.Error("数据库连接测试失败", zap.Error(err))
			return
		}

		h.Logger.Info("数据库连接成功")
	} else {
		h.Logger.Warn("未配置数据库连接")
	}

	h.startServer()
}

// startServer 配置并启动HTTP服务器
func (h *HttpServer) startServer() {
	// 创建HTTP服务器
	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", h.AllConfig.Server.HttpPort),
		Handler:        h.Engine,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// 在goroutine中启动服务器
	go func() {
		h.Logger.Info(fmt.Sprintf("服务器启动在 :%d", h.AllConfig.Server.HttpPort))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			h.Logger.Fatal("HTTP服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT, SIGTERM 信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	h.Logger.Info("正在关闭服务器...")

	// 创建一个5秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试优雅关闭服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		h.Logger.Error("服务器强制关闭", zap.Error(err))
	}

	h.Logger.Info(h.AllConfig.Server.ServerName + " 已退出")
}
