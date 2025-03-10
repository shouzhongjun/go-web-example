package user

import (
	"goWebExample/pkg/infrastructure/db"

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
	dbConnector *db.DBConnector
}

func NewUserRepository(dbConnector *db.DBConnector) RepositoryUser {
	return &userRepositoryImpl{dbConnector: dbConnector}
}

// getDB 获取数据库连接
func (r *userRepositoryImpl) getDB() *gorm.DB {
	if r.dbConnector == nil || !r.dbConnector.IsConnected() {
		return nil
	}
	return r.dbConnector.GetTypedClient()
}

func (r *userRepositoryImpl) Create(user *Users) error {
	db := r.getDB()
	if db == nil {
		return ErrDBNotConnected
	}
	return db.Create(user).Error
}

func (r *userRepositoryImpl) GetByID(id uint64) (*Users, error) {
	db := r.getDB()
	if db == nil {
		return nil, ErrDBNotConnected
	}

	var user Users
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetAll() ([]Users, error) {
	db := r.getDB()
	if db == nil {
		return nil, ErrDBNotConnected
	}

	var users []Users
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepositoryImpl) Delete(id uint) error {
	db := r.getDB()
	if db == nil {
		return ErrDBNotConnected
	}

	return db.Delete(&Users{}, id).Error
}
