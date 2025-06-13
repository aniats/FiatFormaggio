package repository

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

// CreateCashHolding - creates a new cash holding for a user
func (repo *PostgresRepository) CreateCashHolding(ctx context.Context, cash *domain.CashHolding) error {
	query := `
        INSERT INTO cash_holdings (
            user_id, 
            name, 
            amount_minor_units, 
            currency
        ) VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at`

	err := repo.db.QueryRowContext(
		ctx,
		query,
		cash.UserId,
		cash.Name,
		cash.AmountMinorUnits,
		cash.Currency,
	)

	if err != nil {
		return fmt.Errorf("failed to create cash holding: %w", err)
	}

	return nil
}
