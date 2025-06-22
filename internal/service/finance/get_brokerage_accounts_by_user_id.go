package finance

import (
	"context"
	"fmt"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func (s *FinanceService) GetBrokerageAccountsByUserID(ctx context.Context, userID domain.UserId) ([]domain.BrokerageAccount, error) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "FinanceService.GetBrokerageAccountsByUserID")
	defer span.End()

	span.SetAttributes(attribute.Int64("user.id", int64(userID)))

	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	accounts, err := s.repo.GetBrokerageAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get brokerage accounts for user %d: %w", userID, err)
	}

	return accounts, nil
}