package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/middleware"
	_ "github.com/lib/pq"
)

type Database struct {
	DB          *middleware.ProfiledDB
	Interceptor *middleware.UnifiedInterceptor
}

func NewDatabase(connString string) (*Database, error) {
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

	return &Database{
		DB:          profiledDB,
		Interceptor: interceptor,
	}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}

func (d *Database) HealthCheck(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}

func (d *Database) ExecuteWithInterceptor(ctx context.Context, operationName string, operation func(context.Context, interface{}) (interface{}, error)) (interface{}, error) {
	wrappedHandler := d.Interceptor.Chain(operation, operationName)
	return wrappedHandler(ctx, nil)
}
