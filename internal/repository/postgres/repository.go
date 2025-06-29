package postgres

import (
	"context"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type Repository struct {
	db *Database

	User             *UserRepository
	BrokerageAccount *BrokerageAccountRepository
	SavingAccount    *SavingAccountRepository
	Deposit          *DepositRepository
	Cash             *CashRepository
	Currency         *CurrencyRepository
}

func New(connString string) (*Repository, error) {
	db, err := NewDatabase(connString)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:               db,
		User:             NewUserRepository(db),
		BrokerageAccount: NewBrokerageAccountRepository(db),
		SavingAccount:    NewSavingAccountRepository(db),
		Deposit:          NewDepositRepository(db),
		Cash:             NewCashRepository(db),
		Currency:         NewCurrencyRepository(db),
	}, nil
}

func (repo *Repository) Close() error {
	return repo.db.Close()
}

func (repo *Repository) HealthCheck(ctx context.Context) error {
	return repo.db.HealthCheck(ctx)
}

func (repo *Repository) EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error) {
	return repo.User.EnsureUserExists(ctx, userID, username)
}

func (repo *Repository) CreateBrokerageAccount(ctx context.Context, account *domain.BrokerageAccount) error {
	return repo.BrokerageAccount.CreateBrokerageAccount(ctx, account)
}

func (repo *Repository) GetBrokerageAccountsByUserID(ctx context.Context, userId domain.UserId) ([]domain.BrokerageAccount, error) {
	return repo.BrokerageAccount.GetBrokerageAccountsByUserID(ctx, userId)
}

func (repo *Repository) CreateSavingAccount(ctx context.Context, account *domain.SavingAccount) error {
	return repo.SavingAccount.CreateSavingAccount(ctx, account)
}

func (repo *Repository) GetSavingAccountsByUserID(ctx context.Context, userId domain.UserId) ([]domain.SavingAccount, error) {
	return repo.SavingAccount.GetSavingAccountsByUserID(ctx, userId)
}

func (repo *Repository) CreateDeposit(ctx context.Context, deposit *domain.Deposit) error {
	return repo.Deposit.CreateDeposit(ctx, deposit)
}

func (repo *Repository) GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error) {
	return repo.Deposit.GetDepositsByUserID(ctx, userId)
}

func (repo *Repository) CreateCashHolding(ctx context.Context, cash *domain.CashHolding) error {
	return repo.Cash.CreateCashHolding(ctx, cash)
}

func (repo *Repository) GetCashHoldingsByUserID(ctx context.Context, userId domain.UserId) ([]domain.CashHolding, error) {
	return repo.Cash.GetCashHoldingsByUserID(ctx, userId)
}

func (repo *Repository) UpsertCurrencyRate(ctx context.Context, rate *domain.CurrencyRate) error {
	return repo.Currency.UpsertCurrencyRate(ctx, rate)
}

func (repo *Repository) GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error) {
	return repo.Currency.GetCurrencyRates(ctx)
}

func (repo *Repository) GetLastUpdateTime(ctx context.Context) (*time.Time, error) {
	return repo.Currency.GetLastUpdateTime(ctx)
}

var _ repository.Repository = (*Repository)(nil)
