package user

import (
	"goWebExample/internal/infra/db/mysql"

	"gorm.io/gorm"
)

// RepositoryUser 用户数据操作接口
type RepositoryUser interface {
	GetDB() *gorm.DB
	Create(user *Users) error
	GetByID(id uint64) (*Users, error)
	GetAll() ([]Users, error)
	Delete(id uint) error
	GetUserByUsername(username string) (*Users, error)
	UpdateLoginInfo(userID uint64, ip string) error
}

type userRepositoryImpl struct {
	dbConnector *mysql.DBConnector
}

func NewUserRepository(dbConnector *mysql.DBConnector) RepositoryUser {
	return &userRepositoryImpl{dbConnector: dbConnector}
}

func (r *userRepositoryImpl) GetDB() *gorm.DB {
	return r.dbConnector.GetDB()
}

func (r *userRepositoryImpl) Create(user *Users) error {
	db := r.GetDB()
	if db == nil {
		return ErrDBNotConnected
	}
	return db.Create(user).Error
}

func (r *userRepositoryImpl) GetByID(id uint64) (*Users, error) {
	db := r.GetDB()
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
	db := r.GetDB()
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
	db := r.GetDB()
	if db == nil {
		return ErrDBNotConnected
	}

	return db.Delete(&Users{}, id).Error
}

func (r *userRepositoryImpl) GetUserByUsername(username string) (*Users, error) {
	db := r.GetDB()
	if db == nil {
		return nil, ErrDBNotConnected
	}

	var user Users
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) UpdateLoginInfo(userID uint64, ip string) error {
	db := r.GetDB()
	if db == nil {
		return ErrDBNotConnected
	}

	tx := db.Model(&Users{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login":    gorm.Expr("NOW()"),
		"last_login_ip": ip,
	})
	return tx.Error
}
