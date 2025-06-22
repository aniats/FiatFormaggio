package finance

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func (s *FinanceService) GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "FinanceService.GetDepositsByUserID")
	defer span.End()

	span.SetAttributes(attribute.Int64("user.id", int64(userId)))

	result, err := s.repo.GetDepositsByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return result, nil
}
