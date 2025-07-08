package finance

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (s *FinanceService) EnsureUserExists(ctx context.Context, UserID domain.UserID, username string) (bool, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return s.repo.EnsureUserExists(ctx, UserID, username)
	}

	wrappedHandler := s.interceptor.Chain(handler, "FinanceService.EnsureUserExists")
	resultInterface, err := wrappedHandler(ctx, map[string]interface{}{"UserID": UserID, "username": username})
	if err != nil {
		return false, err
	}
	if resultInterface != nil {
		return resultInterface.(bool), nil
	}
	return false, nil
}
