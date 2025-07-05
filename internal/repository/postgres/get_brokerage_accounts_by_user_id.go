package postgres

import (
	"context"
	"database/sql"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
)

func (repo *Repository) GetBrokerageAccountsByUserID(ctx context.Context, userId domain.UserId) ([]domain.BrokerageAccount, error) {
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

	rows, err := repo.db.QueryContext(ctx, query, userId)
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
