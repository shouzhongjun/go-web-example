package user

import (
	"database/sql"
)

// Users  数据库模型定义
type Users struct {
	Id       sql.NullInt32 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	Password string        `gorm:"column:password;NOT NULL;comment:'密码'"`
	UserRole sql.NullInt32 `gorm:"column:user_role;comment:'角色'"`
	UserName string        `gorm:"column:user_name;NOT NULL;comment:'用户名'"`
	Source   string        `gorm:"column:source;default:local;NOT NULL;comment:'用户来源'"`
}

// TableName 返回对应数据库表名
func (Users) TableName() string {
	return "users" // 显式指定表名
}
