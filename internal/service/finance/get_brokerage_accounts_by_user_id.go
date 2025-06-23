package finance

import (
	"context"
	"fmt"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) GetBrokerageAccountsByUserID(ctx context.Context, userID domain.UserId) ([]domain.BrokerageAccount, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if userID <= 0 {
			return nil, fmt.Errorf("invalid user ID: %d", userID)
		}

		accounts, err := s.repo.GetBrokerageAccountsByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get brokerage accounts for user %d: %w", userID, err)
		}

		return accounts, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.GetBrokerageAccountsByUserID")
	resultInterface, err := wrappedHandler(ctx, userID)
	if err != nil {
		return nil, err
	}
	if resultInterface != nil {
		return resultInterface.([]domain.BrokerageAccount), nil
	}
	return nil, nil
}