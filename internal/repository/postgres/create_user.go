package postgres

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (repo *Repository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (
            id, 
            username, 
            first_name, 
            last_name, 
            language_code, 
            timezone, 
            is_active
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	err := repo.db.QueryRowContext(
		ctx,
		query,
		user.UserId,
		user.Username,
		user.FirstName,
		user.LastName,
		user.LanguageCode,
		user.Timezone,
		user.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
