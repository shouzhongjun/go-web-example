package user

import (
	"fmt"
	"goWebExample/internal/repository/user"
	"strconv"
)

const ServiceName = "user"

// ServerUser 定义用户服务接口
type ServerUser interface {
	GetUserDetail(userID string) (*user.Users, error)
}

// UserService 提供用户业务服务
type UserService struct {
	repo user.RepositoryUser
}

// NewUserService 创建 UserService 实例
func NewUserService(repo user.RepositoryUser) *UserService {
	return &UserService{repo: repo}
}

// GetUserDetail 根据 userID 获取用户详细信息
func (s *UserService) GetUserDetail(userID string) (*user.Users, error) {
	// string 转 uint
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid userID: %w", err)
	}

	// 调用 Repo 获取用户信息
	userInfo, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}
