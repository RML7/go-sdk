package db

import (
	"context"

	"github.com/RML7/go-sdk/closer"
	"github.com/RML7/go-sdk/xerrors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
}

type Config struct {
	DSN string `env:"DATABASE_DSN"`
}

// Executor возвращает DBExecutor из контекста (транзакция) или основной пул
func (db *DB) Executor(ctx context.Context) DBExecutor {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}

	return db.Pool
}

func NewDB(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.DSN == "" {
		return nil, xerrors.Errorf("dsn is required")
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, xerrors.Errorf("parse dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, xerrors.Errorf("create pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, xerrors.Errorf("ping db: %w", err)
	}

	closer.AddFunc(func(ctx context.Context) error {
		pool.Close()
		return nil
	})

	return &DB{Pool: pool}, nil
}
