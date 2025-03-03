package db

import (
	"goWebExample/internal/configs"
	"goWebExample/pkg/utils"
	"log"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func NewGormConfig(config *configs.AllConfig, logger *zap.Logger) *gorm.DB {
	gormZap := utils.NewGormZap(logger, gormlogger.Info)
	g := &gorm.Config{
		Logger: gormZap,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	db, err := gorm.Open(mysql.Open(config.Database.GetDSN()), g)
	if err != nil {
		log.Fatalf("数据库连接失败:%s", err)
		return nil
	}
	return db
}
