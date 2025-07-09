package finance

import (
	"context"
	"fmt"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/middleware"
	"github.com/aniats/FiatFormaggio/internal/repository"
	"github.com/aniats/FiatFormaggio/internal/service/chatgpt"
)

type CurrencyService interface {
	GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error)
}

type FinanceService struct {
	repo            repository.Repository
	currencyService CurrencyService
	chatgptClient   *chatgpt.Client
	interceptor     *middleware.Interceptor
}

func NewFinanceService(repo repository.Repository, currencyService CurrencyService, appName string, chatgptClient *chatgpt.Client) *FinanceService {
	interceptor := middleware.NewInterceptor(middleware.DefaultConfig("FinanceService"), appName)

	return &FinanceService{
		repo:            repo,
		currencyService: currencyService,
		chatgptClient:   chatgptClient,
		interceptor:     interceptor,
	}
}

type FinanceDataProvider struct {
	service *FinanceService
}

func (f *FinanceDataProvider) GetDepositsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.Deposit, error) {
	return f.service.GetDepositsByUserID(ctx, UserID)
}

func (f *FinanceDataProvider) GetBrokerageAccountsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.BrokerageAccount, error) {
	return f.service.GetBrokerageAccountsByUserID(ctx, UserID)
}

func (f *FinanceDataProvider) GetSavingAccountsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.SavingAccount, error) {
	return f.service.GetSavingAccountsByUserID(ctx, UserID)
}

func (f *FinanceDataProvider) GetCashHoldingsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.CashHolding, error) {
	return f.service.GetCashHoldingsByUserID(ctx, UserID)
}

func (f *FinanceDataProvider) GetTotalBalance(ctx context.Context, UserID domain.UserID) (float64, error) {
	return f.service.GetTotalBalance(ctx, UserID)
}

func (s *FinanceService) GetFinancialAdvice(ctx context.Context, userID domain.UserID, userMessage string, previousMessages []chatgpt.ChatMessage) (string, error) {
	if s.chatgptClient == nil {
		return "", fmt.Errorf("financial advice is not available - ChatGPT client not initialized")
	}

	dataProvider := &FinanceDataProvider{service: s}

	return s.chatgptClient.GetFinancialAdviceWithContext(ctx, userID, userMessage, previousMessages, dataProvider, s.currencyService)
}
