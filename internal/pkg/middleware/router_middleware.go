package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/go-playground/validator/v10"

	"goWebExample/pkg/utils"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/infra/cache"
	"goWebExample/internal/infra/di/container"
)

// GetRedisConnector 从容器中获取Redis连接器
func getRedisConnector(container *container.ServiceContainer, logger *zap.Logger) *cache.RedisConnector {
	if container == nil || container.GetFactory() == nil {
		logger.Warn("容器或工厂为空，无法获取Redis连接器")
		return nil
	}

	connector := container.GetFactory().GetConnector("redis")
	if connector == nil {
		logger.Warn("Redis连接器不存在")
		return nil
	}

	redisConn, ok := connector.(*cache.RedisConnector)
	if !ok {
		logger.Warn("Redis连接器类型不匹配")
		return nil
	}

	logger.Info("获取到Redis连接器，将用于限流中间件")
	return redisConn
}

// LoadMiddleware 加载所有中间件
func LoadMiddleware(config *configs.AllConfig, logger *zap.Logger, engine *gin.Engine, container *container.ServiceContainer) {
	// 配置 Gin 路由选项
	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true

	// 恢复中间件，用于捕获所有panic并恢复
	engine.Use(gin.Recovery())

	// 请求ID中间件
	engine.Use(RequestIDMiddleware())

	// CORS中间件 - 需要在其他中间件之前，以确保预检请求能够正确处理
	if config.Cors != nil {
		engine.Use(Cors(*config.Cors, logger))
	}

	// 添加链路追踪中间件
	engine.Use(otelgin.Middleware(config.Trace.ServiceName))

	// 请求参数日志中间件
	engine.Use(RequestParamLogger(logger, config))

	// Gzip压缩
	engine.Use(gzip.Gzip(gzip.DefaultCompression))

	// 请求超时中间件
	engine.Use(TimeoutMiddleware(10 * time.Second))

	// 限流中间件
	redisConn := getRedisConnector(container, logger)
	engine.Use(RateLimitMiddleware(config, redisConn))

	// 使用utils包中的全局验证器
	setupValidator(logger)

	// 404处理
	engine.NoRoute(notFoundHandler())
}

// notFoundHandler 处理404路由
func notFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "路由不存在",
		})
	}
}

// setupValidator 设置验证器
func setupValidator(logger *zap.Logger) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 初始化全局验证器
		if err := utils.InitGlobalValidator(); err != nil {
			logger.Error("初始化全局验证器失败", zap.Error(err))
			return
		}

		// 注册常用验证规则
		customRules := map[string]validator.Func{
			"custom_rule": func(fl validator.FieldLevel) bool {
				return len(fl.Field().String()) > 0 && fl.Field().String()[0] == 'G'
			},
			// 可以添加更多自定义规则
		}

		for tag, fn := range customRules {
			if err := v.RegisterValidation(tag, fn); err != nil {
				logger.Error("注册验证规则失败", zap.String("rule", tag), zap.Error(err))
			}
		}
	}
}
