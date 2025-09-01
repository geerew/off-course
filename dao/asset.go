package dao

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAsset inserts a new asset record
func (dao *DAO) CreateAsset(ctx context.Context, asset *models.Asset) error {

	if err := assetValidation(asset); err != nil {
		return err
	}

	if asset.ID == "" {
		asset.RefreshId()
	}

	asset.RefreshCreatedAt()
	asset.RefreshUpdatedAt()

	builderOpts := newBuilderOptions(models.ASSET_TABLE).
		WithData(
			map[string]interface{}{
				models.BASE_ID:              asset.ID,
				models.ASSET_COURSE_ID:      asset.CourseID,
				models.ASSET_ASSET_GROUP_ID: asset.AssetGroupID,
				models.ASSET_TITLE:          asset.Title,
				models.ASSET_PREFIX:         asset.Prefix,
				models.ASSET_SUB_PREFIX:     asset.SubPrefix,
				models.ASSET_SUB_TITLE:      asset.SubTitle,
				models.ASSET_MODULE:         asset.Module,
				models.ASSET_TYPE:           asset.Type,
				models.ASSET_PATH:           asset.Path,
				models.ASSET_FILE_SIZE:      asset.FileSize,
				models.ASSET_MOD_TIME:       asset.ModTime,
				models.ASSET_HASH:           asset.Hash,
				models.BASE_CREATED_AT:      asset.CreatedAt,
				models.BASE_UPDATED_AT:      asset.UpdatedAt,
			},
		)

	return createGeneric(ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountAssets counts the number of asset records
func (dao *DAO) CountAssets(ctx context.Context, dbOpts *database.Options) (int, error) {
	builderOpts := newBuilderOptions(models.ASSET_TABLE).SetDbOpts(dbOpts)
	return countGeneric(ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAsset gets a record from the assets table based upon the where clause in the options. If
// there is no where clause, it will return the first record in the table
//
// By default, progress is not included. Use `WithProgress()` on the options to include it
// By default, video metadata is not included. Use `WithAssetVideoMetadata()` on the options to include it
func (dao *DAO) GetAsset(ctx context.Context, dbOpts *database.Options) (*models.Asset, error) {
	builderOpts := newBuilderOptions(models.ASSET_TABLE).
		WithColumns(models.ASSET_TABLE + ".*").
		SetDbOpts(dbOpts).
		WithLimit(1)

	// When relations are not included, use a simpler query
	if dbOpts == nil || (!dbOpts.IncludeProgress && !dbOpts.IncludeAssetVideoMetadata) {
		return getGeneric[models.Asset](ctx, dao, *builderOpts)
	}

	// Add the progress columns and join
	if dbOpts.IncludeProgress {
		principal, err := principalFromCtx(ctx)
		if err != nil {
			return nil, err
		}

		builderOpts = builderOpts.
			WithColumns(models.AssetProgressJoinColumns()...).
			WithLeftJoin(models.ASSET_PROGRESS_TABLE, fmt.Sprintf("%s = %s AND %s = '%s'", models.ASSET_PROGRESS_TABLE_ASSET_ID, models.ASSET_TABLE_ID, models.ASSET_PROGRESS_TABLE_USER_ID, principal.UserID))
	}

	// Add the asset metadata columns and join
	if dbOpts.IncludeAssetVideoMetadata {
		builderOpts = builderOpts.
			WithColumns(models.VideoMetadataJoinColumns()...).
			WithLeftJoin(models.VIDEO_METADATA_TABLE, fmt.Sprintf("%s = %s", models.VIDEO_METADATA_TABLE_ASSET_ID, models.ASSET_TABLE_ID))
	}

	row, err := getRow(ctx, dao, *builderOpts)
	if err != nil {
		return nil, err
	}

	asset := &models.Asset{}
	scanTargets := []interface{}{
		&asset.ID,
		&asset.CourseID,
		&asset.AssetGroupID,
		&asset.Title,
		&asset.Prefix,
		&asset.SubPrefix,
		&asset.SubTitle,
		&asset.Module,
		&asset.Type,
		&asset.Path,
		&asset.FileSize,
		&asset.ModTime,
		&asset.Hash,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	}

	var (
		// Progress
		videoPos    sql.NullInt64
		completed   sql.NullBool
		completedAt types.DateTime
		// Metadata
		duration sql.NullInt64
		width    sql.NullInt64
		height   sql.NullInt64
		res      sql.NullString
		codec    sql.NullString
	)

	if dbOpts.IncludeProgress {
		scanTargets = append(scanTargets,
			&videoPos,
			&completed,
			&completedAt,
		)
	}

	if dbOpts.IncludeAssetVideoMetadata {
		scanTargets = append(scanTargets,
			&duration,
			&width,
			&height,
			&res,
			&codec,
		)
	}

	if err = row.Scan(scanTargets...); err != nil {
		return nil, err
	}

	// Attach progress
	if dbOpts.IncludeProgress {
		asset.Progress = &models.AssetProgressInfo{
			VideoPos:    int(videoPos.Int64),
			Completed:   completed.Bool,
			CompletedAt: completedAt,
		}
	}

	// Attach video metadata
	if asset.Type.IsVideo() && dbOpts.IncludeAssetVideoMetadata {
		asset.VideoMetadata = &models.VideoMetadataInfo{
			Duration:   int(duration.Int64),
			Width:      int(width.Int64),
			Height:     int(height.Int64),
			Resolution: res.String,
			Codec:      codec.String,
		}
	}

	return asset, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListAssets gets all records from the assets table based upon the where clause and pagination
// in the options
//
// By default, progress is not included. Use `WithProgress()` on the options to include it
func (dao *DAO) ListAssets(ctx context.Context, dbOpts *database.Options) ([]*models.Asset, error) {
	builderOpts := newBuilderOptions(models.ASSET_TABLE).
		WithColumns(models.ASSET_TABLE + ".*").
		SetDbOpts(dbOpts)

	// When relations are not included, use a simpler query
	if dbOpts == nil || (!dbOpts.IncludeProgress && !dbOpts.IncludeAssetVideoMetadata) {
		return listGeneric[models.Asset](ctx, dao, *builderOpts)
	}

	// Add the progress columns and join
	if dbOpts.IncludeProgress {
		principal, err := principalFromCtx(ctx)
		if err != nil {
			return nil, err
		}

		builderOpts = builderOpts.
			WithColumns(models.AssetProgressJoinColumns()...).
			WithLeftJoin(models.ASSET_PROGRESS_TABLE, fmt.Sprintf("%s = %s AND %s = '%s'", models.ASSET_PROGRESS_TABLE_ASSET_ID, models.ASSET_TABLE_ID, models.ASSET_PROGRESS_TABLE_USER_ID, principal.UserID))
	}

	// Add the asset metadata columns and join
	if dbOpts.IncludeAssetVideoMetadata {
		builderOpts = builderOpts.
			WithColumns(models.VideoMetadataJoinColumns()...).
			WithLeftJoin(models.VIDEO_METADATA_TABLE, fmt.Sprintf("%s = %s", models.VIDEO_METADATA_TABLE_ASSET_ID, models.ASSET_TABLE_ID))
	}

	rows, err := getRows(ctx, dao, *builderOpts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*models.Asset
	for rows.Next() {

		asset := &models.Asset{}
		scanTargets := []interface{}{
			&asset.ID,
			&asset.CourseID,
			&asset.AssetGroupID,
			&asset.Title,
			&asset.Prefix,
			&asset.SubPrefix,
			&asset.SubTitle,
			&asset.Module,
			&asset.Type,
			&asset.Path,
			&asset.FileSize,
			&asset.ModTime,
			&asset.Hash,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		}

		var (
			// Progress
			videoPos    sql.NullInt64
			completed   sql.NullBool
			completedAt types.DateTime
			// Metadata
			duration sql.NullInt64
			width    sql.NullInt64
			height   sql.NullInt64
			res      sql.NullString
			codec    sql.NullString
		)

		if dbOpts.IncludeProgress {
			scanTargets = append(scanTargets,
				&videoPos,
				&completed,
				&completedAt,
			)
		}
		if dbOpts.IncludeAssetVideoMetadata {
			scanTargets = append(scanTargets,
				&duration,
				&width,
				&height,
				&res,
				&codec,
			)
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return nil, err
		}

		// Attach progress
		//
		// When no progress is found, each field will be set to its zero value
		if dbOpts.IncludeProgress {
			asset.Progress = &models.AssetProgressInfo{
				VideoPos:    int(videoPos.Int64),
				Completed:   completed.Bool,
				CompletedAt: completedAt,
			}
		}

		// Attach video metadata
		if asset.Type.IsVideo() && dbOpts.IncludeAssetVideoMetadata {
			asset.VideoMetadata = &models.VideoMetadataInfo{
				Duration:   int(duration.Int64),
				Width:      int(width.Int64),
				Height:     int(height.Int64),
				Resolution: res.String,
				Codec:      codec.String,
			}
		}

		assets = append(assets, asset)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return assets, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAsset updates an asset record
func (dao *DAO) UpdateAsset(ctx context.Context, asset *models.Asset) error {
	if err := assetValidation(asset); err != nil {
		return err
	}

	if asset.ID == "" {
		return utils.ErrId
	}

	asset.RefreshUpdatedAt()

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.BASE_ID: asset.ID})

	builderOpts := newBuilderOptions(models.ASSET_TABLE).
		WithData(
			map[string]interface{}{
				models.ASSET_ASSET_GROUP_ID: asset.AssetGroupID,
				models.ASSET_TITLE:          asset.Title,
				models.ASSET_PREFIX:         asset.Prefix,
				models.ASSET_SUB_PREFIX:     asset.SubPrefix,
				models.ASSET_SUB_TITLE:      asset.SubTitle,
				models.ASSET_MODULE:         asset.Module,
				models.ASSET_TYPE:           asset.Type,
				models.ASSET_PATH:           asset.Path,
				models.ASSET_FILE_SIZE:      asset.FileSize,
				models.ASSET_MOD_TIME:       asset.ModTime,
				models.ASSET_HASH:           asset.Hash,
				models.BASE_UPDATED_AT:      asset.UpdatedAt,
			},
		).
		SetDbOpts(dbOpts)

	_, err := updateGeneric(ctx, dao, *builderOpts)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAssets deletes records from the assets table
//
// Errors when a where clause is not provided
func (dao *DAO) DeleteAssets(ctx context.Context, dbOpts *database.Options) error {
	if dbOpts == nil || dbOpts.Where == nil {
		return utils.ErrWhere
	}

	builderOpts := newBuilderOptions(models.ASSET_TABLE).SetDbOpts(dbOpts)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// assetValidation validates the asset fields
func assetValidation(asset *models.Asset) error {
	if asset == nil {
		return utils.ErrNilPtr
	}

	if asset.CourseID == "" {
		return utils.ErrCourseId
	}

	if asset.AssetGroupID == "" {
		return utils.ErrAssetGroupId
	}

	if asset.Title == "" {
		return utils.ErrTitle
	}

	if !asset.Prefix.Valid || asset.Prefix.Int16 < 0 {
		return utils.ErrPrefix
	}

	if asset.Path == "" {
		return utils.ErrPath
	}

	return nil
}
