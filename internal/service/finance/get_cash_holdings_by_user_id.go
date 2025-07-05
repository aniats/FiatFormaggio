package finance

import (
	"context"
	"fmt"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) GetCashHoldingsByUserID(ctx context.Context, userID domain.UserId) ([]domain.CashHolding, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	holdings, err := s.repo.GetCashHoldingsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash holdings for user %d: %w", userID, err)
	}

	return holdings, nil
}