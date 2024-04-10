package daos

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseTagDao is the data access object for courses tags
type CourseTagDao struct {
	db    database.Database
	Table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseTagDao returns a new CourseTagDao
func NewCourseTagDao(db database.Database) *CourseTagDao {
	return &CourseTagDao{
		db:    db,
		Table: "courses_tags",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of course-tags
func (dao *CourseTagDao) Count(dbParams *database.DatabaseParams) (int, error) {
	generic := NewGenericDao(dao.db, dao.Table)
	return generic.Count(dao.baseSelect(), dbParams, nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new course-tag and tag if it does not exist
//
// If `tx` is nil, the function will create a new transaction, else it will use the current
// transaction
func (dao *CourseTagDao) Create(ct *models.CourseTag, tagValue string, tx *sql.Tx) error {
	if tx == nil {
		return dao.db.RunInTransaction(func(tx *sql.Tx) error {
			return dao.create(ct, tagValue, tx)
		})
	} else {
		return dao.create(ct, tagValue, tx)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List selects courses
//
// `tx` allows for the function to be run within a transaction
func (dao *CourseTagDao) List(dbParams *database.DatabaseParams, tx *sql.Tx) ([]*models.CourseTag, error) {
	generic := NewGenericDao(dao.db, dao.Table)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Process the order by clauses
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

	var cts []*models.CourseTag

	for rows.Next() {
		ct, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		cts = append(cts, ct)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cts, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes a course-tag based upon the where clause
//
// `tx` allows for the function to be run within a transaction
func (dao *CourseTagDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao.Table)
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// create inserts a new course-tag and tag if it does not exist
//
// This function is used by Create() and always runs within a transaction
func (dao *CourseTagDao) create(ct *models.CourseTag, tagValue string, tx *sql.Tx) error {
	if tx == nil {
		return ErrNilTransaction
	}

	if ct.ID == "" {
		ct.RefreshId()
	}

	ct.RefreshCreatedAt()
	ct.RefreshUpdatedAt()

	// Check if the tag exists. This should return 0 or 1 tags as tags are unique
	tagDao := NewTagDao(dao.db)

	tags, err := tagDao.List(&database.DatabaseParams{Where: squirrel.Eq{"tag": tagValue}}, tx)
	if err != nil {
		return err
	}

	// Create the tag if it doesn't exist
	if len(tags) == 0 {
		tag := &models.Tag{
			Tag: tagValue,
		}

		if err := tagDao.Create(tag, tx); err != nil {
			return err
		}

		ct.TagId = tag.ID
	} else {
		ct.TagId = tags[0].ID
	}

	// Insert the course-tag
	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table).
		SetMap(dao.data(ct)).
		ToSql()

	_, err = tx.Exec(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// It performs 2 left joins
//   - courses table to get `title`
//   - tags table to get `tag`
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *CourseTagDao) baseSelect() squirrel.SelectBuilder {
	tagDao := NewTagDao(dao.db)
	courseDao := NewCourseDao(dao.db)

	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table).
		LeftJoin(courseDao.Table + " ON " + dao.Table + ".course_id = " + courseDao.Table + ".id").
		LeftJoin(tagDao.Table + " ON " + dao.Table + ".tag_id = " + tagDao.Table + ".id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *CourseTagDao) columns() []string {
	tagDao := NewTagDao(dao.db)
	courseDao := NewCourseDao(dao.db)

	return []string{
		dao.Table + ".*",
		courseDao.Table + ".title as course",
		tagDao.Table + ".tag",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a course-tag
func (dao *CourseTagDao) data(ct *models.CourseTag) map[string]any {
	return map[string]any{
		"id":         ct.ID,
		"tag_id":     NilStr(ct.TagId),
		"course_id":  NilStr(ct.CourseId),
		"created_at": ct.CreatedAt,
		"updated_at": ct.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// processOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon columns() for the current
// DAO
func (dao *CourseTagDao) processOrderBy(orderBy []string) []string {
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

// scanRow scans a course-tag row
func (dao *CourseTagDao) scanRow(scannable Scannable) (*models.CourseTag, error) {
	var ct models.CourseTag

	err := scannable.Scan(
		&ct.ID,
		&ct.TagId,
		&ct.CourseId,
		&ct.CreatedAt,
		&ct.UpdatedAt,
		&ct.Course,
		&ct.Tag,
	)

	if err != nil {
		return nil, err
	}

	return &ct, nil
}