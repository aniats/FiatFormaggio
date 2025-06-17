package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aniats/FiatFormaggio/internal/repository"
	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func New(connString string) (*Repository, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("couldn't establish a successful connection to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("couldn't ping: %v", err)
	}

	return &Repository{db: db}, nil
}

func (repo *Repository) Close() error {
	return repo.db.Close()
}

func (repo *Repository) HealthCheck(ctx context.Context) error {
	return repo.db.PingContext(ctx)
}

// Compile-time check to ensure Repository implements repository.Repository interface
var _ repository.Repository = (*Repository)(nil)
