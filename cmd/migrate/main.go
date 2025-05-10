package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"goWebExample/internal/configs"
	"goWebExample/internal/infra/db/mysql"
	zaplog "goWebExample/pkg/zap"
)

// 定义命令行参数
var (
	configPath = flag.String("conf", "configs/config.dev.yaml", "配置文件路径")
	migrateDir = flag.String("dir", "migrations", "迁移文件目录")
)

func main() {
	// 解析命令行参数
	flag.Parse()

	// 初始化日志
	logger := initLogger()
	defer logger.Sync()

	// 加载配置
	config := loadConfig(logger)
	if config == nil {
		logger.Fatal("加载配置失败")
		return
	}

	// 连接数据库
	db := connectDatabase(logger, &config.Database)
	if db == nil {
		logger.Fatal("连接数据库失败")
		return
	}
	defer func() {
		if err := db.Disconnect(context.Background()); err != nil {
			logger.Error("关闭数据库连接失败", zap.Error(err))
		}
	}()

	// 执行迁移
	if err := runMigrations(logger, db, *migrateDir); err != nil {
		logger.Fatal("执行迁移失败", zap.Error(err))
		return
	}

	logger.Info("数据库迁移完成")
}

// initLogger 初始化日志
func initLogger() *zap.Logger {
	// 创建基本的开发环境日志配置
	config := &configs.AllConfig{
		Log: configs.Log{
			Level:         "debug",
			EnableConsole: true,
			EnableFile:    false,
			Prefix:        "migrate",
		},
	}
	return zaplog.NewZap(config)
}

// loadConfig 加载配置
func loadConfig(logger *zap.Logger) *configs.AllConfig {
	logger.Info("加载配置文件", zap.String("path", *configPath))
	return configs.ReadConfig(*configPath)
}

// connectDatabase 连接数据库
func connectDatabase(logger *zap.Logger, dbConfig *configs.Database) *mysql.DBConnector {
	logger.Info("连接数据库",
		zap.String("host", dbConfig.Host),
		zap.Int("port", dbConfig.Port),
		zap.String("database", dbConfig.DBName))

	// 创建数据库连接器
	db := mysql.NewDBConnector(dbConfig, logger)
	if db == nil {
		logger.Error("创建数据库连接器失败")
		return nil
	}

	// 连接数据库
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		logger.Error("连接数据库失败", zap.Error(err))
		return nil
	}

	// 检查数据库连接
	if ok, err := db.HealthCheck(ctx); !ok {
		logger.Error("数据库健康检查失败", zap.Error(err))
		return nil
	}

	logger.Info("数据库连接成功")
	return db
}

// runMigrations 执行迁移
func runMigrations(logger *zap.Logger, db *mysql.DBConnector, migrateDir string) error {
	// 检查迁移目录是否存在
	if _, err := os.Stat(migrateDir); os.IsNotExist(err) {
		return fmt.Errorf("迁移目录不存在: %s", migrateDir)
	}

	// 获取所有SQL迁移文件
	files, err := filepath.Glob(filepath.Join(migrateDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %w", err)
	}

	// 按文件名排序
	sort.Strings(files)

	// 创建迁移表（如果不存在）
	if err := createMigrationTable(logger, db); err != nil {
		return fmt.Errorf("创建迁移表失败: %w", err)
	}

	// 获取已执行的迁移
	executedMigrations, err := getExecutedMigrations(logger, db)
	if err != nil {
		return fmt.Errorf("获取已执行迁移失败: %w", err)
	}

	// 执行每个迁移文件
	for _, file := range files {
		filename := filepath.Base(file)

		// 检查是否已执行
		if contains(executedMigrations, filename) {
			logger.Info("迁移已执行，跳过", zap.String("file", filename))
			continue
		}

		// 读取SQL文件内容
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("读取迁移文件失败 %s: %w", filename, err)
		}

		// 如果文件为空，记录并继续
		if len(strings.TrimSpace(string(content))) == 0 {
			logger.Warn("迁移文件为空，跳过", zap.String("file", filename))

			// 记录空文件迁移为已执行
			if err := recordMigration(logger, db, filename); err != nil {
				return fmt.Errorf("记录迁移失败 %s: %w", filename, err)
			}

			continue
		}

		// 执行SQL
		logger.Info("执行迁移", zap.String("file", filename))
		if err := executeMigration(logger, db, filename, string(content)); err != nil {
			return fmt.Errorf("执行迁移失败 %s: %w", filename, err)
		}

		logger.Info("迁移成功", zap.String("file", filename))
	}

	return nil
}

// createMigrationTable 创建迁移表
func createMigrationTable(logger *zap.Logger, db *mysql.DBConnector) error {
	logger.Info("检查迁移表是否存在")

	sql := `
	CREATE TABLE IF NOT EXISTS migrations (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE KEY unique_migration_name (name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	return db.GetDB().Exec(sql).Error
}

// getExecutedMigrations 获取已执行的迁移
func getExecutedMigrations(logger *zap.Logger, db *mysql.DBConnector) ([]string, error) {
	var migrations []string
	var results []struct {
		Name string
	}

	if err := db.GetDB().Raw("SELECT name FROM migrations").Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, result := range results {
		migrations = append(migrations, result.Name)
	}

	return migrations, nil
}

// executeMigration 执行迁移
func executeMigration(logger *zap.Logger, db *mysql.DBConnector, filename, content string) error {
	// 使用事务执行迁移
	return db.Transaction(context.Background(), func(tx *gorm.DB) error {
		// 执行SQL语句
		if err := tx.Exec(content).Error; err != nil {
			return err
		}

		// 记录迁移
		if err := tx.Exec("INSERT INTO migrations (name) VALUES (?)", filename).Error; err != nil {
			return err
		}

		return nil
	})
}

// recordMigration 记录迁移为已执行
func recordMigration(logger *zap.Logger, db *mysql.DBConnector, filename string) error {
	return db.GetDB().Exec("INSERT INTO migrations (name) VALUES (?)", filename).Error
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
