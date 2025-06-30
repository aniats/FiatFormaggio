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
	interceptor     *middleware.Interceptor
}

func NewFinanceService(repo repository.Repository, currencyService CurrencyService, appName string) *FinanceService {
	interceptor := middleware.NewInterceptor(middleware.DefaultConfig("FinanceService"), appName)
	
	return &FinanceService{
		repo:            repo,
		currencyService: currencyService,
		interceptor:     interceptor,
	}
}
