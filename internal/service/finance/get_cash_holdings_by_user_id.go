package finance

import (
	"context"
	"fmt"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) GetCashHoldingsByUserID(ctx context.Context, userID domain.UserId) ([]domain.CashHolding, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if userID <= 0 {
			return nil, fmt.Errorf("invalid user ID: %d", userID)
		}

		holdings, err := s.repo.GetCashHoldingsByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cash holdings for user %d: %w", userID, err)
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