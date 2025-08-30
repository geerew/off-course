package dao

import (
	"context"
	"database/sql"
	"errors"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DAO is a data access object
type DAO struct {
	db database.Database
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new DAO
func New(db database.Database) *DAO {
	return &DAO{db: db}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RunInTransaction is a wrapper for database.RunInTransaction
func RunInTransaction(ctx context.Context, dao *DAO, fn func(ctx context.Context) error) error {
	return dao.db.RunInTransaction(ctx, fn)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create is a generic function to create a record in the database
func createGeneric(ctx context.Context, dao *DAO, builderOpts builderOptions) error {
	sqlStr, args, err := insertBuilder(builderOpts)
	if err != nil {
		return err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	_, err = q.ExecContext(ctx, sqlStr, args...)
	return err
}

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // CreateOrReplace is a generic function to create or replace a model in the database
// func CreateOrReplace(ctx context.Context, dao *DAO, model models.Modeler) error {
// 	sch, err := schema.Parse(model)
// 	if err != nil {
// 		return err
// 	}

// 	if model.Id() == "" {
// 		model.RefreshId()
// 	}

// 	model.RefreshCreatedAt()
// 	model.RefreshUpdatedAt()

// 	q := database.QuerierFromContext(ctx, dao.db)
// 	// o := &database.Options{Replace: true}
// 	o := &database.Options{}
// 	_, err = sch.Insert(model, o, q)
// 	return err
// }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count is a generic function to count the number of rows in a table as determined by the model
func countGeneric(ctx context.Context, dao *DAO, builderOpts builderOptions) (int, error) {

	q := database.QuerierFromContext(ctx, dao.db)

	sqlStr, args, err := countBuilder(builderOpts)
	if err != nil {
		return -1, err
	}

	var count int
	err = q.GetContext(ctx, &count, sqlStr, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			count = 0
		} else {
			return -1, err
		}
	}

	return count, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getGeneric is a generic function to get a record from the database
func getGeneric[T any](ctx context.Context, dao *DAO, builderOpts builderOptions) (*T, error) {
	sqlStr, args, err := selectBuilder(builderOpts)
	if err != nil {
		return nil, err
	}

	q := database.QuerierFromContext(ctx, dao.db)

	record := new(T)
	err = q.GetContext(ctx, record, sqlStr, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return record, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getRow runs a query and returns a single row
func getRow(ctx context.Context, dao *DAO, builderOpts builderOptions) (*sql.Row, error) {
	sqlStr, args, err := selectBuilder(builderOpts)
	if err != nil {
		return nil, err
	}

	q := database.QuerierFromContext(ctx, dao.db)

	return q.QueryRowContext(ctx, sqlStr, args...), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// getRows runs a query and returns multiple rows
func getRows(ctx context.Context, dao *DAO, builderOpts builderOptions) (*sql.Rows, error) {
	q := database.QuerierFromContext(ctx, dao.db)

	if builderOpts.DbOpts != nil && builderOpts.DbOpts.Pagination != nil {
		count, err := countGeneric(ctx, dao, builderOpts)
		if err != nil {
			return nil, err
		}

		builderOpts.DbOpts.Pagination.SetCount(count)
	}

	sqlStr, args, err := selectBuilder(builderOpts)
	if err != nil {
		return nil, err
	}

	return q.QueryContext(ctx, sqlStr, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// listGeneric is a generic function to get records from the database
func listGeneric[T any](ctx context.Context, dao *DAO, builderOpts builderOptions) ([]*T, error) {
	q := database.QuerierFromContext(ctx, dao.db)

	if builderOpts.DbOpts != nil && builderOpts.DbOpts.Pagination != nil {
		count, err := countGeneric(ctx, dao, builderOpts)
		if err != nil {
			return nil, err
		}

		builderOpts.DbOpts.Pagination.SetCount(count)
	}

	sqlStr, args, err := selectBuilder(builderOpts)
	if err != nil {
		return nil, err
	}

	records := new([]*T)
	err = q.SelectContext(ctx, records, sqlStr, args...)
	if err != nil {
		return nil, err
	}

	return *records, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListPluck pulls a column into T.
func ListPluck[T any](ctx context.Context, dao *DAO, model any, options *database.Options, column string) (T, error) {
	var zero T

	sch, err := schema.Parse(model)
	if err != nil {
		return zero, err
	}

	var result T
	q := database.QuerierFromContext(ctx, dao.db)

	err = sch.Pluck(column, &result, options, q)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, nil
		}

		return zero, err
	}

	return result, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// updateGeneric is a generic function to update a record in the database
func updateGeneric(ctx context.Context, dao *DAO, builderOpts builderOptions) (bool, error) {
	sqlStr, args, err := updateBuilder(builderOpts)
	if err != nil {
		return false, err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	res, err := q.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return false, err
	}

	rowCount, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowCount > 0, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// principalFromCtx is a helper method to get the principal from the context
func principalFromCtx(ctx context.Context) (types.Principal, error) {
	principal, ok := ctx.Value(types.PrincipalContextKey).(types.Principal)
	if !ok {
		return types.Principal{}, utils.ErrPrincipal
	}

	return principal, nil
}
