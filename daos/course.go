package daos

import (
	"database/sql"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseDao is the data access object for courses
type CourseDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseDao returns a new CourseDao
func NewCourseDao(db database.Database) *CourseDao {
	return &CourseDao{
		db:    db,
		table: "courses",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table returns the table name
func (dao *CourseDao) Table() string {
	return dao.table
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of courses
func (dao *CourseDao) Count(params *database.DatabaseParams) (int, error) {
	generic := NewGenericDao(dao.db, dao)
	return generic.Count(params, nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new course and courses_progress row within a transaction
//
// NOTE: There is currently no support for users, but when there is, the default courses_progress
// should be inserted for the admin user
func (dao *CourseDao) Create(c *models.Course) error {
	if c.ID == "" {
		c.RefreshId()
	}

	c.RefreshCreatedAt()
	c.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table()).
		SetMap(dao.data(c)).
		ToSql()

	return dao.db.RunInTransaction(func(tx *database.Tx) error {
		// Create the course
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}

		// Create the course progress
		cp := &models.CourseProgress{
			CourseID: c.ID,
		}

		cpDao := NewCourseProgressDao(dao.db)
		return cpDao.Create(cp, tx)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects a course with the given ID
//
// `tx` allows for the function to be run within a transaction
func (dao *CourseDao) Get(id string, dbParams *database.DatabaseParams, tx *database.Tx) (*models.Course, error) {
	generic := NewGenericDao(dao.db, dao)

	courseDbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".id": id},
	}

	row, err := generic.Get(courseDbParams, tx)
	if err != nil {
		return nil, err
	}

	course, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List selects courses
//
// `tx` allows for the function to be run within a transaction
func (dao *CourseDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Course, error) {
	generic := NewGenericDao(dao.db, dao)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Process the order by clauses
	dbParams.OrderBy = dao.ProcessOrderBy(dbParams.OrderBy)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	rows, err := generic.List(dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []*models.Course

	for rows.Next() {
		c, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		courses = append(courses, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates a course
//
// Note: Only `card_path` and `available` can be updated
func (dao *CourseDao) Update(course *models.Course, tx *database.Tx) error {
	if course.ID == "" {
		return ErrEmptyId
	}

	course.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.Table()).
		Set("card_path", NilStr(course.CardPath)).
		Set("available", course.Available).
		Set("updated_at", FormatTime(course.UpdatedAt)).
		Where("id = ?", course.ID).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes a course based upon the where clause
//
// `tx` allows for the function to be run within a transaction
func (dao *CourseDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao)
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ClassifyPaths classifies the given paths into one of the following categories:
//   - PathClassificationNone: The path does not exist in the courses table
//   - PathClassificationAncestor: The path is an ancestor of a course path
//   - PathClassificationCourse: The path is an exact match to a course path
//   - PathClassificationDescendant: The path is a descendant of a course path
//
// The paths are returned as a map with the original path as the key and the classification as the
// value
func (dao *CourseDao) ClassifyPaths(paths []string) (map[string]types.PathClassification, error) {
	paths = slices.DeleteFunc(paths, func(s string) bool {
		return s == ""
	})

	if len(paths) == 0 {
		return nil, nil
	}

	// Initialize the results map
	results := make(map[string]types.PathClassification)
	for _, path := range paths {
		results[path] = types.PathClassificationNone
	}

	// Build the where clause
	whereClause := make([]squirrel.Sqlizer, len(paths))
	for i, path := range paths {
		whereClause[i] = squirrel.Like{dao.Table() + ".path": path + "%"}
	}

	query, args, _ := squirrel.
		StatementBuilder.
		Select(dao.Table() + ".path").
		From(dao.table).
		Where(squirrel.Or(whereClause)).
		ToSql()

	rows, err := dao.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Store the found course paths
	var coursePath string
	coursePaths := []string{}
	for rows.Next() {
		if err := rows.Scan(&coursePath); err != nil {
			return nil, err
		}
		coursePaths = append(coursePaths, coursePath)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Process
	for _, path := range paths {
		for _, coursePath := range coursePaths {
			if coursePath == path {
				results[path] = types.PathClassificationCourse
				break
			} else if strings.HasPrefix(coursePath, path) {
				results[path] = types.PathClassificationAncestor
				break
			} else if strings.HasPrefix(path, coursePath) && results[path] != types.PathClassificationAncestor {
				results[path] = types.PathClassificationDescendant
				break
			}
		}
	}

	return results, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ProcessOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon columns() for the current
// DAO. Additionally, it handles the special case of 'scan_status' column, which requires custom
// sorting logic, via a CASE statement.
//
// The custom sorting logic is defined as follows:
//   - NULL values are treated as the lowest value (sorted first in ASC, last in DESC)
//   - 'waiting' status is treated as the second value
//   - 'processing' status is treated as the third value
func (dao *CourseDao) ProcessOrderBy(orderBy []string) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	validTableColumns := dao.columns()
	var processedOrderBy []string

	scanDao := NewScanDao(dao.db)

	for _, ob := range orderBy {
		table, column := extractTableColumn(ob)

		if isValidOrderBy(table, column, validTableColumns) {
			// When the column is 'scan_status', apply the custom sorting logic
			if column == "scan_status" || table+"."+column == scanDao.Table()+".status" {
				// Determine the sort direction, defaulting to ASC if not specified
				parts := strings.Fields(ob)
				sortDirection := "ASC"
				if len(parts) > 1 {
					sortDirection = strings.ToUpper(parts[1])
				}

				caseStmt := "CASE " +
					"WHEN scan_status IS NULL THEN 1 " +
					"WHEN scan_status = 'waiting' THEN 2 " +
					"WHEN scan_status = 'processing' THEN 3 " +
					"END " + sortDirection

				processedOrderBy = append(processedOrderBy, caseStmt)
			} else {
				processedOrderBy = append(processedOrderBy, ob)
			}
		}
	}

	return processedOrderBy
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default select builder for counting
func (dao *CourseDao) countSelect() squirrel.SelectBuilder {
	sDao := NewScanDao(dao.db)
	cpDao := NewCourseProgressDao(dao.db)

	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		LeftJoin(sDao.Table() + " ON " + dao.Table() + ".id = " + sDao.Table() + ".course_id").
		LeftJoin(cpDao.Table() + " ON " + dao.Table() + ".id = " + cpDao.Table() + ".course_id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// It performs 2 left joins
//   - scans table to get `scan_status`
//   - courses progress table to get `started`, `started_at`, `percent`, and `completed_at`
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *CourseDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *CourseDao) columns() []string {
	sDao := NewScanDao(dao.db)
	cpDao := NewCourseProgressDao(dao.db)

	return []string{
		dao.Table() + ".*",
		sDao.Table() + ".status as scan_status",
		cpDao.Table() + ".started",
		cpDao.Table() + ".started_at",
		cpDao.Table() + ".percent",
		cpDao.Table() + ".completed_at",
		cpDao.Table() + ".updated_at as progress_updated_at",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a course
func (dao *CourseDao) data(c *models.Course) map[string]any {
	return map[string]any{
		"id":         c.ID,
		"title":      NilStr(c.Title),
		"path":       NilStr(c.Path),
		"card_path":  NilStr(c.CardPath),
		"available":  c.Available,
		"created_at": FormatTime(c.CreatedAt),
		"updated_at": FormatTime(c.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a course row
func (dao *CourseDao) scanRow(scannable Scannable) (*models.Course, error) {
	var c models.Course

	// Nullable fields
	var cardPath sql.NullString
	var scanStatus sql.NullString
	var createdAt string
	var updatedAt string
	var startedAt sql.NullString
	var completedAt sql.NullString
	var progressUpdatedAt sql.NullString

	err := scannable.Scan(
		// Course
		&c.ID,
		&c.Title,
		&c.Path,
		&cardPath,
		&c.Available,
		&createdAt,
		&updatedAt,
		// Scan
		&scanStatus,
		// Course progress
		&c.Started,
		&startedAt,
		&c.Percent,
		&completedAt,
		&progressUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if cardPath.Valid {
		c.CardPath = cardPath.String
	}

	if scanStatus.Valid {
		c.ScanStatus = scanStatus.String
	}

	if c.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if c.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	if c.StartedAt, err = ParseTimeNull(startedAt); err != nil {
		return nil, err
	}

	if c.CompletedAt, err = ParseTimeNull(completedAt); err != nil {
		return nil, err
	}

	if c.ProgressUpdatedAt, err = ParseTime(progressUpdatedAt.String); err != nil {
		return nil, err
	}

	return &c, nil
}
