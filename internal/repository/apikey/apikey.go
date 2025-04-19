package apikey

import (
	"gorm.io/gorm"
	"time"
)

// APIKey 表示API密钥数据模型
type APIKey struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement;comment:'主键ID'" json:"id,omitempty"`
	APIKey      string         `gorm:"type:varchar(64);unique;not null;comment:'API密钥'" json:"apiKey,omitempty"`
	APISecret   string         `gorm:"type:varchar(128);not null;comment:'API密钥对应的秘钥'" json:"-"`
	Status      int            `gorm:"type:tinyint;default:1;not null;comment:'状态：0-禁用，1-启用'" json:"status,omitempty"`
	ExpiredAt   time.Time      `gorm:"type:timestamp;not null;comment:'过期时间'" json:"expiredAt,omitempty"`
	Description string         `gorm:"type:varchar(255);comment:'描述'" json:"description,omitempty"`
	CreatedAt   time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not null;comment:'创建时间'" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not null;comment:'更新时间'" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (a *APIKey) TableName() string {
	return "api_keys"
}
