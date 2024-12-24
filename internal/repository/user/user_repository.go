package user

import (
	"gorm.io/gorm"
)

// RepositoryUser 用户数据操作接口
type RepositoryUser interface {
	Create(user *Users) error
	GetByID(id uint64) (*Users, error)
	GetAll() ([]Users, error)
	Delete(id uint) error
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) RepositoryUser {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(user *Users) error {
	return r.db.Create(user).Error
}

func (r *userRepositoryImpl) GetByID(id uint64) (*Users, error) {
	var user Users
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetAll() ([]Users, error) {
	var users []Users
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&Users{}, id).Error
}
