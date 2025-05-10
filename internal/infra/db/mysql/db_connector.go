package mysql

import (
	"context"
	"database/sql"
	"errors"
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

// 定义常见错误
var (
	ErrNotConnected = errors.New("MySQL未连接")
	ErrGetSQLDB     = errors.New("获取SQL.DB失败")
)

// DBConnector MySQL数据库连接器
type DBConnector struct {
	connector.Connector
	config *configs.Database
	db     *gorm.DB
}

// checkConnection 检查数据库连接状态并返回SQL DB
func (c *DBConnector) checkConnection() (*sql.DB, error) {
	if !c.IsConnected() || c.db == nil {
		return nil, ErrNotConnected
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetSQLDB, err)
	}
	return sqlDB, nil
}

// setDuration 设置时间相关配置，返回配置的时间值
func (c *DBConnector) setDuration(value *int64, defaultDuration time.Duration, name string) time.Duration {
	var duration time.Duration
	if value == nil {
		duration = defaultDuration
		c.Logger().Info("未设置"+name+"，使用默认值", zap.Duration("默认值", duration))
	} else {
		duration = time.Duration(*value) * time.Second
	}
	return duration
}

// NewDBConnector 创建MySQL连接器
func NewDBConnector(config *configs.Database, logger *zap.Logger) *DBConnector {
	// 检查 logger 是否为 nil
	if logger == nil {
		return nil
	}

	// 检查 config 是否为 nil
	if config == nil {
		// 可以选择返回 nil 或者使用默认配置
		return &DBConnector{
			Connector: *connector.NewConnector("mysql", logger),
			config:    nil, // 或者使用默认配置
		}
	}

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
	// 检查 config 是否为 nil
	if c.config == nil {
		return fmt.Errorf("配置为空，无法连接数据库")
	}

	c.Logger().Info("正在连接MySQL数据库",
		zap.String("dsn", maskDSN(c.config.DSN(), c.config.LogLevel)))

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

	// 设置连接最大生命周期和最大空闲时间
	connMaxLifetime := c.setDuration(c.config.ConnMaxLifetime, time.Hour, "连接最大生命周期")
	connMaxIdleTime := c.setDuration(c.config.ConnMaxIdleTime, time.Hour, "连接最大空闲时间")

	sqlDB.SetConnMaxLifetime(connMaxLifetime)
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

	sqlDB, err := c.checkConnection()
	if err != nil {
		return err
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
	sqlDB, err := c.checkConnection()
	if err != nil {
		return false, err
	}

	// 使用带超时的上下文进行健康检查
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err = sqlDB.PingContext(pingCtx)
	return err == nil, err
}

// Stats 获取数据库连接池统计信息
func (c *DBConnector) Stats(ctx context.Context) (*sql.DBStats, error) {
	sqlDB, err := c.checkConnection()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return &stats, nil
}

// Transaction 执行事务
func (c *DBConnector) Transaction(ctx context.Context, fn func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	_, err := c.checkConnection()
	if err != nil {
		return err
	}

	return c.db.WithContext(ctx).Transaction(fn, opts...)
}

// maskDSN 对DSN中的敏感信息进行掩码处理
func maskDSN(dsn string, logLevel string) string {
	if logLevel == "debug" {
		return dsn
	}
	if dsn == "" {
		return ""
	}

	// 尝试解析标准格式的DSN: username:password@protocol(address)/dbname
	userPwdSplit := strings.SplitN(dsn, ":", 2)
	if len(userPwdSplit) < 2 {
		// 没有找到用户名密码分隔符，返回掩码
		return "******"
	}

	username := userPwdSplit[0]
	restParts := strings.SplitN(userPwdSplit[1], "@", 2)
	if len(restParts) < 2 {
		// 没有找到密码和主机分隔符，返回掩码
		return username + ":******"
	}

	// 返回掩码后的DSN
	return username + ":******@" + restParts[1]
}
