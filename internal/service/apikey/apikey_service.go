package apikey

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"goWebExample/internal/repository/apikey"
	"time"

	"go.uber.org/zap"
)

const ServiceName = "apikey"

// APIKeyDTO API密钥数据传输对象
type APIKeyDTO struct {
	ID          uint64 `json:"id,omitempty"`
	APIKey      string `json:"apiKey,omitempty"`
	Status      int    `json:"status,omitempty"`
	ExpiredAt   string `json:"expiredAt,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ServiceAPIKey 定义API密钥服务接口
type ServiceAPIKey interface {
	GetByAPIKey(apiKey string) (*APIKeyDTO, error)
	Create(apiKey *apikey.APIKey) error
	Update(apiKey *apikey.APIKey) error
	Delete(id uint64) error
	GetAll() ([]APIKeyDTO, error)
	VerifySign(apiKey, sign, timestamp string) (bool, error)
}

// APIKeyService 提供API密钥业务服务
type APIKeyService struct {
	repo   apikey.RepositoryAPIKey
	logger *zap.Logger
}

// NewAPIKeyService 创建 APIKeyService 实例
func NewAPIKeyService(repo apikey.RepositoryAPIKey, logger *zap.Logger) *APIKeyService {
	return &APIKeyService{
		repo:   repo,
		logger: logger,
	}
}

// formatTime 格式化时间
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// toDTO 将 APIKey 模型转换为 DTO
func toDTO(a *apikey.APIKey) *APIKeyDTO {
	if a == nil {
		return nil
	}

	return &APIKeyDTO{
		ID:          a.ID,
		APIKey:      a.APIKey,
		Status:      a.Status,
		ExpiredAt:   formatTime(a.ExpiredAt),
		Description: a.Description,
		CreatedAt:   formatTime(a.CreatedAt),
		UpdatedAt:   formatTime(a.UpdatedAt),
	}
}

// GetByAPIKey 根据API密钥获取记录
func (s *APIKeyService) GetByAPIKey(apiKey string) (*APIKeyDTO, error) {
	s.logger.Info("获取API密钥", zap.String("apiKey", apiKey))

	// 调用仓库层获取API密钥信息
	apiKeyInfo, err := s.repo.GetByAPIKey(apiKey)
	if err != nil {
		s.logger.Error("获取API密钥失败", zap.String("apiKey", apiKey), zap.Error(err))
		return nil, err
	}

	return toDTO(apiKeyInfo), nil
}

// Create 创建API密钥
func (s *APIKeyService) Create(apiKey *apikey.APIKey) error {
	s.logger.Info("创建API密钥", zap.String("apiKey", apiKey.APIKey))
	return s.repo.Create(apiKey)
}

// Update 更新API密钥
func (s *APIKeyService) Update(apiKey *apikey.APIKey) error {
	s.logger.Info("更新API密钥", zap.Uint64("id", apiKey.ID))
	return s.repo.Update(apiKey)
}

// Delete 删除API密钥
func (s *APIKeyService) Delete(id uint64) error {
	s.logger.Info("删除API密钥", zap.Uint64("id", id))
	return s.repo.Delete(id)
}

// GetAll 获取所有API密钥
func (s *APIKeyService) GetAll() ([]APIKeyDTO, error) {
	s.logger.Info("获取所有API密钥")

	// 调用仓库层获取所有API密钥
	apiKeys, err := s.repo.GetAll()
	if err != nil {
		s.logger.Error("获取所有API密钥失败", zap.Error(err))
		return nil, err
	}

	// 转换为DTO
	dtos := make([]APIKeyDTO, len(apiKeys))
	for i, key := range apiKeys {
		dto := toDTO(&key)
		dtos[i] = *dto
	}

	return dtos, nil
}

// VerifySign 验证签名
func (s *APIKeyService) VerifySign(apiKey, sign, timestamp string) (bool, error) {
	s.logger.Info("验证签名", zap.String("apiKey", apiKey), zap.String("timestamp", timestamp))

	// 获取API密钥信息
	apiKeyInfo, err := s.repo.GetByAPIKey(apiKey)
	if err != nil {
		s.logger.Error("获取API密钥失败", zap.String("apiKey", apiKey), zap.Error(err))
		return false, err
	}

	// 生成预期的签名
	expectedSign := generateSign(apiKey, apiKeyInfo.APISecret, timestamp)

	// 比较签名
	if sign != expectedSign {
		s.logger.Warn("签名验证失败",
			zap.String("provided", sign),
			zap.String("expected", expectedSign))
		return false, nil
	}

	return true, nil
}

// generateSign 生成签名
// 签名算法: HMAC-SHA256(apiKey + timestamp, apiSecret)，输出为十六进制字符串
func generateSign(apiKey, apiSecret, timestamp string) string {
	// 组合原始字符串
	message := apiKey + timestamp

	// 创建HMAC-SHA256哈希
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(message))

	// 返回十六进制编码的签名
	return hex.EncodeToString(h.Sum(nil))
}
