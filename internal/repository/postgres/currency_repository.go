package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type CurrencyRepository struct {
	*Database
}

func NewCurrencyRepository(db *Database) *CurrencyRepository {
	return &CurrencyRepository{Database: db}
}

func (r *CurrencyRepository) UpsertCurrencyRate(ctx context.Context, rate *domain.CurrencyRate) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			INSERT INTO currency_rates (
				currency, 
				rate_minor_units, 
				base_currency, 
				source, 
				updated_at
			) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (currency, base_currency) 
			DO UPDATE SET 
				rate_minor_units = EXCLUDED.rate_minor_units,
				source = EXCLUDED.source,
				updated_at = EXCLUDED.updated_at`

		_, err := r.DB.ExecContext(
			ctx,
			query,
			rate.Currency,
			rate.RateMinorUnits,
			rate.BaseCurrency,
			rate.Source,
			rate.UpdatedAt,
		)

		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		return nil, nil
	}

	_, err := r.ExecuteWithInterceptor(ctx, "Repository.UpsertCurrencyRate", handler)
	return err
}

func (r *CurrencyRepository) GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			SELECT 
				id,
				currency,
				rate_minor_units,
				base_currency,
				source,
				updated_at
			FROM currency_rates 
			ORDER BY currency`

		rows, err := r.DB.QueryContext(ctx, query)
		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}
		defer rows.Close()

		var rates []domain.CurrencyRate
		for rows.Next() {
			var rate domain.CurrencyRate

			err := rows.Scan(
				&rate.ID,
				&rate.Currency,
				&rate.RateMinorUnits,
				&rate.BaseCurrency,
				&rate.Source,
				&rate.UpdatedAt,
			)
			if err != nil {
				return nil, errors.WrapRepositoryError(err)
			}

			rates = append(rates, rate)
		}

		if err := rows.Err(); err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		return rates, nil
	}

	resultInterface, err := r.ExecuteWithInterceptor(ctx, "Repository.GetCurrencyRates", handler)
	if err != nil {
		return nil, err
	}

	if resultInterface != nil {
		return resultInterface.([]domain.CurrencyRate), nil
	}
	return nil, nil
}

func (r *CurrencyRepository) GetLastUpdateTime(ctx context.Context) (*time.Time, error) {
	query := `SELECT MAX(updated_at) FROM currency_rates`

	var lastUpdate sql.NullTime
	err := r.DB.QueryRowContext(ctx, query).Scan(&lastUpdate)
	if err != nil {
		return nil, errors.WrapRepositoryError(err)
	}

	if !lastUpdate.Valid {
		return nil, nil
	}

	return &lastUpdate.Time, nil
}

var _ repository.CurrencyRateRepository = (*CurrencyRepository)(nil)