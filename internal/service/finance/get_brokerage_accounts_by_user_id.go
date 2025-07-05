package finance

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
)

func (s *FinanceService) GetBrokerageAccountsByUserID(ctx context.Context, userID domain.UserId) ([]domain.BrokerageAccount, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if userID <= 0 {
			return nil, errors.ErrInvalidUserID.WithContext("userID", userID)
		}

		accounts, err := s.repo.GetBrokerageAccountsByUserID(ctx, userID)
		if err != nil {
			return nil, errors.WrapRepositoryError(err).WithContext("userID", userID)
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