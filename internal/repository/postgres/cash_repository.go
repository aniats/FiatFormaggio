package postgres

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type CashRepository struct {
	*Database
}

func NewCashRepository(db *Database) *CashRepository {
	return &CashRepository{Database: db}
}

func (r *CashRepository) CreateCashHolding(ctx context.Context, cash *domain.CashHolding) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			INSERT INTO cash_holdings (
				user_id, 
				name, 
				amount_minor_units, 
				currency
			) VALUES ($1, $2, $3, $4)
			RETURNING ID
		`

		var ID int64

		err := r.DB.QueryRowContext(
			ctx,
			query,
			cash.UserID,
			cash.Name,
			cash.AmountMinorUnits,
			cash.Currency,
		).Scan(&ID)

		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		cash.ID = ID
		return nil, nil
	}

	_, err := r.ExecuteWithInterceptor(ctx, "Repository.CreateCashHolding", handler)
	return err
}

func (r *CashRepository) GetCashHoldingsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.CashHolding, error) {
	query := `
        SELECT 
            ID, 
            user_id, 
            name, 
            amount_minor_units, 
            currency
        FROM cash_holdings 
        WHERE user_id = $1
        ORDER BY created_at DESC`

	rows, err := r.DB.QueryContext(ctx, query, UserID)
	if err != nil {
		return nil, errors.WrapRepositoryError(err)
	}
	defer rows.Close()

	var cashHoldings []domain.CashHolding
	for rows.Next() {
		var cash domain.CashHolding

		err := rows.Scan(
			&cash.ID,
			&cash.UserID,
			&cash.Name,
			&cash.AmountMinorUnits,
			&cash.Currency,
		)
		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		cashHoldings = append(cashHoldings, cash)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapRepositoryError(err)
	}

	return cashHoldings, nil
}

var _ repository.CashHoldingRepository = (*CashRepository)(nil)
