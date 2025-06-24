package postgres

import (
	"context"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
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
			RETURNING id, expiration_date, created_at, updated_at
		`

		var id int64
		var expirationDate *time.Time
		var createdAt, updatedAt time.Time

		err := repo.db.QueryRowContext(
			ctx,
			query,
			deposit.UserId,
			deposit.Name,
			deposit.AmountMinorUnits,
			deposit.InterestRateBasisPoints,
			deposit.ExpirationDate,
			deposit.Currency,
		).Scan(&id, &expirationDate, &createdAt, &updatedAt)

		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		// Update the deposit object with the returned values
		deposit.Id = id
		deposit.ExpirationDate = expirationDate

		return nil, nil
	}

	wrappedHandler := repo.interceptor.Chain(handler, "Repository.CreateDeposit")
	_, err := wrappedHandler(ctx, deposit)
	return err
}