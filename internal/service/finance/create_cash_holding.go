package finance

import (
	"context"
	"strings"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
)

func (s *FinanceService) CreateCashHolding(ctx context.Context, req *models.CreateCashHoldingRequest) (*domain.CashHolding, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if err := s.ValidateCreateCashHoldingRequest(req); err != nil {
			return nil, errors.WrapValidationError(err)
		}

		currency, err := s.ValidateAndNormalizeCurrency(req.Currency)
		if err != nil {
			return nil, errors.WrapValidationError(err)
		}

		amountMinorUnits := int64(req.AmountRUB * 100)

		cashHolding := &domain.CashHolding{
			UserID:           req.UserID,
			Name:             strings.TrimSpace(req.Name),
			AmountMinorUnits: amountMinorUnits,
			Currency:         currency,
		}

		if err = s.repo.CreateCashHolding(ctx, cashHolding); err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		return cashHolding, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.CreateCashHolding")
	resultInterface, err := wrappedHandler(ctx, req)
	if err != nil {
		return nil, err
	}

	if resultInterface != nil {
		return resultInterface.(*domain.CashHolding), nil
	}
	return nil, nil
}

func (s *FinanceService) ValidateCreateCashHoldingRequest(req *models.CreateCashHoldingRequest) error {
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

	return nil
}
