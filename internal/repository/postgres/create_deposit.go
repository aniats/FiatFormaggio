package postgres

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (repo *Repository) CreateDeposit(ctx context.Context, deposit *domain.Deposit) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			INSERT INTO deposits (
				user_id, 
				name, 
				amount_minor_units, 
				interest_rate_basis_points, 
				expiration_date, 
				currency
			) VALUES ($1, $2, $3, $4, $5, $6)
		`

		_, err := repo.db.ExecContext(
			ctx,
			query,
			deposit.UserId,
			deposit.Name,
			deposit.AmountMinorUnits,
			deposit.InterestRateBasisPoints,
			deposit.ExpirationDate,
			deposit.Currency,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create deposit: %w", err)
		}

		return nil, nil
	}

	wrappedHandler := repo.interceptor.Chain(handler, "Repository.CreateDeposit")
	_, err := wrappedHandler(ctx, deposit)
	return err
}