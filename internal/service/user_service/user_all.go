package user_service

import "goWebExample/internal/repository/user"

// UserService 提供用户业务服务
type UserService struct {
	repo user.RepositoryUser
}

// NewUserService 创建 UserService 实例
func NewUserService(repo user.RepositoryUser) *UserService {
	return &UserService{repo: repo}
}

// GetUserDetail 示例方法
func (s *UserService) GetUserDetail(userID int) string {
	return "User detail for ID: " + string(userID)
}
