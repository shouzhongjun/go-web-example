package gorm

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// ZapLogger 适配 GORM 日志
type ZapLogger struct {
	zapLogger *zap.Logger
	logLevel  logger.LogLevel
	TraceSQL  bool // 是否追踪SQL执行
}

// LogMode 设置日志级别
func (l *ZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 级别日志
func (l *ZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		// 从 context 中获取请求 ID
		if requestID, ok := ctx.Value("request_id").(string); ok {
			l.zapLogger.Sugar().With("request_id", requestID).Infof(msg, data...)
		} else {
			l.zapLogger.Sugar().Infof(msg, data...)
		}
	}
}

// Warn 级别日志
func (l *ZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			l.zapLogger.Sugar().With("request_id", requestID).Warnf(msg, data...)
		} else {
			l.zapLogger.Sugar().Warnf(msg, data...)
		}
	}
}

// Error 级别日志
func (l *ZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			l.zapLogger.Sugar().With("request_id", requestID).Errorf(msg, data...)
		} else {
			l.zapLogger.Sugar().Errorf(msg, data...)
		}
	}
}

// Trace 记录 SQL 执行日志
func (l *ZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 只有当 TraceSQL 为 true 时才记录 SQL
	if !l.TraceSQL {
		return
	}

	// 只有当错误发生时且日志级别为 Error 或以上，或者日志级别为 Info 或以上时才记录 SQL
	if err != nil && l.logLevel >= logger.Error {
		elapsed := time.Since(begin)
		sql, rows := fc()

		if requestID, ok := ctx.Value("request_id").(string); ok {
			l.zapLogger.Sugar().With("request_id", requestID).Errorf("[SQL] %s | %s | rows: %d | err: %v", elapsed, sql, rows, err)
		} else {
			l.zapLogger.Sugar().Errorf("[SQL] %s | %s | rows: %d | err: %v", elapsed, sql, rows, err)
		}
	} else if l.logLevel >= logger.Info {
		// 只有在 Info 级别或更高时才记录正常的 SQL 执行
		elapsed := time.Since(begin)
		sql, rows := fc()

		if requestID, ok := ctx.Value("request_id").(string); ok {
			l.zapLogger.Sugar().With("request_id", requestID).Debugf("[SQL] %s | %s | rows: %d", elapsed, sql, rows)
		} else {
			l.zapLogger.Sugar().Debugf("[SQL] %s | %s | rows: %d", elapsed, sql, rows)
		}
	}
}

func NewGormZap(logger *zap.Logger, logLevel logger.LogLevel, traceSQL *bool) *ZapLogger {
	if traceSQL == nil {
		return &ZapLogger{
			zapLogger: logger,
			logLevel:  logLevel,
			TraceSQL:  false,
		}
	} else {
		return &ZapLogger{
			zapLogger: logger,
			logLevel:  logLevel,
			TraceSQL:  *traceSQL,
		}
	}
}

// GetGormLogLevel 根据项目日志级别返回对应的 Gorm 日志级别
func GetGormLogLevel(level string) logger.LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return logger.Info // Gorm 没有 Debug 级别，使用 Info 作为最详细级别
	case "info":
		return logger.Info
	case "warn", "warning":
		return logger.Warn
	case "error":
		return logger.Error
	case "fatal", "panic":
		return logger.Silent // 严重错误时，数据库操作应该是静默的
	default:
		return logger.Info // 默认使用 Info 级别
	}
}

// WithContext 为日志添加上下文信息
func (l *ZapLogger) WithContext(ctx context.Context) *ZapLogger {
	newLogger := *l
	if requestID, ok := ctx.Value("request_id").(string); ok {
		newLogger.zapLogger = l.zapLogger.With(zap.String("request_id", requestID))
	}
	return &newLogger
}
