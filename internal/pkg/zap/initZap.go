package zap

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
	"time"
)

// Logger 全局日志对象
var (
	logger *zap.Logger
	once   sync.Once
)

// NewZap 提供 Wire 依赖注入
func NewZap() *zap.Logger {
	var err error
	once.Do(func() {
		logger, err = newZapLogger("development") // 默认开发模式
		if err != nil {
			fmt.Printf("初始化日志失败: %v\n", err)
			// 创建一个基本的 logger 作为后备
			logger, _ = zap.NewProduction()
		}
	})
	return logger
}

// newZapLogger 创建 zap 日志实例
func newZapLogger(env string) (*zap.Logger, error) {
	// 验证环境参数
	if env != "development" && env != "production" {
		return nil, fmt.Errorf("invalid environment: %s", env)
	}

	encoder := getEncoder()
	level := zapcore.DebugLevel
	if env == "production" {
		level = zapcore.InfoLevel
	}

	// 确保日志目录存在
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(encoder, getLogWriter("logs/app.log"), zapcore.InfoLevel),
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)), nil
}

// getEncoder 获取日志编码器
func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "message",
		CallerKey:      "caller",
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339), // 设置时间格式
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,          // 颜色编码
		EncodeCaller:   zapcore.ShortCallerEncoder,                // 短文件路径
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// getLogWriter 获取日志文件写入器
func getLogWriter(filePath string) zapcore.WriteSyncer {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		// 如果无法打开日志文件，回退到标准错误输出
		_, err := fmt.Fprintf(os.Stderr, "无法打开日志文件: %v\n", err)
		if err != nil {
			return nil
		}
		return zapcore.AddSync(os.Stderr)
	}
	return zapcore.AddSync(file)
}
