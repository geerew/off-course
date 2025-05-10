package dao

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateOrUpdateAssetProgress creates/updates an asset progress and refreshes course progress
func (dao *DAO) CreateOrUpdateAssetProgress(ctx context.Context, courseId string, assetProgress *models.AssetProgress) error {
	if assetProgress == nil {
		return utils.ErrNilPtr
	}

	// Extract user ID from context
	userId, ok := ctx.Value(types.UserContextKey).(string)
	if !ok || userId == "" {
		return utils.ErrMissingUserId
	}

	// Set the user ID in the progress object
	assetProgress.UserID = userId

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		if assetProgress.VideoPos < 0 {
			assetProgress.VideoPos = 0
		}

		options := &database.Options{}

		// Join the course table
		options.AddJoin(
			models.COURSE_TABLE, models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID,
		)

		options.Where = squirrel.And{
			squirrel.Eq{models.ASSET_TABLE_ID: assetProgress.AssetID},
			squirrel.Eq{models.COURSE_TABLE_ID: courseId},
		}

		asset := &models.Asset{}
		err := dao.GetAsset(txCtx, asset, options)
		if err != nil {
			return err
		}

		// Use both asset_id and user_id to look up the existing progress
		existingProgress := &models.AssetProgress{}
		err = dao.GetAssetProgress(txCtx, existingProgress, &database.Options{
			Where: squirrel.And{
				squirrel.Eq{models.ASSET_PROGRESS_TABLE_ASSET_ID: assetProgress.AssetID},
				squirrel.Eq{models.ASSET_PROGRESS_TABLE_USER_ID: userId},
			},
		})

		if err != nil && err != sql.ErrNoRows {
			return err
		}

		if err == sql.ErrNoRows {
			// Create new progress record
			if assetProgress.Completed {
				assetProgress.CompletedAt = types.NowDateTime()
			}

			err := dao.Create(txCtx, assetProgress)
			if err != nil {
				return err
			}
		} else {
			// Update existing progress
			assetProgress.ID = existingProgress.ID
			if assetProgress.Completed {
				if existingProgress.Completed {
					assetProgress.CompletedAt = existingProgress.CompletedAt
				} else {
					assetProgress.CompletedAt = types.NowDateTime()
				}
			} else {
				assetProgress.CompletedAt = types.DateTime{}
			}

			_, err = dao.Update(txCtx, assetProgress)
			if err != nil {
				return err
			}
		}

		// Pass user ID to RefreshCourseProgress
		return dao.RefreshCourseProgress(txCtx, asset.CourseID)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssetProgress retrieves an asset progress
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetAssetProgress(ctx context.Context, assetProgress *models.AssetProgress, options *database.Options) error {
	if assetProgress == nil {
		return utils.ErrNilPtr
	}

	if options == nil || options.Where == nil {
		if assetProgress.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{
			Where: squirrel.Eq{assetProgress.Table() + "." + models.BASE_ID: assetProgress.Id()},
		}
	}

	if options.Where == nil {
	}

	return dao.Get(ctx, assetProgress, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListAssetProgress retrieves a list of asset progress
func (dao *DAO) ListAssetProgress(ctx context.Context, assetProgress *[]*models.AssetProgress, options *database.Options) error {
	if assetProgress == nil {
		return utils.ErrNilPtr
	}

	return dao.List(ctx, assetProgress, options)
}
