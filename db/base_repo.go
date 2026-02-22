package db

import (
	"context"
	"errors"

	"github.com/RML7/go-sdk/xerrors"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type BaseRepo[E any, U any] struct {
	db        *DB
	tableName string
	idCol     string
}

var goquDB = goqu.Dialect("postgres")

type Filter interface {
	GetOffset() uint
	GetLimit() uint
	GetOrder() []exp.OrderedExpression
	GetExpressions() []goqu.Expression
}

func NewBaseRepo[E any, U any](db *DB, tableName, idCol string) BaseRepo[E, U] {
	return BaseRepo[E, U]{db: db, tableName: tableName, idCol: idCol}
}

func (r *BaseRepo[E, U]) Create(ctx context.Context, entity *E) (*E, error) {
	sql, args, err := goquDB.Insert(r.tableName).
		Rows(entity).
		Returning(goqu.Star()).
		ToSQL()
	if err != nil {
		return nil, xerrors.WithMessage(err, "build insert query")
	}

	rows, err := r.db.Executor(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, xerrors.WithMessage(err, "exec insert query")
	}

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[E])
	if err != nil {
		return nil, xerrors.WithMessage(err, "scan row")
	}

	return &result, nil
}

func (r *BaseRepo[E, U]) Get(ctx context.Context, id uuid.UUID) (*E, error) {
	sql, args, err := goquDB.From(r.tableName).
		Where(goqu.C(r.idCol).Eq(id)).
		ToSQL()
	if err != nil {
		return nil, xerrors.WithMessage(err, "build select query")
	}

	rows, err := r.db.Executor(ctx).Query(ctx, sql, args...)
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

func (r *BaseRepo[E, U]) List(ctx context.Context, filter Filter) ([]*E, error) {
	query := goquDB.From(r.tableName).
		Offset(filter.GetOffset()).
		Limit(filter.GetLimit()).
		Order(filter.GetOrder()...).
		Where(filter.GetExpressions()...)

	sql, args, err := query.ToSQL()
	if err != nil {
		return nil, xerrors.WithMessage(err, "build list query")
	}

	rows, err := r.db.Executor(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, xerrors.WithMessage(err, "exec list query")
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[E])
	if err != nil {
		return nil, xerrors.WithMessage(err, "scan rows")
	}

	return result, nil
}

func (r *BaseRepo[E, U]) Count(ctx context.Context, filter Filter) (int64, error) {
	query := goquDB.Select(goqu.COUNT(goqu.Star())).
		From(r.tableName).
		Where(filter.GetExpressions()...)

	sql, args, err := query.ToSQL()
	if err != nil {
		return 0, xerrors.WithMessage(err, "build count query")
	}

	rows, err := r.db.Executor(ctx).Query(ctx, sql, args...)
	if err != nil {
		return 0, xerrors.WithMessage(err, "exec count query")
	}

	count, err := pgx.CollectOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		return 0, xerrors.WithMessage(err, "scan count")
	}

	return count, nil
}

func (r *BaseRepo[E, U]) Exists(ctx context.Context, filter Filter) (bool, error) {
	query := goquDB.Select(
		goqu.L("EXISTS (?)",
			goquDB.From(r.tableName).
				Where(filter.GetExpressions()...),
		).As("exists"),
	)

	sql, args, err := query.ToSQL()
	if err != nil {
		return false, xerrors.WithMessage(err, "build exists query")
	}

	rows, err := r.db.Executor(ctx).Query(ctx, sql, args...)
	if err != nil {
		return false, xerrors.WithMessage(err, "exec exists query")
	}

	exists, err := pgx.CollectOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, xerrors.WithMessage(err, "scan exists")
	}

	return exists, nil
}

func (r *BaseRepo[E, U]) Update(ctx context.Context, id uuid.UUID, update *U) (*E, error) {
	f, err := fields(*update)
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

	rows, err := r.db.Executor(ctx).Query(ctx, sql, args...)
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
