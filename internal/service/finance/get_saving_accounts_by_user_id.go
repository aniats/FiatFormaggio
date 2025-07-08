package finance

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
)

func (s *FinanceService) GetSavingAccountsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.SavingAccount, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		if UserID <= 0 {
			return nil, errors.ErrInValidUserID.WithContext("UserID", UserID)
		}

		accounts, err := s.repo.GetSavingAccountsByUserID(ctx, UserID)
		if err != nil {
			return nil, errors.WrapRepositoryError(err).WithContext("UserID", UserID)
		}

		return accounts, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.GetSavingAccountsByUserID")
	resultInterface, err := wrappedHandler(ctx, UserID)
	if err != nil {
		return nil, err
	}
	if resultInterface != nil {
		return resultInterface.([]domain.SavingAccount), nil
	}
	return nil, nil
}
