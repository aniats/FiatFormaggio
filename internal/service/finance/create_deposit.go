package finance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
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

	var attrs []attribute.KeyValue
	if req != nil {
		attrs = append(attrs, attribute.Int64("user.id", int64(req.UserID)))
	}

	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if err = s.validateCreateDepositRequest(req); err != nil {
			return nil, errors.WrapValidationError(err)
		}

		currency, currErr := s.validateAndNormalizeCurrency(req.Currency)
		if currErr != nil {
			return nil, errors.WrapValidationError(currErr)
		}

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
			return nil, errors.WrapRepositoryError(repoErr)
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
		return "", errors.NewUnsupportedCurrencyError(currencyInput)
	}

	if !s.isCurrencyAllowedForDeposits(currency) {
		return "", errors.NewCurrencyNotAllowedForDepositsError(string(currency))
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
		return errors.ErrRequestNil
	}

	if req.UserID <= 0 {
		return errors.ErrInvalidUserID
	}

	if strings.TrimSpace(req.Name) == "" {
		return errors.ErrNameEmpty
	}

	if len(req.Name) > 255 {
		return errors.ErrNameTooLong
	}

	if req.AmountRUB <= 0 {
		return errors.ErrAmountNotPositive
	}

	if req.AmountRUB > 1000000000 {
		return errors.ErrAmountTooLarge
	}

	if req.AmountRUB < 1 {
		return errors.ErrAmountTooSmall
	}

	if req.InterestRatePercent != nil {
		if *req.InterestRatePercent < 0 || *req.InterestRatePercent > 100 {
			return errors.NewTechnicalValidationError("interest rate must be between 0 and 100 percent")
		}

		if *req.InterestRatePercent > 50 {
			return errors.NewTechnicalValidationError("interest rate seems too high (max 50% for safety)")
		}
	}

	if req.ExpirationDate != nil {
		if req.ExpirationDate.Before(time.Now()) {
			return errors.NewTechnicalValidationError("expiration date cannot be in the past")
		}

		maxDate := time.Now().AddDate(10, 0, 0)
		if req.ExpirationDate.After(maxDate) {
			return errors.NewTechnicalValidationError("expiration date cannot be more than 10 years in the future")
		}
	}

	return nil
}
