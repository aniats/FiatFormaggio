package repository

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

// CreateBrokerageAccount - creates a new brokerage account for a user
func (repo *PostgresRepository) CreateBrokerageAccount(ctx context.Context, account *domain.BrokerageAccount) error {
	query := `
        INSERT INTO brokerage_accounts (
            user_id, 
            name, 
            amount_minor_units, 
            currency, 
            broker_name, 
            account_type
        ) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := repo.db.ExecContext(
		ctx,
		query,
		account.UserId,
		account.Name,
		account.AmountMinorUnits,
		account.Currency,
		account.Broker,
		account.AccountType,
	)

	if err != nil {
		return fmt.Errorf("failed to create brokerage account: %w", err)
	}

	return nil
}
