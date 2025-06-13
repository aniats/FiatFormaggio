package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresRepository implements the Repository interface
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository - creates a new PostgreSQL repository instance
func NewPostgresRepository(connString string) (Repository, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("couldn't establish a successful connection to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("couldn't ping: %v", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Close closes the database connection
func (repo *PostgresRepository) Close() error {
	return repo.db.Close()
}

// HealthCheck verifies the database connection is healthy
func (repo *PostgresRepository) HealthCheck(ctx context.Context) error {
	return repo.db.PingContext(ctx)
}

// Compile-time check to ensure PostgresRepository implements Repository interface
var _ Repository = (*PostgresRepository)(nil)
