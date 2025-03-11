package dao

import (
	"context"

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

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		if assetProgress.VideoPos < 0 {
			assetProgress.VideoPos = 0
		}

		options := &database.Options{}

		// Join the course table
		options.AdditionalJoins = append(
			options.AdditionalJoins,
			models.COURSE_TABLE+" ON "+models.ASSET_TABLE_COURSE_ID+" = "+models.COURSE_TABLE_ID,
		)

		options.Where = squirrel.And{
			squirrel.Eq{models.ASSET_TABLE_ID: assetProgress.AssetID},
			squirrel.Eq{models.COURSE_TABLE_ID: courseId},
		}

		asset := &models.Asset{}
		err := dao.Get(
			txCtx,
			asset,
			options,
		)

		if err != nil {
			return err
		}

		if asset.Progress == nil {
			// Create
			if assetProgress.Completed {
				assetProgress.CompletedAt = types.NowDateTime()
			}

			err := dao.Create(txCtx, assetProgress)
			if err != nil {
				return err
			}
		} else {
			// Update
			assetProgress.ID = asset.Progress.ID
			if assetProgress.Completed {
				if asset.Progress.Completed {
					assetProgress.CompletedAt = asset.Progress.CompletedAt
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

		// Refresh course progress
		return dao.RefreshCourseProgress(txCtx, asset.CourseID)
	})
}
