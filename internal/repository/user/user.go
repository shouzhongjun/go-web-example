package user

// Users  数据库模型定义
type Users struct {
	Id       int64  `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id,omitempty"`
	Password string `gorm:"column:password;NOT NULL;comment:'密码'" json:"password"`
	UserRole int64  `gorm:"column:user_role;comment:'角色'" json:"user_role,omitempty"`
	UserName string `gorm:"column:user_name;NOT NULL;comment:'用户名'" json:"user_name"`
	Source   string `gorm:"column:source;default:local;NOT NULL;comment:'用户来源'" json:"source"`
}

// TableName 返回对应数据库表名
func (Users) TableName() string {
	return "users" // 显式指定表名
}
