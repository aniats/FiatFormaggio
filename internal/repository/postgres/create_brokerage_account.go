package postgres

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
)

func (repo *Repository) CreateBrokerageAccount(ctx context.Context, account *domain.BrokerageAccount) error {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		query := `
			INSERT INTO brokerage_accounts (
				user_id, 
				name, 
				amount_minor_units, 
				currency, 
				broker_name, 
				account_type
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`

		var id int64

		err := repo.db.QueryRowContext(
			ctx,
			query,
			account.UserId,
			account.Name,
			account.AmountMinorUnits,
			account.Currency,
			account.Broker,
			account.AccountType,
		).Scan(&id)

		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		// Update the account object with the returned ID
		account.Id = id

		return nil, nil
	}

	wrappedHandler := repo.interceptor.Chain(handler, "Repository.CreateBrokerageAccount")
	_, err := wrappedHandler(ctx, account)
	return err
}
