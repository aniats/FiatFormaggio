package repository

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

// GetCashHoldingsByUserID - retrieves all cash holdings for a specific user
func (repo *PostgresRepository) GetCashHoldingsByUserID(ctx context.Context, userId domain.UserId) ([]domain.CashHolding, error) {
	query := `
        SELECT 
            id, 
            user_id, 
            name, 
            amount_minor_units, 
            currency
        FROM cash_holdings 
        WHERE user_id = $1
        ORDER BY created_at DESC
	`

	rows, err := repo.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query cash holdings for user %d: %w", userId, err)
	}
	defer rows.Close()

	var cashHoldings []domain.CashHolding
	for rows.Next() {
		var cash domain.CashHolding

		err := rows.Scan(
			&cash.Id,
			&cash.UserId,
			&cash.Name,
			&cash.AmountMinorUnits,
			&cash.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cash holding row: %w", err)
		}

		cashHoldings = append(cashHoldings, cash)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over cash holding rows: %w", err)
	}

	return cashHoldings, nil
}
