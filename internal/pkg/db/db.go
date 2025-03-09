package db

import (
	"goWebExample/internal/configs"
	initlogger "goWebExample/pkg/logger"
	"log"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func NewGormConfig(config *configs.AllConfig, logger *zap.Logger) *gorm.DB {
	gormZap := initlogger.NewGormZap(logger, gormlogger.Info)
	g := &gorm.Config{
		Logger: gormZap,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	db, err := gorm.Open(mysql.Open(config.Database.Dsn()), g)
	if err != nil {
		log.Fatalf("数据库连接失败:%s", err)
		return nil
	}
	return db
}
