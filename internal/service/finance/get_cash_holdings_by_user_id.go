package finance

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
)

func (s *FinanceService) GetCashHoldingsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.CashHolding, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if UserID <= 0 {
			return nil, errors.ErrInValidUserID.WithContext("UserID", UserID)
		}

		holdings, err := s.repo.GetCashHoldingsByUserID(ctx, UserID)
		if err != nil {
			return nil, errors.WrapRepositoryError(err).WithContext("UserID", UserID)
		}

		return holdings, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.GetCashHoldingsByUserID")
	resultInterface, err := wrappedHandler(ctx, UserID)
	if err != nil {
		return nil, err
	}
	if resultInterface != nil {
		return resultInterface.([]domain.CashHolding), nil
	}
	return nil, nil
}
