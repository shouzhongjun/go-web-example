package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"goWebExample/pkg/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

// LoadMiddleware 加载所有中间件
func LoadMiddleware(logger *zap.Logger, engine *gin.Engine) {
	// 恢复中间件，用于捕获所有panic并恢复
	engine.Use(gin.Recovery())

	// 日志中间件
	engine.Use(GinLogger(logger))

	// CORS中间件
	engine.Use(corsMiddleware())

	// Gzip压缩
	engine.Use(gzip.Gzip(gzip.DefaultCompression))

	// 请求超时中间件
	engine.Use(TimeoutMiddleware(10 * time.Second))

	// 使用utils包中的全局验证器
	setupValidator(logger)

	// 404处理
	engine.NoRoute(notFoundHandler())
}

// corsMiddleware 返回CORS中间件配置
func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
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

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 包装上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 更新请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 创建完成通道
		done := make(chan struct{})

		// 处理请求
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		// 等待请求完成或超时
		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"code":    http.StatusRequestTimeout,
				"message": "请求处理超时",
			})
			return
		}
	}
}

// GinLogger 接收一个zap.Logger并返回一个gin.HandlerFunc
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start)
		logger.Info("请求日志",
			zap.Int("状态", c.Writer.Status()),
			zap.String("方法", c.Request.Method),
			zap.String("路径", path),
			zap.String("查询", query),
			zap.String("IP", c.ClientIP()),
			zap.String("用户代理", c.Request.UserAgent()),
			zap.String("错误", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("耗时", cost),
		)
	}
}
