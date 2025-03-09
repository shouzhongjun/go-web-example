package logger

import (
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"time"
)

// ZapLogger 适配 GORM 日志
type ZapLogger struct {
	zapLogger *zap.Logger
	logLevel  logger.LogLevel
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
	elapsed := time.Since(begin)
	sql, rows := fc()

	if requestID, ok := ctx.Value("request_id").(string); ok {
		sugaredLogger := l.zapLogger.Sugar().With("request_id", requestID)
		if err != nil {
			sugaredLogger.Errorf("[SQL] %s | %s | rows: %d | err: %v", elapsed, sql, rows, err)
		} else {
			sugaredLogger.Debugf("[SQL] %s | %s | rows: %d", elapsed, sql, rows)
		}
	} else {
		if err != nil {
			l.zapLogger.Sugar().Errorf("[SQL] %s | %s | rows: %d | err: %v", elapsed, sql, rows, err)
		} else {
			l.zapLogger.Sugar().Debugf("[SQL] %s | %s | rows: %d", elapsed, sql, rows)
		}
	}
}

func NewGormZap(logger *zap.Logger, logLevel logger.LogLevel) *ZapLogger {
	return &ZapLogger{
		zapLogger: logger,
		logLevel:  logLevel,
	}
}
