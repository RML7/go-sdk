package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type txKey struct{}

// injectTx вставляет транзакцию в контекст
func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// extractTx извлекает транзакцию из контекста
func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}
