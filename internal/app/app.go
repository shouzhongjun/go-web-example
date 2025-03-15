package app

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "goWebExample/api/rest/handlers/datacenter"
	_ "goWebExample/api/rest/handlers/ly_stop"
	_ "goWebExample/api/rest/handlers/stream"
	_ "goWebExample/api/rest/handlers/user"
	"goWebExample/internal/configs"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/middleware"
	"goWebExample/internal/pkg/handlers"
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
func NewGin(config *configs.AllConfig, logger *zap.Logger) *gin.Engine {
	// 设置为发布模式
	gin.SetMode(gin.DebugMode)

	// 创建引擎
	engine := gin.New()

	// 加载所有中间件
	middleware.LoadMiddleware(config, logger, engine)

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
	tp, err := initTracer(config, logger)
	if err != nil {
		log.Fatalf("初始化链路追踪失败: %v", err)
	}

	// 初始化服务注册器
	service.GetRegistry().Init(logger)

	// 初始化所有模块
	module.GetRegistry().InitAll(logger, container)

	//// 添加链路追踪中间件
	//engine.Use(otelgin.Middleware(config.Trace.ServiceName))

	// 初始化全局路由组
	server.InitGroups(engine, logger, container)

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

	// 设置 Shutdowner
	httpServer.SetShutdowner(app)

	return app
}

// initTracer 初始化链路追踪
func initTracer(config *configs.AllConfig, logger *zap.Logger) (*trace.TracerProvider, error) {
	return tracer.InitTracer(&tracer.Config{
		ServiceName:    config.Trace.ServiceName,
		ServiceVersion: config.Trace.ServiceVersion,
		Environment:    config.Trace.Environment,
		Endpoint:       config.Trace.Endpoint,
		SamplingRatio:  config.Trace.SamplingRatio,
	}, logger)
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
	if err := tracer.Shutdown(ctx, app.tp, app.logger); err != nil {
		app.logger.Error("关闭链路追踪失败", zap.Error(err))
	}

	// 关闭 HTTP 服务器（包含其他服务的关闭）
	return app.httpServer.Shutdown(ctx)
}
