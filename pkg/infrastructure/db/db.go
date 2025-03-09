package db

import (
	"goWebExample/internal/configs"
	"goWebExample/pkg/logger"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NewDB 初始化并返回数据库连接
func NewDB(config *configs.AllConfig, zapLogger *zap.Logger) *gorm.DB {
	// 根据项目配置的日志级别设置 Gorm 日志级别
	gormLogLevel := logger.GetGormLogLevel(config.Database.LogLevel)
	gormZap := logger.NewGormZap(zapLogger, gormLogLevel, &config.Database.Trace)
	gormConfig := &gorm.Config{
		Logger: gormZap,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 打开数据库连接
	zapLogger.Info("正在连接数据库", zap.String("dsn", maskDSN(config.Database.DSN())))
	db, err := gorm.Open(mysql.Open(config.Database.DSN()), gormConfig)
	if err != nil {
		zapLogger.Fatal("数据库连接失败", zap.Error(err))
		return nil
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Fatal("获取底层数据库连接失败", zap.Error(err))
		return nil
	}

	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(config.Database.MaxIdleConns)

	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(config.Database.MaxOpenConns)

	// 设置连接最大生命周期
	var connMaxLifetime time.Duration
	if config.Database.ConnMaxLifetime == nil {
		connMaxLifetime = time.Hour
		zapLogger.Info("未设置连接最大生命周期，使用默认值", zap.Duration("默认值", connMaxLifetime))
	} else {
		connMaxLifetime = time.Duration(*config.Database.ConnMaxLifetime)
	}

	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		zapLogger.Fatal("数据库连接测试失败", zap.Error(err))
		return nil
	}

	zapLogger.Info("数据库连接池初始化成功",
		zap.Int("最大空闲连接数", config.Database.MaxIdleConns),
		zap.Int("最大打开连接数", config.Database.MaxOpenConns),
		zap.Duration("连接最大生命周期", connMaxLifetime),
	)

	return db
}

// maskDSN 对DSN中的敏感信息进行掩码处理
func maskDSN(dsn string) string {
	// 简单的实现，实际应用中可能需要更复杂的正则表达式
	parts := strings.Split(dsn, ":")
	if len(parts) > 1 {
		// 假设格式为 user:password@tcp(host:port)/dbname
		passwordParts := strings.Split(parts[1], "@")
		if len(passwordParts) > 0 {
			return parts[0] + ":******@" + strings.Join(passwordParts[1:], "@")
		}
	}
	return "******" // 如果无法解析，则完全掩码
}
