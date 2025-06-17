package finance

import (
	"context"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

// CurrencyService interface for getting currency rates
type CurrencyService interface {
	GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error)
}

type FinanceService struct {
	repo            repository.Repository
	currencyService CurrencyService
}

func NewFinanceService(repo repository.Repository, currencyService CurrencyService) *FinanceService {
	return &FinanceService{
		repo:            repo,
		currencyService: currencyService,
	}
}
