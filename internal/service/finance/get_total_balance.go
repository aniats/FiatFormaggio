package finance

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/tracing"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

const DefaultMinorUnits = 100.0

func (s *FinanceService) GetTotalBalance(ctx context.Context, userID domain.UserID) (float64, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		params := input.(map[string]interface{})
		userID := params[tracing.UserID].(domain.UserID)
		return s.processGetTotalBalance(ctx, userID)
	}

	params := map[string]interface{}{
		tracing.UserID: userID,
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.GetTotalBalance")
	result, err := wrappedHandler(ctx, params)
	if err != nil {
		return 0, err
	}

	if totalBalance, ok := result.(float64); ok {
		return totalBalance, nil
	}

	return 0, fmt.Errorf("unexpected result type from GetTotalBalance")
}

func (s *FinanceService) processGetTotalBalance(ctx context.Context, UserID domain.UserID) (float64, error) {
	rates, err := s.currencyService.GetCurrencyRates(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get currency rates: %w", err)
	}

	exchangeRates := buildExchangeRateMap(rates)
	var totalRUB float64

	deposits, err := s.GetDepositsByUserID(ctx, UserID)
	if err == nil {
		for _, deposit := range deposits {
			amount := float64(deposit.AmountMinorUnits) / DefaultMinorUnits
			amountRUB := convertToRUB(amount, deposit.Currency, exchangeRates)
			totalRUB += amountRUB
		}
	}

	brokerageAccounts, err := s.GetBrokerageAccountsByUserID(ctx, UserID)
	if err == nil {
		for _, account := range brokerageAccounts {
			amount := float64(account.AmountMinorUnits) / DefaultMinorUnits
			amountRUB := convertToRUB(amount, account.Currency, exchangeRates)
			totalRUB += amountRUB
		}
	}

	savingAccounts, err := s.GetSavingAccountsByUserID(ctx, UserID)
	if err == nil {
		for _, account := range savingAccounts {
			amount := float64(account.AmountMinorUnits) / DefaultMinorUnits
			amountRUB := convertToRUB(amount, account.Currency, exchangeRates)
			totalRUB += amountRUB
		}
	}

	cashHoldings, err := s.GetCashHoldingsByUserID(ctx, UserID)
	if err == nil {
		for _, holding := range cashHoldings {
			amount := float64(holding.AmountMinorUnits) / DefaultMinorUnits
			amountRUB := convertToRUB(amount, holding.Currency, exchangeRates)
			totalRUB += amountRUB
		}
	}

	return totalRUB, nil
}

func buildExchangeRateMap(rates []domain.CurrencyRate) map[domain.CurrencyName]float64 {
	exchangeRates := map[domain.CurrencyName]float64{
		domain.RUB: 1.0,
	}

	for _, rate := range rates {
		if rate.BaseCurrency == domain.RUB {
			exchangeRates[rate.Currency] = float64(rate.RateMinorUnits) / DefaultMinorUnits
		}
	}

	return exchangeRates
}

func convertToRUB(amount float64, currency domain.CurrencyName, exchangeRates map[domain.CurrencyName]float64) float64 {
	if currency == domain.RUB {
		return amount
	}
	if rate, exists := exchangeRates[currency]; exists {
		return amount * rate
	}
	return amount
}
