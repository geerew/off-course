package daos

import (
	"database/sql"

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
		table: TableAttachments(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableAttachments returns the name of the attachments table
func TableAttachments() string {
	return "attachments"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of attachments
func (dao *AttachmentDao) Count(params *database.DatabaseParams) (int, error) {
	generic := NewGenericDao(dao.db, dao.table)
	return generic.Count(dao.baseSelect(), params, nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new attachment
func (dao *AttachmentDao) Create(a *models.Attachment) error {
	if a.ID == "" {
		a.RefreshId()
	}

	a.RefreshCreatedAt()
	a.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.table).
		SetMap(dao.data(a)).
		ToSql()

	_, err := dao.db.Exec(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects an attachment with the given ID
//
// `tx` allows for the function to be run within a transaction
func (dao *AttachmentDao) Get(id string, tx *sql.Tx) (*models.Attachment, error) {
	generic := NewGenericDao(dao.db, dao.table)

	dbParams := &database.DatabaseParams{
		Columns: dao.selectColumns(),
		Where:   squirrel.Eq{generic.table + ".id": id},
	}

	row, err := generic.Get(dao.baseSelect(), dbParams, tx)
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
func (dao *AttachmentDao) List(dbParams *database.DatabaseParams, tx *sql.Tx) ([]*models.Attachment, error) {
	generic := NewGenericDao(dao.db, dao.table)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Process the order by clauses
	dbParams.OrderBy = dao.processOrderBy(dbParams.OrderBy)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.selectColumns()
	}

	rows, err := generic.List(dao.baseSelect(), dbParams, tx)
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

// Delete deletes an attachment with the given ID
//
// `tx` allows for the function to be run within a transaction
func (dao *AttachmentDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao.table)
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *AttachmentDao) baseSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.table).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// selectColumns returns the columns to select
func (dao *AttachmentDao) selectColumns() []string {
	return []string{
		dao.table + ".*",
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
		"created_at": a.CreatedAt,
		"updated_at": a.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// processOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon selectColumns() for the current
// DAO
func (dao *AttachmentDao) processOrderBy(orderBy []string) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	validTableColumns := dao.selectColumns()
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

// scanRow scans an attachment row
func (dao *AttachmentDao) scanRow(scannable Scannable) (*models.Attachment, error) {
	var a models.Attachment

	err := scannable.Scan(
		&a.ID,
		&a.CourseID,
		&a.AssetID,
		&a.Title,
		&a.Path,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &a, nil
}
