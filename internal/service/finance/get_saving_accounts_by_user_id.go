package finance

import (
	"context"
	"fmt"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) GetSavingAccountsByUserID(ctx context.Context, userID domain.UserId) ([]domain.SavingAccount, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	accounts, err := s.repo.GetSavingAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get saving accounts for user %d: %w", userID, err)
	}

	return accounts, nil
}