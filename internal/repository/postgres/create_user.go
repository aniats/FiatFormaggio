package postgres

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
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

	_, err := repo.db.ExecContext(
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
		return errors.WrapRepositoryError(err)
	}

	return nil
}
