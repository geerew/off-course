package dao

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourse inserts a new course record
func (dao *DAO) CreateCourse(ctx context.Context, course *models.Course) error {
	if course == nil {
		return utils.ErrNilPtr
	}

	if course.Title == "" {
		return utils.ErrTitle
	}

	if course.Path == "" {
		return utils.ErrPath
	}

	if course.ID == "" {
		course.RefreshId()
	}

	course.RefreshCreatedAt()
	course.RefreshUpdatedAt()

	// Ensure initial scan is false and maintenance is true
	course.InitialScan = false
	course.Maintenance = true

	builderOptions := newBuilderOptions(models.COURSE_TABLE).
		WithData(
			map[string]interface{}{
				models.BASE_ID:             course.ID,
				models.COURSE_TITLE:        course.Title,
				models.COURSE_PATH:         course.Path,
				models.COURSE_CARD_PATH:    course.CardPath,
				models.COURSE_AVAILABLE:    course.Available,
				models.COURSE_DURATION:     course.Duration,
				models.COURSE_INITIAL_SCAN: course.InitialScan,
				models.COURSE_MAINTENANCE:  course.Maintenance,
				models.BASE_CREATED_AT:     course.CreatedAt,
				models.BASE_UPDATED_AT:     course.UpdatedAt,
			},
		)

	return createGeneric(ctx, dao, *builderOptions)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourse gets a record from the courses table based upon the where clause in the options. If
// there is no where clause, it will return the first record in the table
//
// By default, progress is not included. Use `WithProgress()` on the options to include it
func (dao *DAO) GetCourse(ctx context.Context, dbOpts *database.Options) (*models.Course, error) {
	// When progress is not included, use a simpler query
	if dbOpts == nil || !dbOpts.IncludeProgress {
		builderOpts := newBuilderOptions(models.COURSE_TABLE).
			WithColumns(models.COURSE_TABLE + ".*").
			SetDbOpts(dbOpts).
			WithLimit(1)

		return getGeneric[models.Course](ctx, dao, *builderOpts)
	}

	// Include progress in the query
	principal, err := principalFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	builderOpts := newBuilderOptions(models.COURSE_TABLE).
		WithColumns(
			models.COURSE_TABLE+".*",
			fmt.Sprintf("%s AS course_started", models.COURSE_PROGRESS_TABLE_STARTED),
			fmt.Sprintf("%s AS course_started_at", models.COURSE_PROGRESS_TABLE_STARTED_AT),
			fmt.Sprintf("%s AS course_percent", models.COURSE_PROGRESS_TABLE_PERCENT),
			fmt.Sprintf("%s AS course_completed_at", models.COURSE_PROGRESS_TABLE_COMPLETED_AT),
		).
		WithLeftJoin(models.COURSE_PROGRESS_TABLE, fmt.Sprintf("%s = %s AND %s = '%s'", models.COURSE_PROGRESS_TABLE_COURSE_ID, models.COURSE_TABLE_ID, models.COURSE_PROGRESS_TABLE_USER_ID, principal.UserID)).
		SetDbOpts(dbOpts).
		WithLimit(1)

	row, err := getRow(ctx, dao, *builderOpts)
	if err != nil {
		return nil, err
	}

	course := &models.Course{}
	var (
		started   sql.NullBool
		startedAt types.DateTime
		percent   sql.NullInt64
		completed types.DateTime
	)

	err = row.Scan(
		&course.ID,
		&course.Title,
		&course.Path,
		&course.CardPath,
		&course.Available,
		&course.Duration,
		&course.InitialScan,
		&course.Maintenance,
		&course.CreatedAt,
		&course.UpdatedAt,
		// Progress
		&started,
		&startedAt,
		&percent,
		&completed,
	)

	if err != nil {
		return nil, err
	}

	// Attach progress
	//
	// When no progress is found, each field will be set to its zero value
	course.Progress = &models.CourseProgressInfo{
		Started:     started.Bool,
		StartedAt:   startedAt,
		Percent:     int(percent.Int64),
		CompletedAt: completed,
	}

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListCourses gets all records from the courses table based upon the where clause and pagination
// in the options
//
// By default, progress is not included. Use `WithProgress()` on the options to include it
func (dao *DAO) ListCourses(ctx context.Context, dbOpts *database.Options) ([]*models.Course, error) {
	// When progress is not included, use a simpler query
	if dbOpts == nil || !dbOpts.IncludeProgress {
		builderOpts := newBuilderOptions(models.COURSE_TABLE).
			WithColumns(models.COURSE_TABLE + ".*").
			SetDbOpts(dbOpts)

		return listGeneric[models.Course](ctx, dao, *builderOpts)
	}

	// Include progress in the query
	principal, err := principalFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	builderOpts := newBuilderOptions(models.COURSE_TABLE).
		WithColumns(
			models.COURSE_TABLE+".*",
			fmt.Sprintf("%s AS course_started", models.COURSE_PROGRESS_TABLE_STARTED),
			fmt.Sprintf("%s AS course_started_at", models.COURSE_PROGRESS_TABLE_STARTED_AT),
			fmt.Sprintf("%s AS course_percent", models.COURSE_PROGRESS_TABLE_PERCENT),
			fmt.Sprintf("%s AS course_completed_at", models.COURSE_PROGRESS_TABLE_COMPLETED_AT),
		).
		WithLeftJoin(models.COURSE_PROGRESS_TABLE, fmt.Sprintf("%s = %s AND %s = '%s'", models.COURSE_PROGRESS_TABLE_COURSE_ID, models.COURSE_TABLE_ID, models.COURSE_PROGRESS_TABLE_USER_ID, principal.UserID)).
		SetDbOpts(dbOpts)

	rows, err := getRows(ctx, dao, *builderOpts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []*models.Course
	for rows.Next() {
		var (
			course      models.Course
			started     sql.NullBool
			startedAt   types.DateTime
			percent     sql.NullInt64
			completedAt types.DateTime
		)

		err := rows.Scan(
			&course.ID,
			&course.Title,
			&course.Path,
			&course.CardPath,
			&course.Available,
			&course.Duration,
			&course.InitialScan,
			&course.Maintenance,
			&course.CreatedAt,
			&course.UpdatedAt,
			// Progress
			&started,
			&startedAt,
			&percent,
			&completedAt,
		)

		if err != nil {
			return nil, err
		}

		// Attach progress
		//
		// When no progress is found, each field will be set to its zero value
		course.Progress = &models.CourseProgressInfo{
			Started:     started.Bool,
			StartedAt:   startedAt,
			Percent:     int(percent.Int64),
			CompletedAt: completedAt,
		}

		courses = append(courses, &course)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourse updates a course record
func (dao *DAO) UpdateCourse(ctx context.Context, course *models.Course) error {
	if course == nil {
		return utils.ErrNilPtr
	}

	if course.ID == "" {
		return utils.ErrId
	}

	if course.Title == "" {
		return utils.ErrTitle
	}

	if course.Path == "" {
		return utils.ErrPath
	}

	course.RefreshUpdatedAt()

	dbOpts := &database.Options{
		Where: squirrel.Eq{models.BASE_ID: course.ID},
	}

	builderOptions := newBuilderOptions(models.COURSE_TABLE).
		WithData(
			map[string]interface{}{
				models.COURSE_TITLE:        course.Title,
				models.COURSE_PATH:         course.Path,
				models.COURSE_CARD_PATH:    course.CardPath,
				models.COURSE_AVAILABLE:    course.Available,
				models.COURSE_DURATION:     course.Duration,
				models.COURSE_INITIAL_SCAN: course.InitialScan,
				models.COURSE_MAINTENANCE:  course.Maintenance,
				models.BASE_UPDATED_AT:     course.UpdatedAt,
			},
		).
		SetDbOpts(dbOpts)

	_, err := updateGeneric(ctx, dao, *builderOptions)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteCourses deletes records from the courses table
//
// Errors when a where clause is not provided
func (dao *DAO) DeleteCourses(ctx context.Context, dbOpts *database.Options) error {
	if dbOpts == nil || dbOpts.Where == nil {
		return utils.ErrWhere
	}

	builderOpts := newBuilderOptions(models.COURSE_TABLE).SetDbOpts(dbOpts)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ClassifyCoursePaths classifies the given paths into one of the following categories:
//   - PathClassificationNone: The path does not exist in the courses table
//   - PathClassificationAncestor: The path is an ancestor of a course path
//   - PathClassificationCourse: The path is an exact match to a course path
//   - PathClassificationDescendant: The path is a descendant of a course path
//
// The paths are returned as a map with the original path as the key and the classification as the
// value
func (dao *DAO) ClassifyCoursePaths(ctx context.Context, paths []string) (map[string]types.PathClassification, error) {
	paths = slices.DeleteFunc(paths, func(s string) bool {
		return s == ""
	})

	if len(paths) == 0 {
		return nil, nil
	}

	results := make(map[string]types.PathClassification)
	for _, path := range paths {
		results[path] = types.PathClassificationNone
	}

	whereClause := make([]squirrel.Sqlizer, len(paths))
	for i, path := range paths {
		whereClause[i] = squirrel.Like{models.COURSE_TABLE_PATH: path + "%"}
	}

	query, args, _ := squirrel.
		StatementBuilder.
		Select(models.COURSE_TABLE_PATH).
		From(models.COURSE_TABLE).
		Where(squirrel.Or(whereClause)).
		ToSql()

	q := database.QuerierFromContext(ctx, dao.db)
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
