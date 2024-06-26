package daos

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AttachmentDao is the data access object for attachments
type AttachmentDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAttachmentDao returns a new AttachmentDao
func NewAttachmentDao(db database.Database) *AttachmentDao {
	return &AttachmentDao{
		db:    db,
		table: "attachments",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *AttachmentDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of attachments
func (dao *AttachmentDao) Count(params *database.DatabaseParams, tx *database.Tx) (int, error) {
	generic := NewGenericDao(dao.db, dao)
	return generic.Count(params, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new attachment
func (dao *AttachmentDao) Create(a *models.Attachment, tx *database.Tx) error {
	if a.ID == "" {
		a.RefreshId()
	}

	a.RefreshCreatedAt()
	a.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table()).
		SetMap(dao.data(a)).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects an attachment with the given ID
//
// `tx` allows for the function to be run within a transaction
func (dao *AttachmentDao) Get(id string, tx *database.Tx) (*models.Attachment, error) {
	generic := NewGenericDao(dao.db, dao)

	dbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".id": id},
	}

	row, err := generic.Get(dbParams, tx)
	if err != nil {
		return nil, err
	}

	attachment, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List selects attachments
//
// `tx` allows for the function to be run within a transaction
func (dao *AttachmentDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Attachment, error) {
	generic := NewGenericDao(dao.db, dao)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Process the order by clauses
	dbParams.OrderBy = dao.ProcessOrderBy(dbParams.OrderBy, false)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	rows, err := generic.List(dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*models.Attachment

	for rows.Next() {
		a, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		attachments = append(attachments, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes an attachment based upon the where clause
//
// `tx` allows for the function to be run within a transaction
func (dao *AttachmentDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao)
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ProcessOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon columns() for the current
// DAO
func (dao *AttachmentDao) ProcessOrderBy(orderBy []string, explicit bool) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	generic := NewGenericDao(dao.db, dao)
	return generic.ProcessOrderBy(orderBy, dao.columns(), explicit)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default count select builder
func (dao *AttachmentDao) countSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *AttachmentDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *AttachmentDao) columns() []string {
	return []string{
		dao.Table() + ".*",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for an attachment
func (dao *AttachmentDao) data(a *models.Attachment) map[string]any {
	return map[string]any{
		"id":         a.ID,
		"course_id":  NilStr(a.CourseID),
		"asset_id":   NilStr(a.AssetID),
		"title":      NilStr(a.Title),
		"path":       NilStr(a.Path),
		"created_at": FormatTime(a.CreatedAt),
		"updated_at": FormatTime(a.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an attachment row
func (dao *AttachmentDao) scanRow(scannable Scannable) (*models.Attachment, error) {
	var a models.Attachment

	var createdAt string
	var updatedAt string

	err := scannable.Scan(
		&a.ID,
		&a.CourseID,
		&a.AssetID,
		&a.Title,
		&a.Path,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if a.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if a.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	return &a, nil
}
