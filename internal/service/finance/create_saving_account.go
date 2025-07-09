package finance

import (
	"context"
	"strings"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
)

func (s *FinanceService) CreateSavingAccount(ctx context.Context, req *models.CreateSavingAccountRequest) (*domain.SavingAccount, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if err := s.ValidateCreateSavingAccountRequest(req); err != nil {
			return nil, errors.WrapValidationError(err)
		}

		currency, err := s.ValidateAndNormalizeCurrency(req.Currency)
		if err != nil {
			return nil, errors.WrapValidationError(err)
		}

		amountMinorUnits := int64(req.AmountRUB * 100)

		var interestRateBasisPoints int64
		if req.InterestRatePercent != nil {
			interestRateBasisPoints = int64(*req.InterestRatePercent * 100)
		}

		account := &domain.SavingAccount{
			UserID:                  req.UserID,
			Name:                    strings.TrimSpace(req.Name),
			AmountMinorUnits:        amountMinorUnits,
			Currency:                currency,
			InterestRateBasisPoints: interestRateBasisPoints,
			ExpirationDate:          req.ExpirationDate,
		}

		if err := s.repo.CreateSavingAccount(ctx, account); err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		return account, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.CreateSavingAccount")
	resultInterface, err := wrappedHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	if resultInterface != nil {
		return resultInterface.(*domain.SavingAccount), nil
	}
	return nil, nil
}

func (s *FinanceService) ValidateCreateSavingAccountRequest(req *models.CreateSavingAccountRequest) error {
	if req == nil {
		return errors.ErrRequestNil
	}

	if req.UserID <= 0 {
		return errors.ErrInValidUserID
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errors.ErrNameRequired
	}

	if len(name) > 255 {
		return errors.ErrNameTooLong
	}

	if req.AmountRUB < 0 {
		return errors.ErrAmountNegative
	}

	if req.AmountRUB > 1000000000 {
		return errors.ErrAmountTooLarge
	}

	if req.Currency == "" {
		return errors.ErrCurrencyRequired
	}

	if req.InterestRatePercent != nil {
		if *req.InterestRatePercent < 0 || *req.InterestRatePercent > 100 {
			return errors.NewTechnicalValidationError("interest rate must be between 0 and 100 percent")
		}

		if *req.InterestRatePercent > 50 {
			return errors.NewTechnicalValidationError("interest rate seems too high (max 50% for safety)")
		}
	}

	return nil
}
