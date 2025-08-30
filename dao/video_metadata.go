package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateVideoMetadata inserts a new video metadata record
func (dao *DAO) CreateVideoMetadata(ctx context.Context, metadata *models.VideoMetadata) error {
	if metadata == nil {
		return utils.ErrNilPtr
	}

	if metadata.ID == "" {
		metadata.RefreshId()
	}

	metadata.RefreshCreatedAt()
	metadata.RefreshUpdatedAt()

	builderOptions := newBuilderOptions(models.VIDEO_METADATA_TABLE).
		WithData(
			map[string]interface{}{
				models.BASE_ID:                   metadata.ID,
				models.VIDEO_METADATA_ASSET_ID:   metadata.AssetID,
				models.VIDEO_METADATA_DURATION:   metadata.Duration,
				models.VIDEO_METADATA_WIDTH:      metadata.Width,
				models.VIDEO_METADATA_HEIGHT:     metadata.Height,
				models.VIDEO_METADATA_RESOLUTION: metadata.Resolution,
				models.VIDEO_METADATA_CODEC:      metadata.Codec,
				models.BASE_CREATED_AT:           metadata.CreatedAt,
				models.BASE_UPDATED_AT:           metadata.UpdatedAt,
			},
		)

	return createGeneric(ctx, dao, *builderOptions)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoMetadata gets a record from the video metadata table based upon the where clause in
// the options. If there is no where clause, it will return the first record in the table
func (dao *DAO) GetVideoMetadata(ctx context.Context, dbOpts *database.Options) (*models.VideoMetadata, error) {
	builderOpts := newBuilderOptions(models.VIDEO_METADATA_TABLE).
		WithColumns(
			models.VIDEO_METADATA_TABLE + ".*",
		).
		SetDbOpts(dbOpts).
		WithLimit(1)

	return getGeneric[models.VideoMetadata](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListVideoMetadata gets all records from the video metadata table based upon the where clause and
// pagination in the options
func (dao *DAO) ListVideoMetadata(ctx context.Context, dbOpts *database.Options) ([]*models.VideoMetadata, error) {
	builderOpts := newBuilderOptions(models.VIDEO_METADATA_TABLE).
		WithColumns(
			models.VIDEO_METADATA_TABLE + ".*",
		).
		SetDbOpts(dbOpts)

	return listGeneric[models.VideoMetadata](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateVideoMetadata updates a video metadata record
func (dao *DAO) UpdateVideoMetadata(ctx context.Context, metadata *models.VideoMetadata) error {
	if metadata == nil {
		return utils.ErrNilPtr
	}

	if metadata.ID == "" {
		return utils.ErrId
	}

	metadata.RefreshUpdatedAt()

	dbOpts := &database.Options{
		Where: squirrel.Eq{models.BASE_ID: metadata.ID},
	}

	builderOptions := newBuilderOptions(models.VIDEO_METADATA_TABLE).
		WithData(
			map[string]interface{}{
				models.VIDEO_METADATA_DURATION:   metadata.Duration,
				models.VIDEO_METADATA_WIDTH:      metadata.Width,
				models.VIDEO_METADATA_HEIGHT:     metadata.Height,
				models.VIDEO_METADATA_RESOLUTION: metadata.Resolution,
				models.VIDEO_METADATA_CODEC:      metadata.Codec,
				models.BASE_UPDATED_AT:           metadata.UpdatedAt,
			},
		).
		SetDbOpts(dbOpts)

	_, err := updateGeneric(ctx, dao, *builderOptions)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteVideoMetadata deletes records from the video metadata table
//
// Errors when a where clause is not provided
func (dao *DAO) DeleteVideoMetadata(ctx context.Context, dbOpts *database.Options) error {
	if dbOpts == nil || dbOpts.Where == nil {
		return utils.ErrWhere
	}

	builderOpts := newBuilderOptions(models.VIDEO_METADATA_TABLE).SetDbOpts(dbOpts)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}
