package dao

import (
	"context"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateVideoMetadata creates video metadata for an asset
func (dao *DAO) CreateVideoMetadata(ctx context.Context, metadata *models.VideoMetadata) error {
	if metadata == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, metadata)

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateVideoMetadata updates video metadata for an asset
func (dao *DAO) UpdateVideoMetadata(ctx context.Context, metadata *models.VideoMetadata) error {
	if metadata == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, metadata)
	return err
}
