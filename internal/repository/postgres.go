package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(connString string) (*PostgresRepository, error) {
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

func (r *PostgresRepository) Init(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			balance DECIMAL(15,2) NOT NULL,
			currency TEXT NOT NULL,
			interest DECIMAL(5,2),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			account_id INTEGER REFERENCES accounts(id),
			amount DECIMAL(15,2) NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		
		CREATE TABLE IF NOT EXISTS profits (
			id SERIAL PRIMARY KEY,
			date DATE NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			account_id INTEGER REFERENCES accounts(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)

	return err
}
