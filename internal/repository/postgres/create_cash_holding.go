package postgres

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
)

func (repo *Repository) CreateCashHolding(ctx context.Context, cash *domain.CashHolding) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			INSERT INTO cash_holdings (
				user_id, 
				name, 
				amount_minor_units, 
				currency
			) VALUES ($1, $2, $3, $4)
			RETURNING id`

		var id int64

		err := repo.db.QueryRowContext(
			ctx,
			query,
			cash.UserId,
			cash.Name,
			cash.AmountMinorUnits,
			cash.Currency,
		).Scan(&id)

		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		// Update the cash holding object with the returned ID
		cash.Id = id

		return nil, nil
	}

	wrappedHandler := repo.interceptor.Chain(handler, "Repository.CreateCashHolding")
	_, err := wrappedHandler(ctx, cash)
	return err
}
