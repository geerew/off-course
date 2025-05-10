package dao

import (
	"context"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourse creates a course and course progress
func (dao *DAO) CreateCourse(ctx context.Context, course *models.Course) error {
	if course == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, course)

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourse retrieves a course
//
// When options is nil or options.Where is nil, the function will use the ID to filter courses
func (dao *DAO) GetCourse(ctx context.Context, course *models.Course, options *database.Options) error {
	if course == nil {
		return utils.ErrNilPtr
	}

	userId, ok := ctx.Value(types.UserContextKey).(string)
	if !ok || userId == "" {
		return utils.ErrMissingUserId
	}

	if options == nil {
		options = &database.Options{}
	}

	if options.Where == nil {
		if course.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{Where: squirrel.Eq{course.Table() + "." + models.BASE_ID: course.Id()}}
	}

	options.AddRelationFilter("Progress", models.COURSE_PROGRESS_USER_ID, userId)

	return dao.Get(ctx, course, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListCourses retrieves a list of courses
func (dao *DAO) ListCourses(ctx context.Context, courses *[]*models.Course, options *database.Options) error {
	if courses == nil {
		return utils.ErrNilPtr
	}

	userId, ok := ctx.Value(types.UserContextKey).(string)
	if !ok || userId == "" {
		return utils.ErrMissingUserId
	}

	if options == nil {
		options = &database.Options{}
	}

	options.AddRelationFilter("Progress", models.COURSE_PROGRESS_USER_ID, userId)

	return dao.List(ctx, courses, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourse updates a course
func (dao *DAO) UpdateCourse(ctx context.Context, course *models.Course) error {
	if course == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, course)
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
	course := &models.Course{}

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
		From(course.Table()).
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
