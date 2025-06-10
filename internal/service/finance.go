package service

import (
	"context"

	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type FinanceService struct {
	repo repository.AccountRepository
	cbr  *CBRService
}

func NewFinanceService(repo repository.AccountRepository, cbr *CBRService) *FinanceService {
	return &FinanceService{
		repo: repo,
		cbr:  cbr,
	}
}

func (s *FinanceService) GetTotalBalance(ctx context.Context) (float64, error) {
	accounts, err := s.repo.ListAccounts(ctx, &repository.AccountFilter{})
	if err != nil {
		return 0, err
	}

	rates, err := s.cbr.GetCurrencyRates(ctx, time.Now())
	if err != nil {
		return 0, err
	}

	var total float64
	for _, acc := range accounts {
		if acc.Currency == "RUB" {
			total += acc.Balance
			continue
		}

		for _, rate := range rates {
			if rate.CharCode == acc.Currency {
				total += acc.Balance * (rate.Value / float64(rate.Nominal))
				break
			}
		}
	}

	return total, nil
}

func (s *FinanceService) GetBalanceWithoutBrokerage(ctx context.Context) (float64, error) {
	_, err := s.repo.ListAccounts(ctx, &repository.AccountFilter{
		Types: []domain.AccountType{domain.Deposit, domain.Savings, domain.CashCurrency},
	})
	if err != nil {
		return 0, err
	}

	// TODO:
	return 0.0, nil
}

func (s *FinanceService) GetBalanceWithoutCurrency(ctx context.Context) (float64, error) {
	_, err := s.repo.ListAccounts(ctx, &repository.AccountFilter{
		Types: []domain.AccountType{domain.Deposit, domain.Savings, domain.Brokerage},
	})
	if err != nil {
		return 0, err
	}
	// TODO:
	return 0.0, nil
}

func (s *FinanceService) GetAccountsByType(ctx context.Context, accType domain.AccountType) ([]*domain.Account, error) {
	return s.repo.ListAccounts(ctx, &repository.AccountFilter{
		Types: []domain.AccountType{accType},
	})
}

func (s *FinanceService) GetSumByAccountType(ctx context.Context, accType domain.AccountType) (float64, error) {
	_, err := s.GetAccountsByType(ctx, accType)
	if err != nil {
		return 0, err
	}

	// TODO:
	return 0.0, nil
}

func (s *FinanceService) GetProfit(ctx context.Context, filter *repository.ProfitFilter) ([]*domain.Profit, error) {
	return s.repo.GetProfits(ctx, filter)
}
