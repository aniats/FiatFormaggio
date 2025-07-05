package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/middleware"
	"github.com/aniats/FiatFormaggio/internal/repository"
	_ "github.com/lib/pq"
)

type Repository struct {
	db          *middleware.ProfiledDB
	interceptor *middleware.UnifiedInterceptor
}

func New(connString string) (*Repository, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, errors.WrapDatabaseError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.WrapDatabaseError(err)
	}

	interceptor := middleware.NewUnifiedInterceptor(middleware.DefaultConfig("Repository"))
	profiledDB := middleware.ProfileDatabase(db, true, true, 10*time.Millisecond)

	return &Repository{
		db:          profiledDB,
		interceptor: interceptor,
	}, nil
}

func (repo *Repository) Close() error {
	return repo.db.Close()
}

func (repo *Repository) HealthCheck(ctx context.Context) error {
	return repo.db.PingContext(ctx)
}

var _ repository.Repository = (*Repository)(nil)
