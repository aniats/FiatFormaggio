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

func (r *UserRepository) EnsureUserExists(ctx context.Context, UserID domain.UserID, username string) (bool, error) {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE ID = $1)`
	err := r.DB.QueryRowContext(ctx, checkQuery, UserID).Scan(&exists)
	if err != nil {
		return false, errors.WrapRepositoryError(err)
	}

	if exists {
		updateQuery := `
            UPDATE users 
            SET username = $2, updated_at = NOW() 
            WHERE ID = $1 AND (username != $2 OR username IS NULL)
        `
		_, err = r.DB.ExecContext(ctx, updateQuery, UserID, username)
		return false, err
	}

	insertQuery := `
        INSERT INTO users (ID, username, created_at, updated_at) 
        VALUES ($1, $2, NOW(), NOW())
    `
	_, err = r.DB.ExecContext(ctx, insertQuery, UserID, username)
	if err != nil {
		return false, errors.WrapRepositoryError(err)
	}

	return true, nil
}

var _ repository.UserRepository = (*UserRepository)(nil)
