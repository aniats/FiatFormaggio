package finance

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	"strings"
	"time"
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

func (fs *FinanceService) CreateDeposit(ctx context.Context, req *models.CreateDepositRequest) (*domain.Deposit, error) {
	if err := fs.validateCreateDepositRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	currency, err := fs.validateAndNormalizeCurrency(req.Currency)
	if err != nil {
		return nil, fmt.Errorf("currency validation failed: %w", err)
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

	if err := fs.repo.CreateDeposit(ctx, deposit); err != nil {
		return nil, fmt.Errorf("failed to create deposit in repository: %w", err)
	}

	return deposit, nil
}

func (fs *FinanceService) validateAndNormalizeCurrency(currencyInput string) (domain.CurrencyName, error) {
	if currencyInput == "" {
		return domain.RUB, nil
	}

	currency, err := domain.CurrencyFromHuman(currencyInput)
	if err != nil {
		return "", fmt.Errorf("неподдерживаемая валюта '%s'. Поддерживаемые валюты: %s",
			currencyInput,
			fs.getSupportedCurrenciesString())
	}

	if !fs.isCurrencyAllowedForDeposits(currency) {
		return "", fmt.Errorf("валюта '%s' (%s) не поддерживается для депозитов",
			currency,
			currency.ToHumanRussian())
	}

	return currency, nil
}

func (fs *FinanceService) isCurrencyAllowedForDeposits(currency domain.CurrencyName) bool {
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

func (fs *FinanceService) getSupportedCurrenciesString() string {
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

func (fs *FinanceService) validateCreateDepositRequest(req *models.CreateDepositRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.UserID <= 0 {
		return fmt.Errorf("user ID must be positive")
	}

	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("deposit name cannot be empty")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("deposit name too long (max 255 characters)")
	}

	if req.AmountRUB <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if req.AmountRUB > 1000000000 {
		return fmt.Errorf("amount too large (max 1,000,000,000)")
	}

	if req.AmountRUB < 1 {
		return fmt.Errorf("minimum deposit amount is 1")
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
