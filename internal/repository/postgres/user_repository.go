package postgres

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type UserRepository struct {
	*Database
}

func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{Database: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
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

	_, err := r.DB.ExecContext(
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

func (r *UserRepository) EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error) {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := r.DB.QueryRowContext(ctx, checkQuery, userID).Scan(&exists)
	if err != nil {
		return false, errors.WrapRepositoryError(err)
	}

	if exists {
		updateQuery := `
            UPDATE users 
            SET username = $2, updated_at = NOW() 
            WHERE id = $1 AND (username != $2 OR username IS NULL)
        `
		_, err = r.DB.ExecContext(ctx, updateQuery, userID, username)
		return false, err
	}

	insertQuery := `
        INSERT INTO users (id, username, created_at, updated_at) 
        VALUES ($1, $2, NOW(), NOW())
    `
	_, err = r.DB.ExecContext(ctx, insertQuery, userID, username)
	if err != nil {
		return false, errors.WrapRepositoryError(err)
	}

	return true, nil
}

var _ repository.UserRepository = (*UserRepository)(nil)