package finance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	"go.opentelemetry.io/otel/attribute"
)

type DepositCreationStep int

const (
	StepName DepositCreationStep = iota
	StepAmount
	StepCurrency
	StepInterestRate
	StepExpirationDate
	StepConfirmation
)

func (s *FinanceService) CreateDeposit(ctx context.Context, req *models.CreateDepositRequest) (*domain.Deposit, error) {
	var result *domain.Deposit
	var err error
	
	// Use the universal wrapper for clean tracing
	attrs := []attribute.KeyValue{}
	if req != nil {
		attrs = append(attrs, attribute.Int64("user.id", int64(req.UserID)))
	}
	
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if err = s.validateCreateDepositRequest(req); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		currency, currErr := s.validateAndNormalizeCurrency(req.Currency)
		if currErr != nil {
			return nil, fmt.Errorf("currency validation failed: %w", currErr)
		}

		// TODO: go to db to calc
		amountMinorUnits := int64(req.AmountRUB * 100)

		var interestRateBasisPoints int64
		if req.InterestRatePercent != nil {
			basisPoints := int64(*req.InterestRatePercent * 100)
			interestRateBasisPoints = basisPoints
		}

		deposit := &domain.Deposit{
			UserId:                  req.UserID,
			Name:                    req.Name,
			AmountMinorUnits:        amountMinorUnits,
			InterestRateBasisPoints: interestRateBasisPoints,
			ExpirationDate:          req.ExpirationDate,
			Currency:                currency,
		}

		if repoErr := s.repo.CreateDeposit(ctx, deposit); repoErr != nil {
			return nil, fmt.Errorf("failed to create deposit in repository: %w", repoErr)
		}

		return deposit, nil
	}
	
	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.CreateDeposit")
	resultInterface, err := wrappedHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	
	if resultInterface != nil {
		result = resultInterface.(*domain.Deposit)
	}
	
	return result, err
}

func (s *FinanceService) validateAndNormalizeCurrency(currencyInput string) (domain.CurrencyName, error) {
	if currencyInput == "" {
		return domain.RUB, nil
	}

	currency, err := domain.CurrencyFromHuman(currencyInput)
	if err != nil {
		return "", fmt.Errorf("неподдерживаемая валюта '%s'. Поддерживаемые валюты: %s",
			currencyInput,
			s.getSupportedCurrenciesString())
	}

	if !s.isCurrencyAllowedForDeposits(currency) {
		return "", fmt.Errorf("валюта '%s' (%s) не поддерживается для депозитов",
			currency,
			currency.ToHumanRussian())
	}

	return currency, nil
}

func (s *FinanceService) isCurrencyAllowedForDeposits(currency domain.CurrencyName) bool {
	allowedCurrencies := map[domain.CurrencyName]bool{
		domain.RUB: true,
		domain.USD: true,
		domain.EUR: true,
		domain.CNY: true,
		domain.GBP: true,
		// domain.XBT: false,
		// domain.KZT: false,
	}

	return allowedCurrencies[currency]
}

func (s *FinanceService) getSupportedCurrenciesString() string {
	allowedCurrencies := []domain.CurrencyName{
		domain.RUB, domain.USD, domain.EUR, domain.CNY, domain.GBP,
	}

	var currencies []string
	for _, currency := range allowedCurrencies {
		currencies = append(currencies, fmt.Sprintf("%s (%s)",
			currency,
			currency.ToHumanRussian()))
	}

	return strings.Join(currencies, ", ")
}

func (s *FinanceService) validateCreateDepositRequest(req *models.CreateDepositRequest) error {
	if req == nil {
		return fmt.Errorf("запрос не может быть пустым")
	}

	if req.UserID <= 0 {
		return fmt.Errorf("ID пользователя должен быть положительным")
	}

	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("название депозита не может быть пустым")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("название депозита слишком длинное (максимум 255 символов)")
	}

	if req.AmountRUB <= 0 {
		return fmt.Errorf("сумма должна быть положительной")
	}

	if req.AmountRUB > 1000000000 {
		return fmt.Errorf("сумма слишком большая (максимум 1,000,000,000)")
	}

	if req.AmountRUB < 1 {
		return fmt.Errorf("минимальная сумма депозита: 1")
	}

	if req.InterestRatePercent != nil {
		if *req.InterestRatePercent < 0 || *req.InterestRatePercent > 100 {
			return fmt.Errorf("interest rate must be between 0 and 100 percent")
		}

		if *req.InterestRatePercent > 50 {
			return fmt.Errorf("interest rate seems too high (max 50%% for safety)")
		}
	}

	if req.ExpirationDate != nil {
		if req.ExpirationDate.Before(time.Now()) {
			return fmt.Errorf("expiration date cannot be in the past")
		}

		maxDate := time.Now().AddDate(10, 0, 0)
		if req.ExpirationDate.After(maxDate) {
			return fmt.Errorf("expiration date cannot be more than 10 years in the future")
		}
	}

	return nil
}
