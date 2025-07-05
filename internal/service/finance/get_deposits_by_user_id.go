package finance

import (
	"context"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error) {
	result, err := s.repo.GetDepositsByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return result, nil
}
