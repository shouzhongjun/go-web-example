package zap

import (
	"fmt"
	"goWebExample/internal/configs"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 全局日志对象
var (
	logger *zap.Logger
	once   sync.Once
)

// NewZap 提供 Wire 依赖注入
func NewZap(config *configs.AllConfig) *zap.Logger {
	var err error
	once.Do(func() {
		// 从配置文件获取环境模式
		env := "development"
		if config.IsDev() {
			env = "development" // 默认开发模式
		}

		logger, err = newZapLogger(env, config.Log.Path)
		if err != nil {
			fmt.Printf("初始化日志失败: %v\n", err)
			// 创建一个基本的 logger 作为后备
			logger, _ = zap.NewProduction()
		}
	})
	return logger
}

// newZapLogger 创建 zap 日志实例
func newZapLogger(env string, logPath string) (*zap.Logger, error) {
	// 验证环境参数
	if env != "development" && env != "production" {
		return nil, fmt.Errorf("invalid environment: %s", env)
	}

	// 如果日志路径为空，使用默认路径
	if logPath == "" {
		logPath = "logs"
	}

	encoder := getEncoder()
	level := zapcore.DebugLevel
	if env == "production" {
		level = zapcore.InfoLevel
	}

	// 确保日志目录存在
	if err := os.MkdirAll(logPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	logFilePath := fmt.Sprintf("%s/app.log", logPath)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(encoder, getLogWriter(logFilePath), zapcore.InfoLevel),
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
		EncodeCaller:   customCallerEncoder,                       // 自定义路径编码器
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// customCallerEncoder 自定义调用者编码器，只显示项目相对路径
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// 如果是项目内的路径，只显示相对路径部分
	path := caller.TrimmedPath()
	if idx := findProjectPath(path); idx >= 0 {
		path = path[idx:]
	}
	enc.AppendString(path)
}

// findProjectPath 查找项目路径的起始位置
func findProjectPath(path string) int {
	// 查找项目名称在路径中的位置
	projectName := "goWebExample"
	idx := -1

	// 寻找项目名称在路径中的位置
	if i := strings.Index(path, projectName); i >= 0 {
		idx = i
	}

	return idx
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
