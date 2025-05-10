package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAttachment creates an attachment
func (dao *DAO) CreateAttachment(ctx context.Context, attachment *models.Attachment) error {
	if attachment == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, attachment)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAttachment updates an attachment
func (dao *DAO) UpdateAttachment(ctx context.Context, attachment *models.Attachment) error {
	if attachment == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, attachment)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAttachment retrieves an attachment
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetAttachment(ctx context.Context, attachment *models.Attachment, options *database.Options) error {
	if attachment == nil {
		return utils.ErrNilPtr
	}

	if options == nil || options.Where == nil {
		if attachment.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{
			Where: squirrel.Eq{attachment.Table() + "." + models.BASE_ID: attachment.Id()},
		}
	}

	if options.Where == nil {
	}

	return dao.Get(ctx, attachment, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListAttachments retrieves a list of attachments
func (dao *DAO) ListAttachments(ctx context.Context, attachments *[]*models.Attachment, options *database.Options) error {
	if attachments == nil {
		return utils.ErrNilPtr
	}

	return dao.List(ctx, attachments, options)
}
