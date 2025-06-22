package finance

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func (s *FinanceService) EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "FinanceService.EnsureUserExists")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", int64(userID)),
		attribute.String("user.name", username),
	)

	return s.repo.EnsureUserExists(ctx, userID, username)
}
