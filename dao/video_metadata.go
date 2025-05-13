package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateVideoMetadata creates video metadata for an asset
// TODO Change to CreateOrReplace
func (dao *DAO) CreateVideoMetadata(ctx context.Context, metadata *models.VideoMetadata) error {
	if metadata == nil {
		return utils.ErrNilPtr
	}

	return Create(ctx, dao, metadata)

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoMetadata retrieves a video metadata
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetVideoMetadata(ctx context.Context, videoMetadata *models.VideoMetadata, options *database.Options) error {
	if videoMetadata == nil {
		return utils.ErrNilPtr
	}

	if options == nil || options.Where == nil {
		if videoMetadata.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{
			Where: squirrel.Eq{models.VIDEO_METADATA_TABLE_ID: videoMetadata.Id()},
		}
	}

	if options.Where == nil {
	}

	return Get(ctx, dao, videoMetadata, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListVideoMetadata retrieves a list of videoMetadata
func (dao *DAO) ListVideoMetadata(ctx context.Context, videoMetadata *[]*models.VideoMetadata, options *database.Options) error {
	if videoMetadata == nil {
		return utils.ErrNilPtr
	}

	return List(ctx, dao, videoMetadata, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateVideoMetadata updates video metadata for an asset
func (dao *DAO) UpdateVideoMetadata(ctx context.Context, metadata *models.VideoMetadata) error {
	if metadata == nil {
		return utils.ErrNilPtr
	}

	_, err := Update(ctx, dao, metadata)
	return err
}
