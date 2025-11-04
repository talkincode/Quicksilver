package service

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// BalanceService 余额管理服务
type BalanceService struct {
	db     *gorm.DB
	cfg    *config.Config
	logger *zap.Logger
}

// NewBalanceService 创建余额服务
func NewBalanceService(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *BalanceService {
	return &BalanceService{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}
}

// GetBalance 获取用户指定资产的余额
func (s *BalanceService) GetBalance(userID uint, asset string) (*model.Balance, error) {
	var balance model.Balance
	if err := s.db.Where("user_id = ? AND asset = ?", userID, asset).First(&balance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("balance not found for user %d and asset %s", userID, asset)
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &balance, nil
}

// GetAllBalances 获取用户所有资产余额
func (s *BalanceService) GetAllBalances(userID uint) ([]model.Balance, error) {
	var balances []model.Balance
	if err := s.db.Where("user_id = ?", userID).Find(&balances).Error; err != nil {
		return nil, fmt.Errorf("failed to get balances: %w", err)
	}

	return balances, nil
}

// FreezeBalance 冻结余额（从可用余额转到冻结余额）
func (s *BalanceService) FreezeBalance(userID uint, asset string, amount float64) error {
	// 1. 参数验证
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// 2. 使用事务确保原子性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取并锁定余额记录
		var balance model.Balance
		if err := tx.Where("user_id = ? AND asset = ?", userID, asset).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&balance).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("balance not found")
			}
			return fmt.Errorf("failed to lock balance: %w", err)
		}

		// 检查可用余额是否足够
		if balance.Available < amount {
			return fmt.Errorf("insufficient balance: available %.8f, required %.8f", balance.Available, amount)
		}

		// 更新余额
		balance.Available -= amount
		balance.Locked += amount

		if err := tx.Save(&balance).Error; err != nil {
			return fmt.Errorf("failed to freeze balance: %w", err)
		}

		s.logger.Info("Balance frozen",
			zap.Uint("user_id", userID),
			zap.String("asset", asset),
			zap.Float64("amount", amount),
		)

		return nil
	})
}

// UnfreezeBalance 解冻余额（从冻结余额转回可用余额）
func (s *BalanceService) UnfreezeBalance(userID uint, asset string, amount float64) error {
	// 1. 参数验证
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// 2. 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取并锁定余额记录
		var balance model.Balance
		if err := tx.Where("user_id = ? AND asset = ?", userID, asset).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&balance).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("balance not found")
			}
			return fmt.Errorf("failed to lock balance: %w", err)
		}

		// 检查冻结余额是否足够
		if balance.Locked < amount {
			return fmt.Errorf("insufficient locked balance: locked %.8f, required %.8f", balance.Locked, amount)
		}

		// 更新余额
		balance.Locked -= amount
		balance.Available += amount

		if err := tx.Save(&balance).Error; err != nil {
			return fmt.Errorf("failed to unfreeze balance: %w", err)
		}

		s.logger.Info("Balance unfrozen",
			zap.Uint("user_id", userID),
			zap.String("asset", asset),
			zap.Float64("amount", amount),
		)

		return nil
	})
}

// DeductBalance 从冻结余额中扣除（通常用于订单成交）
func (s *BalanceService) DeductBalance(userID uint, asset string, amount float64) error {
	// 1. 参数验证
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// 2. 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取并锁定余额记录
		var balance model.Balance
		if err := tx.Where("user_id = ? AND asset = ?", userID, asset).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&balance).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("balance not found")
			}
			return fmt.Errorf("failed to lock balance: %w", err)
		}

		// 检查冻结余额是否足够
		if balance.Locked < amount {
			return fmt.Errorf("insufficient locked balance: locked %.8f, required %.8f", balance.Locked, amount)
		}

		// 从冻结余额扣除
		balance.Locked -= amount

		if err := tx.Save(&balance).Error; err != nil {
			return fmt.Errorf("failed to deduct balance: %w", err)
		}

		s.logger.Info("Balance deducted",
			zap.Uint("user_id", userID),
			zap.String("asset", asset),
			zap.Float64("amount", amount),
		)

		return nil
	})
}

// AddBalance 增加可用余额（通常用于充值或订单成交收款）
func (s *BalanceService) AddBalance(userID uint, asset string, amount float64) error {
	// 1. 参数验证
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// 2. 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 尝试获取余额记录
		var balance model.Balance
		err := tx.Where("user_id = ? AND asset = ?", userID, asset).First(&balance).Error

		if err == gorm.ErrRecordNotFound {
			// 余额不存在，创建新记录
			balance = model.Balance{
				UserID:    userID,
				Asset:     asset,
				Available: amount,
				Locked:    0,
			}
			if err := tx.Create(&balance).Error; err != nil {
				return fmt.Errorf("failed to create balance: %w", err)
			}

			s.logger.Info("Balance created and added",
				zap.Uint("user_id", userID),
				zap.String("asset", asset),
				zap.Float64("amount", amount),
			)
		} else if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		} else {
			// 余额存在，增加金额
			balance.Available += amount
			if err := tx.Save(&balance).Error; err != nil {
				return fmt.Errorf("failed to add balance: %w", err)
			}

			s.logger.Info("Balance added",
				zap.Uint("user_id", userID),
				zap.String("asset", asset),
				zap.Float64("amount", amount),
			)
		}

		return nil
	})
}

// TransferBalance 在两个用户之间转账
func (s *BalanceService) TransferBalance(fromUserID, toUserID uint, asset string, amount float64) error {
	// 1. 参数验证
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if fromUserID == toUserID {
		return fmt.Errorf("cannot transfer to yourself")
	}

	// 2. 使用事务确保原子性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 锁定并获取发送方余额
		var fromBalance model.Balance
		if err := tx.Where("user_id = ? AND asset = ?", fromUserID, asset).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&fromBalance).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("sender balance not found")
			}
			return fmt.Errorf("failed to lock sender balance: %w", err)
		}

		// 检查发送方余额是否足够
		if fromBalance.Available < amount {
			return fmt.Errorf("insufficient balance: available %.8f, required %.8f", fromBalance.Available, amount)
		}

		// 扣除发送方余额
		fromBalance.Available -= amount
		if err := tx.Save(&fromBalance).Error; err != nil {
			return fmt.Errorf("failed to deduct sender balance: %w", err)
		}

		// 增加接收方余额
		var toBalance model.Balance
		err := tx.Where("user_id = ? AND asset = ?", toUserID, asset).First(&toBalance).Error

		if err == gorm.ErrRecordNotFound {
			// 接收方余额不存在，创建新记录
			toBalance = model.Balance{
				UserID:    toUserID,
				Asset:     asset,
				Available: amount,
				Locked:    0,
			}
			if err := tx.Create(&toBalance).Error; err != nil {
				return fmt.Errorf("failed to create receiver balance: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to get receiver balance: %w", err)
		} else {
			// 接收方余额存在，增加金额
			toBalance.Available += amount
			if err := tx.Save(&toBalance).Error; err != nil {
				return fmt.Errorf("failed to add receiver balance: %w", err)
			}
		}

		s.logger.Info("Balance transferred",
			zap.Uint("from_user_id", fromUserID),
			zap.Uint("to_user_id", toUserID),
			zap.String("asset", asset),
			zap.Float64("amount", amount),
		)

		return nil
	})
}
