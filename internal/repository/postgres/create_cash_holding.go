package postgres

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (repo *Repository) CreateCashHolding(ctx context.Context, cash *domain.CashHolding) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			INSERT INTO cash_holdings (
				user_id, 
				name, 
				amount_minor_units, 
				currency
			) VALUES ($1, $2, $3, $4)`

		_, err := repo.db.ExecContext(
			ctx,
			query,
			cash.UserId,
			cash.Name,
			cash.AmountMinorUnits,
			cash.Currency,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create cash holding: %w", err)
		}

		return nil, nil
	}

	wrappedHandler := repo.interceptor.Chain(handler, "Repository.CreateCashHolding")
	_, err := wrappedHandler(ctx, cash)
	return err
}
