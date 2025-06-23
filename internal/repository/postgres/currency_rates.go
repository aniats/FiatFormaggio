package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (repo *Repository) UpsertCurrencyRate(ctx context.Context, rate *domain.CurrencyRate) error {
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

		_, err := repo.db.ExecContext(
			ctx,
			query,
			rate.Currency,
			rate.RateMinorUnits,
			rate.BaseCurrency,
			rate.Source,
			rate.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to upsert currency rate: %w", err)
		}

		return nil, nil
	}

	wrappedHandler := repo.interceptor.Chain(handler, "Repository.UpsertCurrencyRate")
	_, err := wrappedHandler(ctx, rate)
	return err
}

func (repo *Repository) GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error) {
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

		rows, err := repo.db.QueryContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query currency rates: %w", err)
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
				return nil, fmt.Errorf("failed to scan currency rate row: %w", err)
			}

			rates = append(rates, rate)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over currency rate rows: %w", err)
		}

		return rates, nil
	}

	wrappedHandler := repo.interceptor.Chain(handler, "Repository.GetCurrencyRates")
	resultInterface, err := wrappedHandler(ctx, nil)
	if err != nil {
		return nil, err
	}

	if resultInterface != nil {
		return resultInterface.([]domain.CurrencyRate), nil
	}
	return nil, nil
}

func (repo *Repository) GetLastUpdateTime(ctx context.Context) (*time.Time, error) {
	query := `SELECT MAX(updated_at) FROM currency_rates`

	var lastUpdate sql.NullTime
	err := repo.db.QueryRowContext(ctx, query).Scan(&lastUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to get last update time: %w", err)
	}

	if !lastUpdate.Valid {
		return nil, nil // No rates in database yet
	}

	return &lastUpdate.Time, nil
}
