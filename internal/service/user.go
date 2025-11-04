package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// UserService 用户管理服务
type UserService struct {
	db     *gorm.DB
	cfg    *config.Config
	logger *zap.Logger
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Email string `json:"email"`
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *UserService {
	return &UserService{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}
}

// CreateUser 创建新用户
func (s *UserService) CreateUser(req CreateUserRequest) (*model.User, error) {
	// 1. 参数验证
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if !isValidEmail(req.Email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// 2. 检查邮箱是否已存在
	var existingUser model.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("email already exists")
	}

	// 3. 生成 API 凭证
	apiKey, apiSecret := generateAPICredentials()

	// 4. 创建用户
	user := &model.User{
		Email:     req.Email,
		APIKey:    apiKey,
		APISecret: apiSecret,
		Status:    "active",
	}

	if err := s.db.Create(user).Error; err != nil {
		s.logger.Error("Failed to create user",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User created successfully",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByAPIKey 根据 API Key 获取用户
func (s *UserService) GetUserByAPIKey(apiKey string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// RegenerateAPIKey 重新生成用户的 API Key 和 Secret
func (s *UserService) RegenerateAPIKey(userID uint) (*model.User, error) {
	// 1. 获取用户
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// 2. 生成新的凭证
	apiKey, apiSecret := generateAPICredentials()

	// 3. 更新用户
	user.APIKey = apiKey
	user.APISecret = apiSecret

	if err := s.db.Save(user).Error; err != nil {
		s.logger.Error("Failed to regenerate API key",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to regenerate API key: %w", err)
	}

	s.logger.Info("API key regenerated",
		zap.Uint("user_id", userID),
	)

	return user, nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(userID uint, status string) (*model.User, error) {
	// 1. 验证状态
	validStatuses := map[string]bool{
		"active":    true,
		"inactive":  true,
		"suspended": true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status: must be one of active, inactive, suspended")
	}

	// 2. 获取用户
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// 3. 更新状态
	user.Status = status
	if err := s.db.Save(user).Error; err != nil {
		s.logger.Error("Failed to update user status",
			zap.Uint("user_id", userID),
			zap.String("status", status),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	s.logger.Info("User status updated",
		zap.Uint("user_id", userID),
		zap.String("status", status),
	)

	return user, nil
}

// generateAPICredentials 生成 API Key 和 Secret
func generateAPICredentials() (apiKey string, apiSecret string) {
	// API Key: 32 字节 = 64 hex 字符
	apiKeyBytes := make([]byte, 32)
	rand.Read(apiKeyBytes)
	apiKey = hex.EncodeToString(apiKeyBytes)

	// API Secret: 48 字节 = 96 hex 字符
	apiSecretBytes := make([]byte, 48)
	rand.Read(apiSecretBytes)
	apiSecret = hex.EncodeToString(apiSecretBytes)

	return apiKey, apiSecret
}

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	// 简单的邮箱正则验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
