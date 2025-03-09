package app

import (
	"goWebExample/api/rest/handlers"
	"goWebExample/internal/middleware"
	"goWebExample/internal/pkg/db"
	"goWebExample/internal/pkg/server"
	internalzap "goWebExample/internal/pkg/zap"
	"goWebExample/internal/repository/user"
	"goWebExample/internal/service/datacenter_service"
	"goWebExample/internal/service/user_service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// NewGin 创建并配置一个新的 Gin 引擎实例
func NewGin(logger *zap.Logger) *gin.Engine {
	// 根据配置文件日志级别，设置gin的模式
	if logger.Core().Enabled(zap.DebugLevel) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	// 加载中间件
	middleware.LoadMiddleware(logger, engine)
	return engine
}

// 依赖注入集合定义

var (
	// DatabaseSet 数据库相关依赖
	DatabaseSet = wire.NewSet(db.NewGormConfig)

	// LoggerSet 日志相关依赖
	LoggerSet = wire.NewSet(
		internalzap.NewZap,
	)

	// RepositorySet 仓储层依赖
	RepositorySet = wire.NewSet(
		user.NewUserRepository,
		// 其他仓储
	)

	// ServiceSet 服务层依赖
	ServiceSet = wire.NewSet(
		user_service.NewUserService,
		datacenter_service.NewMockDataCenter,
		// 其他服务
	)

	// HandlerSet Handler层依赖
	HandlerSet = wire.NewSet(
		handlers.NewUserHandler,
		handlers.NewDataCenterHandler,
		// 其他Handler
	)

	// RouterSet 路由相关依赖
	RouterSet = wire.NewSet(
		wire.Struct(new(server.Router), "*"),
	)

	// ProviderSet 汇总所有依赖
	ProviderSet = wire.NewSet(
		RepositorySet,
		ServiceSet,
		HandlerSet,
	)
)
