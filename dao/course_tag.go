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
			return dao.Create(txCtx, courseTag)
		}

		// Get the tag by tag name (case-insensitive)
		tag := models.Tag{}
		options := &database.Options{
			Where: squirrel.Eq{fmt.Sprintf("%s.%s", models.TAG_TABLE, models.TAG_TAG): courseTag.Tag},
		}

		err := dao.Get(txCtx, &tag, options)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// If the tag does not exist, create it
		if err == sql.ErrNoRows {
			tag.Tag = courseTag.Tag
			err = dao.Create(txCtx, &tag)
			if err != nil {
				return err
			}
		}

		courseTag.TagID = tag.ID

		return dao.Create(txCtx, courseTag)

	})
}
