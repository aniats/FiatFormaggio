package repository

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"

	"time"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, account *domain.Account) error
	GetAccount(ctx context.Context, id int) (*domain.Account, error)
	ListAccounts(ctx context.Context, filter *AccountFilter) ([]*domain.Account, error)
	UpdateAccount(ctx context.Context, account *domain.Account) error
	DeleteAccount(ctx context.Context, id int) error

	AddTransaction(ctx context.Context, transaction *domain.Transaction) error
	GetTransactions(ctx context.Context, filter *TransactionFilter) ([]*domain.Transaction, error)

	AddProfit(ctx context.Context, profit *domain.Profit) error
	GetProfits(ctx context.Context, filter *ProfitFilter) ([]*domain.Profit, error)
}

type AccountFilter struct {
	Types []domain.AccountType
}

type TransactionFilter struct {
	AccountID int
	From      time.Time
	To        time.Time
}

type ProfitFilter struct {
	AccountID int
	From      time.Time
	To        time.Time
	GroupBy   string
}
