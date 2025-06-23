package finance

import (
	"context"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/middleware"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type CurrencyService interface {
	GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error)
}

type FinanceService struct {
	repo            repository.Repository
	currencyService CurrencyService
	interceptor     *middleware.UnifiedInterceptor
}

func NewFinanceService(repo repository.Repository, currencyService CurrencyService) *FinanceService {
	interceptor := middleware.NewUnifiedInterceptor(middleware.DefaultConfig("FinanceService"))
	
	return &FinanceService{
		repo:            repo,
		currencyService: currencyService,
		interceptor:     interceptor,
	}
}
