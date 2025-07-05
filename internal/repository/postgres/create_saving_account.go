package postgres

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (repo *Repository) CreateSavingAccount(ctx context.Context, account *domain.SavingAccount) error {
	query := `
        INSERT INTO saving_accounts (
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
		account.UserId,
		account.Name,
		account.AmountMinorUnits,
		account.InterestRateBasisPoints,
		account.ExpirationDate,
		account.Currency,
	)

	if err != nil {
		return fmt.Errorf("failed to create saving account: %w", err)
	}

	return nil
}
