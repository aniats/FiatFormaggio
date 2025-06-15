package finance

import (
	"context"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (fs *FinanceService) EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error) {
	return fs.repo.EnsureUserExists(ctx, userID, username)
}
