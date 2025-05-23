package dao

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourseTag creates a course tag
func (dao *DAO) CreateCourseTag(ctx context.Context, courseTag *models.CourseTag) error {
	if courseTag == nil {
		return utils.ErrNilPtr
	}

	if courseTag.TagID == "" && courseTag.Tag == "" {
		return fmt.Errorf("tag ID and tag cannot be empty")
	}

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		if courseTag.TagID != "" {
			return Create(txCtx, dao, courseTag)
		}

		// Get the tag by tag name (case-insensitive)
		tag := models.Tag{}
		options := &database.Options{
			Where: squirrel.Eq{models.TAG_TABLE_TAG: courseTag.Tag},
		}

		err := dao.GetTag(txCtx, &tag, options)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// If the tag does not exist, create it
		if err == sql.ErrNoRows {
			tag.Tag = courseTag.Tag
			err = dao.CreateTag(txCtx, &tag)
			if err != nil {
				return err
			}
		}

		courseTag.TagID = tag.ID

		return Create(txCtx, dao, courseTag)

	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourseTag retrieves a course tag
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetCourseTag(ctx context.Context, courseTag *models.CourseTag, options *database.Options) error {
	if courseTag == nil {
		return utils.ErrNilPtr
	}

	if options == nil {
		options = &database.Options{}
	}

	if options.Where == nil {
		if courseTag.Id() == "" {
			return utils.ErrInvalidId
		}

		options.Where = squirrel.Eq{models.COURSE_TAG_TABLE_ID: courseTag.Id()}
	}

	if options.Where == nil {
	}

	return Get(ctx, dao, courseTag, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListCourseTags retrieves a list of course tags
func (dao *DAO) ListCourseTags(ctx context.Context, courseTags *[]*models.CourseTag, options *database.Options) error {
	if courseTags == nil {
		return utils.ErrNilPtr
	}

	return List(ctx, dao, courseTags, options)
}
