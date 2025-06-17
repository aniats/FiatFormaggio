package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (repo *Repository) GetSavingAccountsByUserID(ctx context.Context, userId domain.UserId) ([]domain.SavingAccount, error) {
	query := `
        SELECT 
            id, 
            user_id, 
            name, 
            amount_minor_units, 
            interest_rate_basis_points, 
            expiration_date, 
            currency
        FROM saving_accounts 
        WHERE user_id = $1
        ORDER BY created_at DESC`

	rows, err := repo.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query saving accounts for user %d: %w", userId, err)
	}
	defer rows.Close()

	var savingAccounts []domain.SavingAccount
	for rows.Next() {
		var account domain.SavingAccount
		var expirationDate sql.NullTime

		err := rows.Scan(
			&account.Id,
			&account.UserId,
			&account.Name,
			&account.AmountMinorUnits,
			&account.InterestRateBasisPoints,
			&expirationDate,
			&account.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan saving account row: %w", err)
		}

		if expirationDate.Valid {
			account.ExpirationDate = &expirationDate.Time
		}
		savingAccounts = append(savingAccounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over saving account rows: %w", err)
	}

	return savingAccounts, nil
}
