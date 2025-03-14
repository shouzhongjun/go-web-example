package user

import (
	"gorm.io/gorm"
	"time"
)

type Users struct {
	ID                  uint64         `gorm:"primaryKey;autoIncrement;comment:'主键ID'" json:"ID,omitempty"`
	UUID                string         `gorm:"type:char(36);unique;not null;comment:'全局唯一标识符'" json:"UUID,omitempty"`
	Username            string         `gorm:"type:varchar(50);unique;not null;comment:'用户名'" json:"username,omitempty"`
	PasswordHash        string         `gorm:"type:varchar(255);not null;comment:'密码哈希值'" json:"-"`
	Email               string         `gorm:"type:varchar(255);unique;not null;comment:'电子邮箱'" json:"email,omitempty"`
	EmailVerified       bool           `gorm:"type:tinyint(1);default:0;not null;comment:'邮箱验证状态'" json:"emailVerified,omitempty"`
	PhoneCountryCode    *string        `gorm:"type:varchar(5);comment:'国际电话区号'" json:"phoneCountryCode,omitempty"`
	PhoneNumber         *string        `gorm:"type:varchar(20);comment:'手机号码'" json:"phoneNumber,omitempty"`
	FirstName           *string        `gorm:"type:varchar(50);comment:'名字'" json:"firstName,omitempty"`
	LastName            *string        `gorm:"type:varchar(50);comment:'姓氏'" json:"lastName,omitempty"`
	Gender              string         `gorm:"type:enum('male','female','other','unknown');default:'unknown';comment:'性别'" json:"gender,omitempty"`
	Birthdate           *time.Time     `gorm:"type:date;comment:'出生日期'" json:"birthdate,omitempty"`
	AvatarURL           *string        `gorm:"type:varchar(512);comment:'头像URL'" json:"avatarURL,omitempty"`
	Timezone            string         `gorm:"type:varchar(50);default:'UTC';comment:'时区设置'" json:"timezone,omitempty"`
	Locale              string         `gorm:"type:varchar(10);default:'en-US';comment:'语言地区设置'" json:"locale,omitempty"`
	IsActive            bool           `gorm:"type:tinyint;default:1;not null;comment:'账户激活状态'" json:"isActive,omitempty"`
	IsSuperuser         bool           `gorm:"type:tinyint;default:0;not null;comment:'超级管理员标志'" json:"isSuperuser,omitempty"`
	Is2FAEnabled        bool           `gorm:"column:is_2fa_enabled;type:tinyint;default:0;not null;comment:'双重认证状态'" json:"is2FAEnabled,omitempty"`
	LastLogin           *time.Time     `gorm:"type:datetime;comment:'最后登录时间'" json:"lastLogin,omitempty"`
	CreatedAt           time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP;not null;comment:'创建时间'" json:"createdAt"`
	UpdatedAt           time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP;not null;autoUpdateTime;comment:'更新时间'" json:"updatedAt"`
	DeletedAt           gorm.DeletedAt `gorm:"type:datetime;index;comment:'软删除时间'" json:"deletedAt"`
	RegistrationIP      *string        `gorm:"type:varchar(45);comment:'注册IP地址'" json:"registrationIP,omitempty"`
	LastLoginIP         *string        `gorm:"type:varchar(45);comment:'最后登录IP'" json:"lastLoginIP,omitempty"`
	SecurityStamp       *string        `gorm:"type:varchar(100);comment:'安全验证戳'" json:"securityStamp,omitempty"`
	PasswordSalt        string         `gorm:"type:varchar(100);not null;comment:'密码盐值'" json:"-"`
	FailedLoginAttempts *int           `gorm:"type:int;default:0;comment:'连续登录失败次数'" json:"failedLoginAttempts,omitempty"`
	LockoutEnd          *time.Time     `gorm:"type:datetime;comment:'账户锁定截止时间'" json:"-"`
	TwoFactorSecret     *string        `gorm:"type:varchar(100);comment:'双重认证秘钥'" json:"twoFactorSecret,omitempty"`
	RecoveryCodes       *string        `gorm:"type:json;comment:'恢复代码'" json:"recoveryCodes,omitempty"`
	AddressCountry      *string        `gorm:"type:varchar(100);comment:'国家'" json:"addressCountry,omitempty"`
	AddressState        *string        `gorm:"type:varchar(100);comment:'省/州'" json:"addressState,omitempty"`
	AddressCity         *string        `gorm:"type:varchar(100);comment:'城市'" json:"addressCity,omitempty"`
	AddressStreet       *string        `gorm:"type:varchar(255);comment:'街道地址'" json:"addressStreet,omitempty"`
	AddressPostalCode   *string        `gorm:"type:varchar(20);comment:'邮政编码'" json:"addressPostalCode,omitempty"`
	Metadata            *string        `gorm:"type:json;comment:'扩展元数据'" json:"metadata,omitempty"`
}

func (u *Users) TableName() string {
	return "t_users"
}
