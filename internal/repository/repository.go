package repository

import (
	"context"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error)
}

type DepositRepository interface {
	CreateDeposit(ctx context.Context, deposit *domain.Deposit) error
	GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error)
}

type CashHoldingRepository interface {
	CreateCashHolding(ctx context.Context, cash *domain.CashHolding) error
	GetCashHoldingsByUserID(ctx context.Context, userId domain.UserId) ([]domain.CashHolding, error)
}

type BrokerageAccountRepository interface {
	CreateBrokerageAccount(ctx context.Context, account *domain.BrokerageAccount) error
	GetBrokerageAccountsByUserID(ctx context.Context, userId domain.UserId) ([]domain.BrokerageAccount, error)
}

type SavingAccountRepository interface {
	CreateSavingAccount(ctx context.Context, account *domain.SavingAccount) error
	GetSavingAccountsByUserID(ctx context.Context, userId domain.UserId) ([]domain.SavingAccount, error)
}

type CurrencyRateRepository interface {
	UpsertCurrencyRate(ctx context.Context, rate *domain.CurrencyRate) error
	GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error)
	GetLastUpdateTime(ctx context.Context) (*time.Time, error)
}

type Repository interface {
	UserRepository
	DepositRepository
	CashHoldingRepository
	BrokerageAccountRepository
	SavingAccountRepository
	CurrencyRateRepository

	Close() error
	HealthCheck(ctx context.Context) error
}
