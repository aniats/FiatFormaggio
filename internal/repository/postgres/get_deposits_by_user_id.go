package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (repo *Repository) GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			SELECT
				id,
				user_id,
				name, 
				amount_minor_units, 
				interest_rate_basis_points, 
				expiration_date, 
				currency
			FROM deposits 
			WHERE user_id = $1
			ORDER BY created_at DESC`

		rows, queryErr := repo.db.QueryContext(ctx, query, userId)
		if queryErr != nil {
			return nil, fmt.Errorf("failed to query deposits for user %d: %w", userId, queryErr)
		}
		defer rows.Close()

		var deposits []domain.Deposit
		for rows.Next() {
			var deposit domain.Deposit
			var expirationDate sql.NullTime
			var interestRateBasisPoints sql.NullInt64

			scanErr := rows.Scan(
				&deposit.Id,
				&deposit.UserId,
				&deposit.Name,
				&deposit.AmountMinorUnits,
				&interestRateBasisPoints,
				&expirationDate,
				&deposit.Currency,
			)
			if scanErr != nil {
				return nil, fmt.Errorf("failed to scan deposit row: %w", scanErr)
			}

			if expirationDate.Valid {
				deposit.ExpirationDate = &expirationDate.Time
			}
			if interestRateBasisPoints.Valid {
				deposit.InterestRateBasisPoints = interestRateBasisPoints.Int64
			}

			deposits = append(deposits, deposit)
		}

		if rowsErr := rows.Err(); rowsErr != nil {
			return nil, fmt.Errorf("error iterating over deposit rows: %w", rowsErr)
		}

		return deposits, nil
	}

	wrappedHandler := repo.interceptor.Chain(handler, "Repository.GetDepositsByUserID")
	resultInterface, err := wrappedHandler(ctx, userId)
	if err != nil {
		return nil, err
	}

	if resultInterface != nil {
		return resultInterface.([]domain.Deposit), nil
	}
	return nil, nil
}
