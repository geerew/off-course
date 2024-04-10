package daos

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TagDao is the data access object for tags
type TagDao struct {
	db    database.Database
	Table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTagDao returns a new TagDao
func NewTagDao(db database.Database) *TagDao {
	return &TagDao{
		db:    db,
		Table: "tags",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of tags
func (dao *TagDao) Count(params *database.DatabaseParams) (int, error) {
	generic := NewGenericDao(dao.db, dao.Table)
	return generic.Count(dao.baseSelect(), params, nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new tag
//
// `tx` allows for the function to be run within a transaction
func (dao *TagDao) Create(t *models.Tag, tx *sql.Tx) error {
	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	if t.ID == "" {
		t.RefreshId()
	}

	t.RefreshCreatedAt()
	t.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table).
		SetMap(dao.data(t)).
		ToSql()

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List selects tags
//
// `tx` allows for the function to be run within a transaction
func (dao *TagDao) List(dbParams *database.DatabaseParams, tx *sql.Tx) ([]*models.Tag, error) {
	generic := NewGenericDao(dao.db, dao.Table)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	dbParams.OrderBy = dao.processOrderBy(dbParams.OrderBy)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	rows, err := generic.List(dao.baseSelect(), dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*models.Tag

	for rows.Next() {
		t, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		tags = append(tags, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes a tag based upon the where clause
//
// `tx` allows for the function to be run within a transaction
func (dao *TagDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao.Table)
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *TagDao) baseSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *TagDao) columns() []string {
	return []string{
		dao.Table + ".*",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a tag
func (dao *TagDao) data(t *models.Tag) map[string]any {
	return map[string]any{
		"id":         t.ID,
		"tag":        NilStr(t.Tag),
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// processOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon columns() for the current
// DAO
func (dao *TagDao) processOrderBy(orderBy []string) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	validTableColumns := dao.columns()
	var processedOrderBy []string

	for _, ob := range orderBy {
		table, column := extractTableColumn(ob)

		if isValidOrderBy(table, column, validTableColumns) {
			processedOrderBy = append(processedOrderBy, ob)
		}
	}

	return processedOrderBy
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an tag row
func (dao *TagDao) scanRow(scannable Scannable) (*models.Tag, error) {
	var t models.Tag

	err := scannable.Scan(
		&t.ID,
		&t.Tag,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &t, nil
}