package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	SecretKey string        `mapstructure:"secretKey"`
	Issuer    string        `mapstructure:"issuer"`
	Duration  time.Duration `mapstructure:"duration"`
}

type JwtManager struct {
	config Config
}

func NewJWTManager(config Config) *JwtManager {
	return &JwtManager{
		config: config,
	}
}

// CustomClaims 自定义 Claims
type CustomClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
	IsAdmin  bool   `json:"is_admin"`
}

// GenerateToken 生成 token
func (m *JwtManager) GenerateToken(userID, username, nickname string, isAdmin bool) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.Duration)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:   userID,
		Username: username,
		Nickname: nickname,
		IsAdmin:  isAdmin,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.SecretKey))
}

// ParseToken 解析 token
func (m *JwtManager) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		// 只检查过期错误，其他错误统一处理
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token已过期")
		}
		return nil, fmt.Errorf("token无效: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的token")
}

// ValidateToken 验证 token
func (m *JwtManager) ValidateToken(tokenString string) bool {
	_, err := m.ParseToken(tokenString)
	return err == nil
}
