package postgres

import (
	"context"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
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
        RETURNING id, expiration_date
	`

	var id int64
	var expirationDate *time.Time

	err := repo.db.QueryRowContext(
		ctx,
		query,
		account.UserId,
		account.Name,
		account.AmountMinorUnits,
		account.InterestRateBasisPoints,
		account.ExpirationDate,
		account.Currency,
	).Scan(&id, &expirationDate)

	if err != nil {
		return errors.WrapRepositoryError(err)
	}

	// Update the account object with the returned values
	account.Id = id
	account.ExpirationDate = expirationDate

	return nil
}
