package finance

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return s.repo.EnsureUserExists(ctx, userID, username)
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.EnsureUserExists")
	resultInterface, err := wrappedHandler(ctx, map[string]interface{}{"userID": userID, "username": username})
	if err != nil {
		return false, err
	}
	if resultInterface != nil {
		return resultInterface.(bool), nil
	}
	return false, nil
}
