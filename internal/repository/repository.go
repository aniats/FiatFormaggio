package repository

import (
	"context"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

// UserRepository handles user-related operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
}

// DepositRepository handles deposit-related operations
type DepositRepository interface {
	CreateDeposit(ctx context.Context, deposit *domain.Deposit) error
	GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error)
}

// CashHoldingRepository handles cash holding operations
type CashHoldingRepository interface {
	CreateCashHolding(ctx context.Context, cash *domain.CashHolding) error
	GetCashHoldingsByUserID(ctx context.Context, userId domain.UserId) ([]domain.CashHolding, error)
}

// BrokerageAccountRepository handles brokerage account operations
type BrokerageAccountRepository interface {
	CreateBrokerageAccount(ctx context.Context, account *domain.BrokerageAccount) error
	GetBrokerageAccountsByUserID(ctx context.Context, userId domain.UserId) ([]domain.BrokerageAccount, error)
}

// SavingAccountRepository handles saving account operations
type SavingAccountRepository interface {
	CreateSavingAccount(ctx context.Context, account *domain.SavingAccount) error
	GetSavingAccountsByUserID(ctx context.Context, userId domain.UserId) ([]domain.SavingAccount, error)
}

// Repository is the main interface that composes all sub-repositories
type Repository interface {
	UserRepository
	DepositRepository
	CashHoldingRepository
	BrokerageAccountRepository
	SavingAccountRepository

	// Database management
	Close() error
	HealthCheck(ctx context.Context) error
}
