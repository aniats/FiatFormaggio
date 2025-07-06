package finance

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error) {
	var (
		result []domain.Deposit
		err    error
	)

	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		deposits, repoErr := s.repo.GetDepositsByUserID(ctx, userId)
		return deposits, repoErr
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.GetDepositsByUserID")
	resultInterface, err := wrappedHandler(ctx, userId)
	if err != nil {
		return nil, err
	}

	if resultInterface != nil {
		result = resultInterface.([]domain.Deposit)
	}

	return result, err
}
