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

// CreateTag creates a tag
func (dao *DAO) CreateTag(ctx context.Context, tag *models.Tag) error {
	if tag == nil {
		return utils.ErrNilPtr
	}

	// Check if the tag already exists
	options := &database.Options{
		Where: squirrel.Expr(
			fmt.Sprintf("LOWER(%s) = LOWER(?)", models.TAG_TABLE_TAG),
			tag.Tag,
		),
	}
	existingTag := &models.Tag{}
	err := dao.GetTag(ctx, existingTag, options)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// The tag already exists, update the tag with the existing tag and attempt to create it. This
	// will result in an error but it gives a more specific error message
	if err == nil {
		tag.Tag = existingTag.Tag
	}

	return Create(ctx, dao, tag)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetTag retrieves a tag
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetTag(ctx context.Context, tag *models.Tag, options *database.Options) error {
	if tag == nil {
		return utils.ErrNilPtr
	}

	if options == nil || options.Where == nil {
		if tag.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{
			Where: squirrel.Eq{models.TAG_TABLE_ID: tag.Id()},
		}
	}

	if options.Where == nil {
	}

	return Get(ctx, dao, tag, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListTags retrieves a list of tags
func (dao *DAO) ListTags(ctx context.Context, tags *[]*models.Tag, options *database.Options) error {
	if tags == nil {
		return utils.ErrNilPtr
	}

	return List(ctx, dao, tags, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateTag updates a tag
func (dao *DAO) UpdateTag(ctx context.Context, tag *models.Tag) error {
	if tag == nil {
		return utils.ErrNilPtr
	}

	_, err := Update(ctx, dao, tag)
	return err
}
