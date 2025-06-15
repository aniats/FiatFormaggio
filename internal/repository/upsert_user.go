package repository

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
)

func (repo *PostgresRepository) EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error) {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := repo.db.QueryRowContext(ctx, checkQuery, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		updateQuery := `
            UPDATE users 
            SET username = $2, updated_at = NOW() 
            WHERE id = $1 AND (username != $2 OR username IS NULL)
        `
		_, err = repo.db.ExecContext(ctx, updateQuery, userID, username)
		return false, err
	}

	insertQuery := `
        INSERT INTO users (id, username, created_at, updated_at) 
        VALUES ($1, $2, NOW(), NOW())
    `
	_, err = repo.db.ExecContext(ctx, insertQuery, userID, username)
	if err != nil {
		return false, fmt.Errorf("failed to create user: %w", err)
	}

	return true, nil
}
