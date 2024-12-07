package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ProvideDB(config *DBConfig) *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.DSN), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}
