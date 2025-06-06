package dao

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
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

// Create is a generic function to create a model in the database
func Create(ctx context.Context, dao *DAO, model models.Modeler) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	if model.Id() == "" {
		model.RefreshId()
	}

	model.RefreshCreatedAt()
	model.RefreshUpdatedAt()

	q := database.QuerierFromContext(ctx, dao.db)
	_, err = sch.Insert(model, nil, q)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateOrReplace is a generic function to create or replace a model in the database
func CreateOrReplace(ctx context.Context, dao *DAO, model models.Modeler) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	if model.Id() == "" {
		model.RefreshId()
	}

	model.RefreshCreatedAt()
	model.RefreshUpdatedAt()

	q := database.QuerierFromContext(ctx, dao.db)
	o := &database.Options{Replace: true}
	_, err = sch.Insert(model, o, q)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count is a generic function to count the number of rows in a table as determined by the model
func Count(ctx context.Context, dao *DAO, model any, options *database.Options) (int, error) {
	sch, err := schema.Parse(model)
	if err != nil {
		return 0, err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	return sch.Count(options, q)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get is a generic function to get a model (row)
func Get(ctx context.Context, dao *DAO, model any, options *database.Options) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	return sch.Select(model, options, q)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List is a generic function to list models (rows)
func List(ctx context.Context, dao *DAO, model any, options *database.Options) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	if options != nil && options.Pagination != nil {
		count, err := Count(ctx, dao, model, options)
		if err != nil {
			return err
		}

		options.Pagination.SetCount(count)
	}

	q := database.QuerierFromContext(ctx, dao.db)
	err = sch.Select(model, options, q)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
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

// Update is a generic function to update a model in the database
func Update(ctx context.Context, dao *DAO, model models.Modeler) (bool, error) {
	sch, err := schema.Parse(model)
	if err != nil {
		return false, err
	}

	if model.Id() == "" {
		return false, utils.ErrInvalidId
	}

	model.RefreshUpdatedAt()

	q := database.QuerierFromContext(ctx, dao.db)
	res, err := sch.Update(model, &database.Options{Where: squirrel.Eq{model.Table() + "." + models.BASE_ID: model.Id()}}, q)
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

// Delete is a generic function to delete a model (row)
//
// When options is nil or options.Where is nil, the model ID will be used
func Delete(ctx context.Context, dao *DAO, model models.Modeler, options *database.Options) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	if options == nil || options.Where == nil {
		if model.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{Where: squirrel.Eq{model.Table() + "." + models.BASE_ID: model.Id()}}
	}

	q := database.QuerierFromContext(ctx, dao.db)
	_, err = sch.Delete(options, q)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAll is a generic function to delete all rows in a table as determined by the model
func DeleteAll(ctx context.Context, dao *DAO, model models.Modeler) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	_, err = sch.Delete(nil, q)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RawQuery runs any SQL and scans into model (struct or slice)
func RawQuery(ctx context.Context, dao *DAO, model any, query string, args ...any) error {
	sch, err := schema.Parse(model)
	if err != nil {
		return err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	return sch.RawSelect(model, query, args, q)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RawExec runs any SQL that doesn’t return rows (INSERT/UPDATE/DELETE/etc)
func RawExec(ctx context.Context, dao *DAO, query string, args ...any) (sql.Result, error) {
	q := database.QuerierFromContext(ctx, dao.db)
	return q.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// principalFromCtx is a helper method to get the principal from the context
func principalFromCtx(ctx context.Context) (types.Principal, error) {
	principal, ok := ctx.Value(types.PrincipalContextKey).(types.Principal)
	if !ok {
		return types.Principal{}, utils.ErrMissingPrincipal
	}

	return principal, nil
}
