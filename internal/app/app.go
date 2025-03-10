package app

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/middleware"
	"goWebExample/internal/pkg/server"
	"goWebExample/pkg/infrastructure/container"
)

// App 应用程序结构体
type App struct {
	httpServer *server.HTTPServer
}

// NewGin 创建 Gin 引擎
func NewGin(logger *zap.Logger) *gin.Engine {
	// 设置为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建引擎
	engine := gin.New()

	// 加载所有中间件
	middleware.LoadMiddleware(logger, engine)

	return engine
}

// NewApp 创建应用程序实例
func NewApp(
	config *configs.AllConfig,
	logger *zap.Logger,
	engine *gin.Engine,
	container *container.ServiceContainer,
	handlers *server.Handlers,
) *App {
	// 创建路由
	router := server.NewRouter(
		engine,
		logger,
		handlers.User,
		handlers.DataCenter,
	)

	// 创建HTTP服务器
	httpServer := server.NewHTTPServer(
		config,
		logger,
		engine,
		router,
		container,
	)

	return &App{
		httpServer: httpServer,
	}
}

// GetHTTPServer 获取HTTP服务器实例
func (a *App) GetHTTPServer() *server.HTTPServer {
	return a.httpServer
}

// Run 运行应用程序
func (a *App) Run(ctx context.Context) error {
	// 运行HTTP服务器
	a.httpServer.RunServer()
	return nil
}
