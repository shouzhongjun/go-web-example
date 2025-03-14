package user

import (
	"fmt"
	"goWebExample/internal/repository/user"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const ServiceName = "user"
const TimeFormat = "2006-01-02 15:04:05"

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID               uint64  `json:"ID,omitempty"`
	UUID             string  `json:"UUID,omitempty"`
	Username         string  `json:"username,omitempty"`
	Email            string  `json:"email,omitempty"`
	EmailVerified    bool    `json:"emailVerified,omitempty"`
	PhoneCountryCode *string `json:"phoneCountryCode,omitempty"`
	PhoneNumber      *string `json:"phoneNumber,omitempty"`
	FirstName        *string `json:"firstName,omitempty"`
	LastName         *string `json:"lastName,omitempty"`
	Gender           string  `json:"gender,omitempty"`
	Birthdate        string  `json:"birthdate,omitempty"`
	AvatarURL        *string `json:"avatarURL,omitempty"`
	Timezone         string  `json:"timezone,omitempty"`
	Locale           string  `json:"locale,omitempty"`
	IsActive         bool    `json:"isActive,omitempty"`
	IsSuperuser      bool    `json:"isSuperuser,omitempty"`
	Is2FAEnabled     bool    `json:"is2FAEnabled,omitempty"`
	LastLogin        string  `json:"lastLogin,omitempty"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
	RegistrationIP   *string `json:"registrationIP,omitempty"`
	LastLoginIP      *string `json:"lastLoginIP,omitempty"`
}

// formatTime 格式化时间
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(TimeFormat)
}

// toDTO 将 Users 模型转换为 DTO
func toDTO(u *user.Users) *UserDTO {
	if u == nil {
		return nil
	}

	return &UserDTO{
		ID:               u.ID,
		UUID:             u.UUID,
		Username:         u.Username,
		Email:            u.Email,
		EmailVerified:    u.EmailVerified,
		PhoneCountryCode: u.PhoneCountryCode,
		PhoneNumber:      u.PhoneNumber,
		FirstName:        u.FirstName,
		LastName:         u.LastName,
		Gender:           u.Gender,
		Birthdate:        formatTime(u.Birthdate),
		AvatarURL:        u.AvatarURL,
		Timezone:         u.Timezone,
		Locale:           u.Locale,
		IsActive:         u.IsActive,
		IsSuperuser:      u.IsSuperuser,
		Is2FAEnabled:     u.Is2FAEnabled,
		LastLogin:        formatTime(u.LastLogin),
		CreatedAt:        formatTime(&u.CreatedAt),
		UpdatedAt:        formatTime(&u.UpdatedAt),
		RegistrationIP:   u.RegistrationIP,
		LastLoginIP:      u.LastLoginIP,
	}
}

// ServerUser 定义用户服务接口
type ServerUser interface {
	GetUserDetail(userID string) (*UserDTO, error)
}

// UserService 提供用户业务服务
type UserService struct {
	repo   user.RepositoryUser
	logger *zap.Logger
}

// NewUserService 创建 UserService 实例
func NewUserService(repo user.RepositoryUser, logger *zap.Logger) *UserService {
	return &UserService{repo: repo, logger: logger}
}

// GetUserDetail 根据 userID 获取用户详细信息
func (s *UserService) GetUserDetail(userID string) (*UserDTO, error) {
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

	return toDTO(userInfo), nil
}

func (s *UserService) Login(username string, password string, ip string) (*UserDTO, error) {
	s.logger.Info("用户登录", zap.String("username", username), zap.String("ip", ip))

	// 1. 根据用户名获取用户信息
	userInfo, err := s.repo.GetUserByUsername(username)
	if err != nil {
		s.logger.Error("用户不存在", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("用户不存在或密码错误")
	}

	// 2. 验证密码
	if userInfo.PasswordHash != password { // 注意：实际应用中应该使用安全的密码哈希比较
		s.logger.Warn("密码错误", zap.String("username", username))
		return nil, fmt.Errorf("用户不存在或密码错误")
	}
	if userInfo.LockoutEnd != nil && time.Now().Before(*userInfo.LockoutEnd) || userInfo.IsActive == false {
		s.logger.Warn("用户被锁定", zap.String("username", username))
		return nil, fmt.Errorf("用户被锁定")
	}

	// 3. 更新登录信息
	if err := s.repo.UpdateLoginInfo(userInfo.ID, ip); err != nil {
		s.logger.Error("更新登录信息失败", zap.String("username", username), zap.Error(err))
		return toDTO(userInfo), nil
	}

	return toDTO(userInfo), nil
}
