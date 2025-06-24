package finance

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
)

func (s *FinanceService) GetCashHoldingsByUserID(ctx context.Context, userID domain.UserId) ([]domain.CashHolding, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if userID <= 0 {
			return nil, errors.ErrInvalidUserID.WithContext("userID", userID)
		}

		holdings, err := s.repo.GetCashHoldingsByUserID(ctx, userID)
		if err != nil {
			return nil, errors.WrapRepositoryError(err).WithContext("userID", userID)
		}

		return holdings, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.GetCashHoldingsByUserID")
	resultInterface, err := wrappedHandler(ctx, userID)
	if err != nil {
		return nil, err
	}
	if resultInterface != nil {
		return resultInterface.([]domain.CashHolding), nil
	}
	return nil, nil
}