package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"runtime/pprof"
)

func (repo *Repository) GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error) {
	var deposits []domain.Deposit
	var err error

	pprof.Do(ctx, pprof.Labels("db_operation", "get_deposits_by_user_id"), func(ctx context.Context) {
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
			err = fmt.Errorf("failed to query deposits for user %d: %w", userId, queryErr)
			return
		}
		defer rows.Close()

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
				err = fmt.Errorf("failed to scan deposit row: %w", scanErr)
				return
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
			err = fmt.Errorf("error iterating over deposit rows: %w", rowsErr)
			return
		}
	})

	return deposits, err
}
