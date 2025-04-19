package apikey

import (
	"errors"
	"goWebExample/internal/infra/db/mysql"
	"time"

	"gorm.io/gorm"
)

var (
	// ErrDBNotConnected 数据库未连接错误
	ErrDBNotConnected = errors.New("数据库未连接")
	// ErrAPIKeyNotFound API密钥未找到错误
	ErrAPIKeyNotFound = errors.New("API密钥未找到")
	// ErrAPIKeyDisabled API密钥已禁用错误
	ErrAPIKeyDisabled = errors.New("API密钥已禁用")
	// ErrAPIKeyExpired API密钥已过期错误
	ErrAPIKeyExpired = errors.New("API密钥已过期")
)

// RepositoryAPIKey API密钥数据操作接口
type RepositoryAPIKey interface {
	GetDB() *gorm.DB
	GetByAPIKey(apiKey string) (*APIKey, error)
	Create(apiKey *APIKey) error
	Update(apiKey *APIKey) error
	Delete(id uint64) error
	GetAll() ([]APIKey, error)
}

// apiKeyRepositoryImpl API密钥仓库实现
type apiKeyRepositoryImpl struct {
	dbConnector *mysql.DBConnector
}

// NewAPIKeyRepository 创建API密钥仓库
func NewAPIKeyRepository(dbConnector *mysql.DBConnector) RepositoryAPIKey {
	return &apiKeyRepositoryImpl{dbConnector: dbConnector}
}

// GetDB 获取数据库连接
func (r *apiKeyRepositoryImpl) GetDB() *gorm.DB {
	return r.dbConnector.GetDB()
}

// GetByAPIKey 根据API密钥获取记录
func (r *apiKeyRepositoryImpl) GetByAPIKey(apiKey string) (*APIKey, error) {
	db := r.GetDB()
	if db == nil {
		return nil, ErrDBNotConnected
	}

	var key APIKey
	if err := db.Where("api_key = ?", apiKey).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	// 检查API密钥状态
	if key.Status != 1 {
		return nil, ErrAPIKeyDisabled
	}

	// 检查API密钥是否过期
	if key.ExpiredAt.Before(time.Now()) {
		return nil, ErrAPIKeyExpired
	}

	return &key, nil
}

// Create 创建API密钥
func (r *apiKeyRepositoryImpl) Create(apiKey *APIKey) error {
	db := r.GetDB()
	if db == nil {
		return ErrDBNotConnected
	}
	return db.Create(apiKey).Error
}

// Update 更新API密钥
func (r *apiKeyRepositoryImpl) Update(apiKey *APIKey) error {
	db := r.GetDB()
	if db == nil {
		return ErrDBNotConnected
	}
	return db.Save(apiKey).Error
}

// Delete 删除API密钥
func (r *apiKeyRepositoryImpl) Delete(id uint64) error {
	db := r.GetDB()
	if db == nil {
		return ErrDBNotConnected
	}
	return db.Delete(&APIKey{}, id).Error
}

// GetAll 获取所有API密钥
func (r *apiKeyRepositoryImpl) GetAll() ([]APIKey, error) {
	db := r.GetDB()
	if db == nil {
		return nil, ErrDBNotConnected
	}

	var keys []APIKey
	if err := db.Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}
