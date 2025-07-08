package postgres

import (
	"context"
	"database/sql"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type SavingAccountRepository struct {
	*Database
}

func NewSavingAccountRepository(db *Database) *SavingAccountRepository {
	return &SavingAccountRepository{Database: db}
}

func (r *SavingAccountRepository) CreateSavingAccount(ctx context.Context, account *domain.SavingAccount) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			INSERT INTO saving_accounts (
				user_id, 
				name, 
				amount_minor_units, 
				interest_rate_basis_points, 
				expiration_date, 
				currency
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING ID`

		var ID int64

		err := r.DB.QueryRowContext(
			ctx,
			query,
			account.UserID,
			account.Name,
			account.AmountMinorUnits,
			account.InterestRateBasisPoints,
			account.ExpirationDate,
			account.Currency,
		).Scan(&ID)

		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		account.ID = ID
		return nil, nil
	}

	_, err := r.ExecuteWithInterceptor(ctx, "Repository.CreateSavingAccount", handler)
	return err
}

func (r *SavingAccountRepository) GetSavingAccountsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.SavingAccount, error) {
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

	rows, err := r.DB.QueryContext(ctx, query, UserID)
	if err != nil {
		return nil, errors.WrapRepositoryError(err)
	}
	defer rows.Close()

	var savingAccounts []domain.SavingAccount
	for rows.Next() {
		var account domain.SavingAccount
		var interestRateBasisPoints sql.NullInt64
		var expirationDate sql.NullTime

		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Name,
			&account.AmountMinorUnits,
			&interestRateBasisPoints,
			&expirationDate,
			&account.Currency,
		)
		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		if interestRateBasisPoints.Valid {
			account.InterestRateBasisPoints = interestRateBasisPoints.Int64
		}
		if expirationDate.Valid {
			account.ExpirationDate = &expirationDate.Time
		}

		savingAccounts = append(savingAccounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapRepositoryError(err)
	}

	return savingAccounts, nil
}

var _ repository.SavingAccountRepository = (*SavingAccountRepository)(nil)
