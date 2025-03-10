package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"goWebExample/internal/configs"
	"goWebExample/pkg/infrastructure/connector"
	"goWebExample/pkg/logger"
)

// DBConnector 数据库连接器实现
type DBConnector struct {
	*connector.BaseConnector
	config *configs.Database
	db     *gorm.DB
}

// NewDBConnector 创建数据库连接器
func NewDBConnector(config *configs.Database, logger *zap.Logger) *DBConnector {
	base := connector.NewBaseConnector("database", logger)
	return &DBConnector{
		BaseConnector: base,
		config:        config,
	}
}

// Connect 连接到数据库
func (c *DBConnector) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return nil
	}

	c.Logger().Info("正在连接数据库", zap.String("dsn", maskDSNString(c.config.DSN())))

	// 根据项目配置的日志级别设置 Gorm 日志级别
	gormLogLevel := logger.GetGormLogLevel(c.config.LogLevel)
	gormZap := logger.NewGormZap(c.Logger(), gormLogLevel, &c.config.Trace)
	gormConfig := &gorm.Config{
		Logger: gormZap,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(c.config.DSN()), gormConfig)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(c.config.MaxIdleConns)

	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(c.config.MaxOpenConns)

	// 设置连接最大生命周期
	var connMaxLifetime time.Duration
	if c.config.ConnMaxLifetime == nil {
		connMaxLifetime = time.Hour
		c.Logger().Info("未设置连接最大生命周期，使用默认值", zap.Duration("默认值", connMaxLifetime))
	} else {
		connMaxLifetime = time.Duration(*c.config.ConnMaxLifetime) * time.Minute
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// 设置连接最大空闲时间
	var connMaxIdleTime time.Duration
	if c.config.ConnMaxIdleTime == nil {
		connMaxIdleTime = time.Hour
		c.Logger().Info("未设置连接最大空闲时间，使用默认值", zap.Duration("默认值", connMaxIdleTime))
	} else {
		connMaxIdleTime = time.Duration(*c.config.ConnMaxIdleTime) * time.Minute
	}
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	c.db = db
	c.SetConnected(true)
	c.SetClient(db)
	c.Logger().Info("数据库连接池初始化成功",
		zap.Int("最大空闲连接数", c.config.MaxIdleConns),
		zap.Int("最大打开连接数", c.config.MaxOpenConns),
		zap.Duration("连接最大生命周期", connMaxLifetime),
		zap.Duration("连接最大空闲时间", connMaxIdleTime),
	)

	return nil
}

// Disconnect 断开数据库连接
func (c *DBConnector) Disconnect(ctx context.Context) error {
	if !c.IsConnected() || c.db == nil {
		return nil
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}

	c.SetConnected(false)
	c.SetClient(nil)
	c.db = nil
	c.Logger().Info("数据库连接已关闭")

	return nil
}

// GetTypedClient 获取类型化的数据库客户端
func (c *DBConnector) GetTypedClient() *gorm.DB {
	return c.db
}

// HealthCheck 检查数据库健康状态
func (c *DBConnector) HealthCheck(ctx context.Context) (bool, error) {
	if !c.IsConnected() || c.db == nil {
		return false, fmt.Errorf("数据库未连接")
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return false, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return false, fmt.Errorf("数据库健康检查失败: %w", err)
	}

	return true, nil
}

// maskDSNString 对DSN中的敏感信息进行掩码处理
func maskDSNString(dsn string) string {
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
