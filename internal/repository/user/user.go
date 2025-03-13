package user

import (
	"gorm.io/gorm"
	"time"
)

type Users struct {
	ID                  uint64         `gorm:"primaryKey;autoIncrement;comment:'主键ID'"`
	UUID                string         `gorm:"type:char(36);unique;not null;comment:'全局唯一标识符'"`
	Username            string         `gorm:"type:varchar(50);unique;not null;comment:'用户名'"`
	PasswordHash        string         `gorm:"type:varchar(255);not null;comment:'密码哈希值'"`
	Email               string         `gorm:"type:varchar(255);unique;not null;comment:'电子邮箱'"`
	EmailVerified       bool           `gorm:"type:tinyint(1);default:0;not null;comment:'邮箱验证状态'"`
	PhoneCountryCode    *string        `gorm:"type:varchar(5);comment:'国际电话区号'"`
	PhoneNumber         *string        `gorm:"type:varchar(20);comment:'手机号码'"`
	FirstName           *string        `gorm:"type:varchar(50);comment:'名字'"`
	LastName            *string        `gorm:"type:varchar(50);comment:'姓氏'"`
	Gender              string         `gorm:"type:enum('male','female','other','unknown');default:'unknown';comment:'性别'"`
	Birthdate           *time.Time     `gorm:"type:date;comment:'出生日期'"`
	AvatarURL           *string        `gorm:"type:varchar(512);comment:'头像URL'"`
	Timezone            string         `gorm:"type:varchar(50);default:'UTC';comment:'时区设置'"`
	Locale              string         `gorm:"type:varchar(10);default:'en-US';comment:'语言地区设置'"`
	IsActive            bool           `gorm:"type:tinyint;default:1;not null;comment:'账户激活状态'"`
	IsSuperuser         bool           `gorm:"type:tinyint;default:0;not null;comment:'超级管理员标志'"`
	Is2FAEnabled        bool           `gorm:"column:is_2fa_enabled;type:tinyint;default:0;not null;comment:'双重认证状态'"`
	LastLogin           *time.Time     `gorm:"type:datetime;comment:'最后登录时间'"`
	CreatedAt           time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP;not null;comment:'创建时间'"`
	UpdatedAt           time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP;not null;autoUpdateTime;comment:'更新时间'"`
	DeletedAt           gorm.DeletedAt `gorm:"type:datetime;index;comment:'软删除时间'"`
	RegistrationIP      *string        `gorm:"type:varchar(45);comment:'注册IP地址'"`
	LastLoginIP         *string        `gorm:"type:varchar(45);comment:'最后登录IP'"`
	SecurityStamp       *string        `gorm:"type:varchar(100);comment:'安全验证戳'"`
	PasswordSalt        string         `gorm:"type:varchar(100);not null;comment:'密码盐值'"`
	FailedLoginAttempts *int           `gorm:"type:int;default:0;comment:'连续登录失败次数'"`
	LockoutEnd          *time.Time     `gorm:"type:datetime;comment:'账户锁定截止时间'"`
	TwoFactorSecret     *string        `gorm:"type:varchar(100);comment:'双重认证秘钥'"`
	RecoveryCodes       *string        `gorm:"type:json;comment:'恢复代码'"`
	AddressCountry      *string        `gorm:"type:varchar(100);comment:'国家'"`
	AddressState        *string        `gorm:"type:varchar(100);comment:'省/州'"`
	AddressCity         *string        `gorm:"type:varchar(100);comment:'城市'"`
	AddressStreet       *string        `gorm:"type:varchar(255);comment:'街道地址'"`
	AddressPostalCode   *string        `gorm:"type:varchar(20);comment:'邮政编码'"`
	Metadata            *string        `gorm:"type:json;comment:'扩展元数据'"`
}

func (u *Users) TableName() string {
	return "t_users"
}
