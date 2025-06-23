package finance

import (
	"context"
	"fmt"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) GetSavingAccountsByUserID(ctx context.Context, userID domain.UserId) ([]domain.SavingAccount, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if userID <= 0 {
			return nil, fmt.Errorf("invalid user ID: %d", userID)
		}

		accounts, err := s.repo.GetSavingAccountsByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get saving accounts for user %d: %w", userID, err)
		}

		return accounts, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.GetSavingAccountsByUserID")
	resultInterface, err := wrappedHandler(ctx, userID)
	if err != nil {
		return nil, err
	}
	if resultInterface != nil {
		return resultInterface.([]domain.SavingAccount), nil
	}
	return nil, nil
}