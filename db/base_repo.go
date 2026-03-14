package db

import (
	"context"
	"errors"

	"github.com/RML7/go-sdk/xerrors"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/jackc/pgx/v5"
)

type BaseRepo[E any, U any, F Filter, ID comparable] struct {
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

type BaseRepoOptions struct {
	TableName string
	IDCol     string
}

func NewBaseRepo[E any, U any, F Filter, ID comparable](db *DB, opts BaseRepoOptions) (BaseRepo[E, U, F, ID], error) {
	if db == nil {
		return BaseRepo[E, U, F, ID]{}, xerrors.Errorf("db is required")
	}
	if opts.TableName == "" {
		return BaseRepo[E, U, F, ID]{}, xerrors.Errorf("table name is required")
	}
	if opts.IDCol == "" {
		return BaseRepo[E, U, F, ID]{}, xerrors.Errorf("id column is required")
	}

	return BaseRepo[E, U, F, ID]{db: db, tableName: opts.TableName, idCol: opts.IDCol}, nil
}

func (r *BaseRepo[E, U, F, ID]) Create(ctx context.Context, entity *E) (*E, error) {
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

func (r *BaseRepo[E, U, F, ID]) BulkCreate(ctx context.Context, entities []*E) ([]*E, error) {
	if len(entities) == 0 {
		return []*E{}, nil
	}

	sql, args, err := goquDB.Insert(r.tableName).
		Rows(entities).
		Returning(goqu.Star()).
		ToSQL()
	if err != nil {
		return nil, xerrors.WithMessage(err, "build bulk insert query")
	}

	rows, err := r.db.Executor(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, xerrors.WithMessage(err, "exec bulk insert query")
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[E])
	if err != nil {
		return nil, xerrors.WithMessage(err, "scan rows")
	}

	return result, nil
}

type ReadOption func(*goqu.SelectDataset) *goqu.SelectDataset

func (r *BaseRepo[E, U, F, ID]) Get(ctx context.Context, id ID, opts ...ReadOption) (*E, error) {
	query := goquDB.From(r.tableName).
		Where(goqu.C(r.idCol).Eq(id))

	for _, opt := range opts {
		query = opt(query)
	}

	sql, args, err := query.ToSQL()
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

func (r *BaseRepo[E, U, F, ID]) List(ctx context.Context, filter F, opts ...ReadOption) ([]*E, error) {
	query := goquDB.From(r.tableName).
		Offset(filter.GetOffset()).
		Limit(filter.GetLimit()).
		Order(filter.GetOrder()...).
		Where(filter.GetExpressions()...)

	for _, opt := range opts {
		query = opt(query)
	}

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

func (r *BaseRepo[E, U, F, ID]) Count(ctx context.Context, filter F) (uint32, error) {
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

	count, err := pgx.CollectOneRow(rows, pgx.RowTo[uint32])
	if err != nil {
		return 0, xerrors.WithMessage(err, "scan count")
	}

	return count, nil
}

func (r *BaseRepo[E, U, F, ID]) Exists(ctx context.Context, filter F) (bool, error) {
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

func (r *BaseRepo[E, U, F, ID]) Update(ctx context.Context, id ID, update *U) (*E, error) {
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
