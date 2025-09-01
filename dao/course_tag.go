package dao

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourseTag inserts a new course tag record
func (dao *DAO) CreateCourseTag(ctx context.Context, courseTag *models.CourseTag) error {
	if courseTag == nil {
		return utils.ErrNilPtr
	}

	if courseTag.CourseID == "" {
		return utils.ErrCourseId
	}

	if courseTag.ID == "" {
		courseTag.RefreshId()
	}

	courseTag.RefreshCreatedAt()
	courseTag.RefreshUpdatedAt()

	courseTagData := map[string]interface{}{
		models.BASE_ID:              courseTag.ID,
		models.COURSE_TAG_COURSE_ID: courseTag.CourseID,
		models.BASE_CREATED_AT:      courseTag.CreatedAt,
		models.BASE_UPDATED_AT:      courseTag.UpdatedAt,
	}

	// When the tag ID is set, we can just create the course tag
	if courseTag.TagID != "" {
		courseTagData[models.COURSE_TAG_TAG_ID] = courseTag.TagID
		builderOpts := newBuilderOptions(models.COURSE_TAG_TABLE).WithData(courseTagData)
		return createGeneric(ctx, dao, *builderOpts)
	}

	// If the tag ID is not set, we need to find the tag by name
	if courseTag.Tag == "" {
		return utils.ErrTag
	}

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.TAG_TABLE_TAG: courseTag.Tag})

		tag, err := dao.GetTag(txCtx, dbOpts)
		if err != nil {
			return err
		}

		// If the tag does not exist, create it
		if tag == nil {
			tag = &models.Tag{Tag: courseTag.Tag}
			err = dao.CreateTag(txCtx, tag)
			if err != nil {
				return err
			}
		}

		courseTag.TagID = tag.ID

		courseTagData[models.COURSE_TAG_TAG_ID] = courseTag.TagID
		builderOpts := newBuilderOptions(models.COURSE_TAG_TABLE).WithData(courseTagData)
		return createGeneric(txCtx, dao, *builderOpts)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourseTag gets a record from the course tags table based upon the where clause in the options. If
// there is no where clause, it will return the first record in the table
func (dao *DAO) GetCourseTag(ctx context.Context, dbOpts *database.Options) (*models.CourseTag, error) {
	builderOpts := newBuilderOptions(models.COURSE_TAG_TABLE).
		WithColumns(
			models.COURSE_TAG_TABLE+".*",
			models.COURSE_TABLE_TITLE+" AS course_title",
			models.TAG_TABLE_TAG+" AS tag_tag",
		).
		WithJoin(models.COURSE_TABLE, fmt.Sprintf("%s = %s", models.COURSE_TABLE_ID, models.COURSE_TAG_TABLE_COURSE_ID)).
		WithJoin(models.TAG_TABLE, fmt.Sprintf("%s = %s", models.TAG_TABLE_ID, models.COURSE_TAG_TABLE_TAG_ID)).
		SetDbOpts(dbOpts).
		WithLimit(1)

	return getGeneric[models.CourseTag](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListCourseTags gets all records from the course tags table based upon the where clause and pagination
// in the options
func (dao *DAO) ListCourseTags(ctx context.Context, dbOpts *database.Options) ([]*models.CourseTag, error) {
	builderOpts := newBuilderOptions(models.COURSE_TAG_TABLE).
		WithColumns(
			models.COURSE_TAG_TABLE+".*",
			models.COURSE_TABLE_TITLE+" AS course_title",
			models.TAG_TABLE_TAG+" AS tag_tag",
		).
		WithJoin(models.COURSE_TABLE, fmt.Sprintf("%s = %s", models.COURSE_TABLE_ID, models.COURSE_TAG_TABLE_COURSE_ID)).
		WithJoin(models.TAG_TABLE, fmt.Sprintf("%s = %s", models.TAG_TABLE_ID, models.COURSE_TAG_TABLE_TAG_ID)).
		SetDbOpts(dbOpts)

	return listGeneric[models.CourseTag](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteCourseTags deletes records from the course tags table
//
// Errors when a where clause is not provided
func (dao *DAO) DeleteCourseTags(ctx context.Context, dbOpts *database.Options) error {
	if dbOpts == nil || dbOpts.Where == nil {
		return utils.ErrWhere
	}

	builderOpts := newBuilderOptions(models.COURSE_TAG_TABLE).SetDbOpts(dbOpts)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}
