package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "goWebExample/api/rest/handlers/datacenter"
	_ "goWebExample/api/rest/handlers/ly_stop"
	_ "goWebExample/api/rest/handlers/user"
	"goWebExample/internal/configs"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/middleware"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
	"goWebExample/internal/pkg/server"
	"goWebExample/internal/service"
)

// App 应用程序结构体
type App struct {
	httpServer *server.HTTPServer
	engine     *gin.Engine
	logger     *zap.Logger
	container  *container.ServiceContainer
}

// NewGin 创建 Gin 引擎
func NewGin(logger *zap.Logger) *gin.Engine {
	// 设置为发布模式
	gin.SetMode(gin.DebugMode)

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
	handlerRegistry *handlers.Registry,
) *App {
	app := &App{
		engine:    engine,
		logger:    logger,
		container: container,
	}

	// 初始化服务注册器
	service.GetRegistry().Init(logger)

	// 初始化所有模块
	module.GetRegistry().InitAll(logger, container)

	// 初始化全局路由组
	server.InitGroups(engine, logger)

	// 注册所有处理器的路由
	for _, h := range handlerRegistry.GetHandlers() {
		switch h.GetRouteGroup() {
		case handlers.API:
			h.RegisterRoutes(server.GlobalGroups.API)
		case handlers.Admin:
			h.RegisterRoutes(server.GlobalGroups.Admin)
		case handlers.Public:
			h.RegisterRoutes(server.GlobalGroups.Public)
		case handlers.V1:
			h.RegisterRoutes(server.GlobalGroups.V1)
		case handlers.DataCenter:
			h.RegisterRoutes(server.GlobalGroups.DataCenter)
		default:
			h.RegisterRoutes(server.GlobalGroups.API)
		}
	}

	// 创建HTTP服务器
	app.httpServer = server.NewHTTPServer(
		config,
		logger,
		engine,
		container,
	)

	return app
}

// GetHTTPServer 获取HTTP服务器实例
func (a *App) GetHTTPServer() *server.HTTPServer {
	return a.httpServer
}
