package zap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"goWebExample/internal/configs"
)

// parseLevel 将字符串转换为 zapcore.Level
func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// customLevelEncoder 自定义日志级别编码器，添加颜色
func customLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var colorCode string
	switch l {
	case zapcore.DebugLevel:
		colorCode = "\x1b[36m" // Cyan
	case zapcore.InfoLevel:
		colorCode = "\x1b[34m" // Blue
	case zapcore.WarnLevel:
		colorCode = "\x1b[33m" // Yellow
	case zapcore.ErrorLevel:
		colorCode = "\x1b[31m" // Red
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		colorCode = "\x1b[35m" // Magenta
	default:
		colorCode = "\x1b[0m" // Reset
	}
	enc.AppendString(fmt.Sprintf("\t%s%s%s\t", colorCode, l.CapitalString(), "\x1b[0m"))
}

// customCallerEncoder 自定义调用者编码器
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%s\t", caller.TrimmedPath()))
}

// customTimeEncoder 自定义时间编码器
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.DateTime))
}

// getLogWriter 获取日志文件写入器
func getLogWriter(path, filename string) zapcore.WriteSyncer {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic(err)
	}

	logFile := filepath.Join(path, filename)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	return zapcore.AddSync(file)
}

// NewZap 创建一个新的 zap 日志记录器
func NewZap(config *configs.AllConfig) *zap.Logger {
	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    customLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   customCallerEncoder,
	}

	// 解析日志级别
	level := parseLevel(config.Log.Level)

	var cores []zapcore.Core

	// 如果启用了控制台输出
	if config.Log.EnableConsole {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 如果启用了文件输出
	if config.Log.EnableFile {
		today := time.Now().Format("2006-01-02")

		// 普通日志（info及以下级别）
		normalLogWriter := getLogWriter(config.Log.Path, today+".info.log")
		normalCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			normalLogWriter,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl <= zapcore.InfoLevel && lvl >= level
			}),
		)
		cores = append(cores, normalCore)

		// 错误日志（warn及以上级别）
		errorLogWriter := getLogWriter(config.Log.Path, today+".error.log")
		errorCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			errorLogWriter,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl > zapcore.InfoLevel && lvl >= level
			}),
		)
		cores = append(cores, errorCore)
	}

	// 创建日志记录器
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller())
	return logger
}
