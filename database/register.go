package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"gosalusa.com/clog"
	"gosalusa.com/di"
)

type Upper interface {
	Up(ctx context.Context, db DB) error
}

func Register(ctx context.Context, config Config, migrations Upper) {
	di.RegisterLazySingleton(ctx, func() (*sqlx.DB, error) {
		db, err := sqlx.Open(config.DriverName(), config.DataSourceName())
		if err != nil {
			return nil, fmt.Errorf("database.Register: open database: %w", err)

		}
		if migrations != nil {
			err = migrations.Up(ctx, db)
			if err != nil {
				return nil, fmt.Errorf("database.Register: migrate database: %w", err)
			}
		}
		clog.Use(ctx).Info("database ready")

		return db, nil
	})
	RegisterTransactions(ctx, nil)
}

func RegisterDB(ctx context.Context, db *sqlx.DB) {
	di.RegisterSingleton(ctx, func() *sqlx.DB {
		return db
	})
	RegisterTransactions(ctx, nil)
}

func RegisterTransactions(ctx context.Context, mtx sync.Locker) {
	di.RegisterWith(ctx, func(ctx context.Context, tag string, db *sqlx.DB) (Read, error) {
		return NewRead(ctx, mtx, db), nil
	})
	di.RegisterWith(ctx, func(ctx context.Context, tag string, db *sqlx.DB) (Update, error) {
		return NewUpdate(ctx, mtx, db), nil
	})
}
