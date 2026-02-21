package db

import (
	"context"
	"errors"

	"github.com/RML7/go-sdk/xerrors"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type baseRepo[E any, U any] struct {
	db        *DB
	tableName string
	idCol     string
}

var goquDB = goqu.Dialect("postgres")

func newBaseRepo[E any, U any](db *DB, tableName, idCol string) baseRepo[E, U] {
	return baseRepo[E, U]{db: db, tableName: tableName, idCol: idCol}
}

func (r *baseRepo[E, U]) Create(ctx context.Context, entity *E) (*E, error) {
	sql, args, err := goquDB.Insert(r.tableName).
		Rows(entity).
		Returning(goqu.Star()).
		ToSQL()
	if err != nil {
		return nil, xerrors.WithMessage(err, "build insert query")
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, xerrors.WithMessage(err, "exec insert query")
	}

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[E])
	if err != nil {
		return nil, xerrors.WithMessage(err, "scan row")
	}

	return &result, nil
}

func (r *baseRepo[E, U]) Get(ctx context.Context, id uuid.UUID) (*E, error) {
	sql, args, err := goquDB.From(r.tableName).
		Where(goqu.C(r.idCol).Eq(id)).
		ToSQL()
	if err != nil {
		return nil, xerrors.WithMessage(err, "build select query")
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, xerrors.WithMessage(err, "exec select query")
	}

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[E])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrNotFound
		}
		return nil, xerrors.WithMessage(err, "scan row")
	}

	return &result, nil
}

func (r *baseRepo[E, U]) Update(ctx context.Context, id uuid.UUID, update U) (*E, error) {
	f, err := fields(update)
	if err != nil {
		return nil, xerrors.WithMessage(err, "build update fields")
	}

	if len(f) == 0 {
		return nil, xerrors.Errorf("no fields for update")
	}

	sql, args, err := goquDB.Update(r.tableName).
		Set(goqu.Record(f)).
		Where(goqu.C(r.idCol).Eq(id)).
		Returning(goqu.Star()).
		ToSQL()
	if err != nil {
		return nil, xerrors.WithMessage(err, "build update query")
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, xerrors.WithMessage(err, "exec update query")
	}

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[E])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrNotFound
		}
		return nil, xerrors.WithMessage(err, "scan row")
	}

	return &result, nil
}
