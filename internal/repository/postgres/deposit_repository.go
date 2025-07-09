package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type DepositRepository struct {
	*Database
}

func NewDepositRepository(db *Database) *DepositRepository {
	return &DepositRepository{Database: db}
}

func (r *DepositRepository) CreateDeposit(ctx context.Context, deposit *domain.Deposit) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			INSERT INTO deposits (
				user_id, 
				name, 
				amount_minor_units, 
				interest_rate_basis_points, 
				expiration_date, 
				currency
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING ID, expiration_date, created_at, updated_at
		`

		var ID int64
		var expirationDate *time.Time
		var createdAt, updatedAt time.Time

		err := r.DB.QueryRowContext(
			ctx,
			query,
			deposit.UserID,
			deposit.Name,
			deposit.AmountMinorUnits,
			deposit.InterestRateBasisPoints,
			deposit.ExpirationDate,
			deposit.Currency,
		).Scan(&ID, &expirationDate, &createdAt, &updatedAt)

		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		deposit.ID = ID
		deposit.ExpirationDate = expirationDate

		return nil, nil
	}

	_, err := r.ExecuteWithInterceptor(ctx, "Repository.CreateDeposit", handler)
	return err
}

func (r *DepositRepository) GetDepositsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.Deposit, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			SELECT
				ID,
				user_id,
				name,
				amount_minor_units, 
				interest_rate_basis_points, 
				expiration_date, 
				currency
			FROM deposits 
			WHERE user_id = $1
			ORDER BY created_at DESC`

		rows, queryErr := r.DB.QueryContext(ctx, query, UserID)
		if queryErr != nil {
			return nil, errors.WrapRepositoryError(queryErr)
		}
		defer rows.Close()

		var deposits []domain.Deposit
		for rows.Next() {
			var deposit domain.Deposit
			var expirationDate sql.NullTime
			var interestRateBasisPoints sql.NullInt64

			scanErr := rows.Scan(
				&deposit.ID,
				&deposit.UserID,
				&deposit.Name,
				&deposit.AmountMinorUnits,
				&interestRateBasisPoints,
				&expirationDate,
				&deposit.Currency,
			)
			if scanErr != nil {
				return nil, errors.WrapRepositoryError(scanErr)
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
			return nil, errors.WrapRepositoryError(rowsErr)
		}

		return deposits, nil
	}

	resultInterface, err := r.ExecuteWithInterceptor(ctx, "Repository.GetDepositsByUserID", handler)
	if err != nil {
		return nil, err
	}

	if resultInterface != nil {
		return resultInterface.([]domain.Deposit), nil
	}
	return nil, nil
}

var _ repository.DepositRepository = (*DepositRepository)(nil)
