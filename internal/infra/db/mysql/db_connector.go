package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"goWebExample/internal/configs"
	"goWebExample/internal/infra/connector"
	gormLogger "goWebExample/pkg/logger/gorm"
)

// DBConnector MySQL数据库连接器
type DBConnector struct {
	connector.Connector
	config *configs.Database
	db     *gorm.DB
}

// NewDBConnector 创建MySQL连接器
func NewDBConnector(config *configs.Database, logger *zap.Logger) *DBConnector {
	return &DBConnector{
		Connector: *connector.NewConnector("mysql", logger),
		config:    config,
	}
}

// Connect 连接数据库
func (c *DBConnector) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return nil
	}

	c.Logger().Info("正在连接MySQL数据库",
		zap.String("dsn", maskDSN(c.config.DSN())))

	// 根据项目配置的日志级别设置 Gorm 日志级别
	gormLogLevel := gormLogger.GetGormLogLevel(c.config.LogLevel)
	gormZap := gormLogger.NewGormZap(c.Logger(), gormLogLevel, &c.config.Trace)
	gormConfig := &gorm.Config{
		Logger: gormZap,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,  // 启用预编译语句缓存
		SkipDefaultTransaction:                   true,  // 禁用默认事务
		AllowGlobalUpdate:                        false, // 禁止全局更新
		QueryFields:                              true,  // 启用全字段查询
	}

	db, err := gorm.Open(mysql.Open(c.config.DSN()), gormConfig)
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %w", err)
	}
	// 添加 OpenTelemetry 插件
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取SQL.DB失败: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxOpenConns(c.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.config.MaxIdleConns)

	// 设置连接最大生命周期
	var connMaxLifetime time.Duration
	if c.config.ConnMaxLifetime == nil {
		connMaxLifetime = time.Hour
		c.Logger().Info("未设置连接最大生命周期，使用默认值", zap.Duration("默认值", connMaxLifetime))
	} else {
		connMaxLifetime = time.Duration(*c.config.ConnMaxLifetime) * time.Second
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// 设置连接最大空闲时间
	var connMaxIdleTime time.Duration
	if c.config.ConnMaxIdleTime == nil {
		connMaxIdleTime = time.Hour
		c.Logger().Info("未设置连接最大空闲时间，使用默认值", zap.Duration("默认值", connMaxIdleTime))
	} else {
		connMaxIdleTime = time.Duration(*c.config.ConnMaxIdleTime) * time.Second
	}
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	c.db = db
	c.SetConnected(true)
	c.Logger().Info("MySQL数据库连接成功",
		zap.Int("最大空闲连接数", c.config.MaxIdleConns),
		zap.Int("最大打开连接数", c.config.MaxOpenConns),
		zap.Duration("连接最大生命周期", connMaxLifetime),
		zap.Duration("连接最大空闲时间", connMaxIdleTime))

	return nil
}

// Disconnect 断开数据库连接
func (c *DBConnector) Disconnect(ctx context.Context) error {
	if !c.IsConnected() {
		return nil
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("获取SQL.DB失败: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭MySQL连接失败: %w", err)
	}

	c.db = nil
	c.SetConnected(false)
	c.Logger().Info("MySQL数据库连接已关闭")

	return nil
}

// GetDB 获取数据库连接
func (c *DBConnector) GetDB() *gorm.DB {
	return c.db
}

// GetDBWithContext 获取带有上下文的数据库连接
func (c *DBConnector) GetDBWithContext(ctx context.Context) *gorm.DB {
	return c.db.WithContext(ctx)
}

// HealthCheck 健康检查
func (c *DBConnector) HealthCheck(ctx context.Context) (bool, error) {
	if !c.IsConnected() || c.db == nil {
		return false, fmt.Errorf("MySQL未连接")
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return false, fmt.Errorf("获取SQL.DB失败: %w", err)
	}

	// 使用带超时的上下文进行健康检查
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err = sqlDB.PingContext(pingCtx)
	return err == nil, err
}

// Stats 获取数据库连接池统计信息
func (c *DBConnector) Stats(ctx context.Context) (*sql.DBStats, error) {
	if !c.IsConnected() || c.db == nil {
		return nil, fmt.Errorf("MySQL未连接")
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取SQL.DB失败: %w", err)
	}

	stats := sqlDB.Stats()
	return &stats, nil
}

// Transaction 执行事务
func (c *DBConnector) Transaction(ctx context.Context, fn func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	if !c.IsConnected() || c.db == nil {
		return fmt.Errorf("MySQL未连接")
	}

	return c.db.WithContext(ctx).Transaction(fn, opts...)
}

// maskDSN 对DSN中的敏感信息进行掩码处理
func maskDSN(dsn string) string {
	parts := strings.Split(dsn, ":")
	if len(parts) > 1 {
		passwordParts := strings.Split(parts[1], "@")
		if len(passwordParts) > 0 {
			return parts[0] + ":******@" + strings.Join(passwordParts[1:], "@")
		}
	}
	return "******"
}
