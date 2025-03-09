package app

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
	"goWebExample/internal/middleware"
	"goWebExample/internal/pkg/server"
	internalzap "goWebExample/internal/pkg/zap"
	initwire "goWebExample/internal/wire"
	"goWebExample/pkg/infrastructure/db"
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

// 核心依赖注入集合

var (
	// DatabaseSet 数据库相关依赖
	DatabaseSet = wire.NewSet(db.NewDB)

	// LoggerSet 日志相关依赖
	LoggerSet = wire.NewSet(
		internalzap.NewZap,
	)

	// RouterSet 路由相关依赖
	RouterSet = wire.NewSet(
		wire.Struct(new(server.Router), "*"),
	)

	// ProviderSet 汇总所有业务模块依赖
	ProviderSet = wire.NewSet(
		initwire.AllModules,
		// 其他业务模块...
	)
)
