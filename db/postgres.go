package db

import (
	"context"
	"fmt"

	"github.com/RML7/go-sdk/closer"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
}

// Executor возвращает DBExecutor из контекста (транзакция) или основной пул
func (db *DB) Executor(ctx context.Context) DBExecutor {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}

	return db.Pool
}

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	closer.AddFunc(func() error {
		pool.Close()
		return nil
	})

	return &DB{Pool: pool}, nil
}
