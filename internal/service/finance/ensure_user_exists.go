package finance

import (
	"context"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error) {
	return s.repo.EnsureUserExists(ctx, userID, username)
}
