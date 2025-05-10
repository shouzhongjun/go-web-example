package app

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "goWebExample/api/rest/handlers" // 导入所有 handlers
	"goWebExample/internal/configs"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/middleware"
	"goWebExample/internal/pkg/module"
	"goWebExample/internal/pkg/server"
	"goWebExample/internal/pkg/tracer"
	"goWebExample/internal/service"

	"go.opentelemetry.io/otel/sdk/trace"
)

// App 应用程序结构体
type App struct {
	config     *configs.AllConfig
	httpServer *server.HTTPServer
	engine     *gin.Engine
	logger     *zap.Logger
	container  *container.ServiceContainer
	tp         *trace.TracerProvider
}

// NewGin 创建 Gin 引擎
func NewGin(config *configs.AllConfig) *gin.Engine {
	switch config.Log.Level {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release", "info", "warn", "error":
		gin.SetMode(gin.ReleaseMode)
	default:
		// 默认使用 release 模式，这样更安全
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
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
	// 初始化链路追踪
	tp, err := tracer.InitTracer(config, logger)
	if err != nil {
		log.Fatalf("初始化链路追踪失败: %v", err)
	}

	// 初始化服务注册器
	service.GetRegistry().Init(logger)

	// 初始化所有模块
	module.GetRegistry().InitAll(logger, container)

	// 初始化全局路由组
	server.InitGroups(engine, logger, container)

	// 加载所有中间件，传入container以便获取Redis连接器
	middleware.LoadMiddleware(config, logger, engine, container)

	// 为OpenAPI路由组应用认证中间件
	if config.OpenAPI.Enable {
		logger.Info("为OpenAPI路由组应用认证中间件")
		server.GlobalGroups.OpenAPI.Use(middleware.OpenAPIAuthMiddleware(&config.OpenAPI, logger))
	} else {
		logger.Warn("OpenAPI未启用，跳过认证中间件")
	}

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
		case handlers.OpenAPI:
			h.RegisterRoutes(server.GlobalGroups.OpenAPI)
		default:
			h.RegisterRoutes(server.GlobalGroups.API)
		}
	}

	// 创建HTTP服务器
	httpServer := server.NewHTTPServer(
		config,
		logger,
		engine,
		container,
	)

	app := &App{
		config:     config,
		httpServer: httpServer,
		engine:     engine,
		logger:     logger,
		container:  container,
		tp:         tp,
	}

	// 设置 ShutdownHandler
	httpServer.SetShutdownHandler(app)

	return app
}

// Run 运行应用程序
func (app *App) Run() error {
	// 运行 HTTP 服务器（包含信号处理和优雅关闭）
	if err := app.httpServer.RunServer(); err != nil {
		return fmt.Errorf("运行 HTTP 服务器失败: %w", err)
	}

	return nil
}

// Shutdown 优雅关闭应用程序
func (app *App) Shutdown(ctx context.Context) error {
	// 关闭链路追踪
	cfg := tracer.DefaultShutdownConfig(app.tp, app.logger)
	if err := tracer.Shutdown(ctx, cfg); err != nil {
		app.logger.Error("关闭链路追踪失败", zap.Error(err))
		return err
	}
	return nil
}
