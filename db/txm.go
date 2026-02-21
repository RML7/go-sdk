package db

import (
	"context"

	"github.com/RML7/go-sdk/xerrors"
)

// Txm интерфейс менеджера транзакций
type Txm interface {
	Tx(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactionManager struct {
	db *DB
}

// NewTxm создает новый менеджер транзакций
func NewTxm(db *DB) Txm {
	return &transactionManager{db: db}
}

// Tx выполняет функцию внутри транзакции.
// Если транзакция уже есть в контексте, использует её.
// Иначе создает новую транзакцию, выполняет функцию и коммитит.
// При ошибке откатывает транзакцию.
func (t *transactionManager) Tx(ctx context.Context, fn func(ctx context.Context) error) error {
	// Если транзакция уже в контексте, будет использоваться она
	if tx := extractTx(ctx); tx != nil {
		return fn(ctx)
	}

	tx, err := t.db.Begin(ctx)
	if err != nil {
		return xerrors.WithMessage(err, "transaction begin failed")
	}

	ctx = injectTx(ctx, tx)

	// выполняем callback
	if err = fn(ctx); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			// Здесь можно добавить логирование, если у вас есть logger
			// logger.Error(ctx, "transaction rollback failed", errRollback)
			_ = errRollback
		}

		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return xerrors.WithMessage(err, "failed commit transaction")
	}

	return nil
}
