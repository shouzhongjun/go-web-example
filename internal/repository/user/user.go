package user

import "time"

// User 数据库模型定义
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"type:varchar(100);not null"`
	Email     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 返回对应数据库表名
func (User) TableName() string {
	return "users" // 显式指定表名
}
