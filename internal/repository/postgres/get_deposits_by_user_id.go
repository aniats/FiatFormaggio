package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (repo *Repository) GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error) {
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

	rows, err := repo.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query deposits for user %d: %w", userId, err)
	}
	defer rows.Close()

	var deposits []domain.Deposit
	for rows.Next() {
		var deposit domain.Deposit
		var expirationDate sql.NullTime
		var interestRateBasisPoints sql.NullInt64

		err := rows.Scan(
			&deposit.Id,
			&deposit.UserId,
			&deposit.Name,
			&deposit.AmountMinorUnits,
			&interestRateBasisPoints,
			&expirationDate,
			&deposit.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deposit row: %w", err)
		}

		if expirationDate.Valid {
			deposit.ExpirationDate = &expirationDate.Time
		}
		if interestRateBasisPoints.Valid {
			deposit.InterestRateBasisPoints = interestRateBasisPoints.Int64
		}

		deposits = append(deposits, deposit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over deposit rows: %w", err)
	}

	return deposits, nil
}
