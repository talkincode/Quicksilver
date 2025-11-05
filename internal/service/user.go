package service

import (
	"crypto/rand"
	"encoding/base64"
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
func (s *UserService) CreateUser(req CreateUserRequest) (*model.User, string, error) {
	// 1. 参数验证
	if req.Email == "" {
		return nil, "", fmt.Errorf("email is required")
	}

	if !isValidEmail(req.Email) {
		return nil, "", fmt.Errorf("invalid email format")
	}

	// 2. 检查邮箱是否已存在
	var existingUser model.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, "", fmt.Errorf("email already exists")
	}

	// 3. 生成 API 凭证
	apiKey, apiSecret, err := s.generateAPICredentials()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API credentials: %w", err)
	}

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
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User created successfully",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return user, apiSecret, nil
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

// ListUsers 获取用户列表（支持分页和搜索）
func (s *UserService) ListUsers(page, limit int, search, status string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := s.db.Model(&model.User{})

	// 搜索条件
	if search != "" {
		query = query.Where("email LIKE ? OR api_key LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// RegenerateAPIKey 重新生成用户的 API Key 和 Secret
func (s *UserService) RegenerateAPIKey(userID uint) (*model.User, string, error) {
	// 1. 获取用户
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, "", err
	}

	// 2. 生成新的凭证
	apiKey, apiSecret, err := s.generateAPICredentials()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API credentials: %w", err)
	}

	// 3. 更新用户
	user.APIKey = apiKey
	user.APISecret = apiSecret

	if err := s.db.Save(user).Error; err != nil {
		s.logger.Error("Failed to regenerate API key",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return nil, "", fmt.Errorf("failed to regenerate API key: %w", err)
	}

	s.logger.Info("API key regenerated",
		zap.Uint("user_id", userID),
	)

	return user, apiSecret, nil
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
func (s *UserService) generateAPICredentials() (string, string, error) {
	// 生成 API Key (32字节，base64编码)
	apiKeyBytes := make([]byte, 32)
	if _, err := rand.Read(apiKeyBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate API key: %w", err)
	}
	apiKey := base64.URLEncoding.EncodeToString(apiKeyBytes)

	// 生成 API Secret (48字节，base64编码)
	apiSecretBytes := make([]byte, 48)
	if _, err := rand.Read(apiSecretBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate API secret: %w", err)
	}
	apiSecret := base64.URLEncoding.EncodeToString(apiSecretBytes)

	return apiKey, apiSecret, nil
}

// DeleteUser 彻底删除用户及其所有相关数据
func (s *UserService) DeleteUser(userID uint) error {
	// 开始数据库事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除用户的交易记录
	if err := tx.Where("user_id = ?", userID).Delete(&model.Trade{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user trades: %w", err)
	}

	// 2. 删除用户的订单
	if err := tx.Where("user_id = ?", userID).Delete(&model.Order{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user orders: %w", err)
	}

	// 3. 删除用户的余额
	if err := tx.Where("user_id = ?", userID).Delete(&model.Balance{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user balances: %w", err)
	}

	// 4. 删除用户本身
	if err := tx.Delete(&model.User{}, userID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("User and all related data deleted successfully",
		zap.Uint("user_id", userID),
	)

	return nil
}

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	// 简单的邮箱正则验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
