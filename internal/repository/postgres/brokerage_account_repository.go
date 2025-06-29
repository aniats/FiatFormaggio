package postgres

import (
	"context"
	"database/sql"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type BrokerageAccountRepository struct {
	*Database
}

func NewBrokerageAccountRepository(db *Database) *BrokerageAccountRepository {
	return &BrokerageAccountRepository{Database: db}
}

func (r *BrokerageAccountRepository) CreateBrokerageAccount(ctx context.Context, account *domain.BrokerageAccount) error {
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

		err := r.DB.QueryRowContext(
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

		account.Id = id
		return nil, nil
	}

	_, err := r.ExecuteWithInterceptor(ctx, "Repository.CreateBrokerageAccount", handler)
	return err
}

func (r *BrokerageAccountRepository) GetBrokerageAccountsByUserID(ctx context.Context, userId domain.UserId) ([]domain.BrokerageAccount, error) {
	query := `
        SELECT 
            id, 
            user_id, 
            name, 
            amount_minor_units, 
            currency, 
            broker_name,
            account_type
        FROM brokerage_accounts 
        WHERE user_id = $1
        ORDER BY created_at DESC`

	rows, err := r.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, errors.WrapRepositoryError(err)
	}
	defer rows.Close()

	var brokerageAccounts []domain.BrokerageAccount
	for rows.Next() {
		var account domain.BrokerageAccount
		var brokerName sql.NullString

		err := rows.Scan(
			&account.Id,
			&account.UserId,
			&account.Name,
			&account.AmountMinorUnits,
			&account.Currency,
			&brokerName,
			&account.AccountType,
		)
		if err != nil {
			return nil, errors.WrapRepositoryError(err)
		}

		if brokerName.Valid {
			account.Broker = &brokerName.String
		}

		brokerageAccounts = append(brokerageAccounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapRepositoryError(err)
	}

	return brokerageAccounts, nil
}

var _ repository.BrokerageAccountRepository = (*BrokerageAccountRepository)(nil)