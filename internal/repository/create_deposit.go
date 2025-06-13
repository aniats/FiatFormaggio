package repository

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

// CreateDeposit - creates a new deposit for a user
func (repo *PostgresRepository) CreateDeposit(ctx context.Context, deposit *domain.Deposit) error {
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

	err := repo.db.QueryRowContext(
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
		return fmt.Errorf("failed to create deposit: %w", err)
	}

	return nil
}
