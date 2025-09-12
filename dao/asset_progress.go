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

// UpsertAssetProgress creates or updates an asset progress record for a user
//
// TODO rewrite to use a single method to use withSuffix and ON CONFLICT (like course_progress)
func (dao *DAO) UpsertAssetProgress(ctx context.Context, courseID string, assetProgress *models.AssetProgress) error {
	if assetProgress == nil {
		return utils.ErrNilPtr
	}

	principal, err := principalFromCtx(ctx)
	if err != nil {
		return err
	}
	assetProgress.UserID = principal.UserID

	if assetProgress.Position < 0 {
		assetProgress.Position = 0
	}

	// Get the existing asset, ensuring it belongs to the course
	dbOpts := database.NewOptions().
		WithJoin(models.COURSE_TABLE, fmt.Sprintf("%s = %s", models.ASSET_TABLE_COURSE_ID, models.COURSE_TABLE_ID)).
		WithWhere(squirrel.And{
			squirrel.Eq{models.ASSET_TABLE_ID: assetProgress.AssetID},
			squirrel.Eq{models.COURSE_TABLE_ID: courseID},
		})

	asset := &models.Asset{}
	asset, err = dao.GetAsset(ctx, dbOpts)
	if err != nil {
		return err
	}

	if asset == nil {
		return sql.ErrNoRows
	}

	// Get the existing asset progress if it exists
	dbOpts = database.NewOptions().WithWhere(squirrel.And{
		squirrel.Eq{models.ASSET_PROGRESS_TABLE_ASSET_ID: assetProgress.AssetID},
		squirrel.Eq{models.ASSET_PROGRESS_TABLE_USER_ID: assetProgress.UserID},
	})

	existing, err := dao.GetAssetProgress(ctx, dbOpts)
	if err != nil {
		return err
	}

	now := types.NowDateTime()

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		if existing == nil {
			assetProgress.RefreshId()
			assetProgress.RefreshCreatedAt()
			assetProgress.RefreshUpdatedAt()

			if assetProgress.Completed {
				assetProgress.CompletedAt = now
			}

			builderOpts := newBuilderOptions(models.ASSET_PROGRESS_TABLE).
				WithData(map[string]interface{}{
					models.BASE_ID:                      assetProgress.ID,
					models.ASSET_PROGRESS_ASSET_ID:      assetProgress.AssetID,
					models.ASSET_PROGRESS_USER_ID:       assetProgress.UserID,
					models.ASSET_PROGRESS_POSITION:      assetProgress.Position,
					models.ASSET_PROGRESS_PROGRESS_FRAC: assetProgress.ProgressFrac,
					models.ASSET_PROGRESS_COMPLETED:     assetProgress.Completed,
					models.ASSET_PROGRESS_COMPLETED_AT:  assetProgress.CompletedAt,
					models.BASE_CREATED_AT:              assetProgress.CreatedAt,
					models.BASE_UPDATED_AT:              assetProgress.UpdatedAt,
				})

			if err := createGeneric(txCtx, dao, *builderOpts); err != nil {
				return err
			}
		} else {
			assetProgress.ID = existing.ID

			// Only bump completed_at when first flipping to true
			if assetProgress.Completed {
				if existing.Completed {
					assetProgress.CompletedAt = existing.CompletedAt
				} else {
					assetProgress.CompletedAt = now
				}
			} else {
				assetProgress.CompletedAt = types.DateTime{}
			}

			assetProgress.RefreshUpdatedAt()

			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.BASE_ID: assetProgress.ID})

			builderOpts := newBuilderOptions(models.ASSET_PROGRESS_TABLE).
				WithData(map[string]interface{}{
					models.ASSET_PROGRESS_POSITION:      assetProgress.Position,
					models.ASSET_PROGRESS_PROGRESS_FRAC: assetProgress.ProgressFrac,
					models.ASSET_PROGRESS_COMPLETED:     assetProgress.Completed,
					models.ASSET_PROGRESS_COMPLETED_AT:  assetProgress.CompletedAt,
					models.BASE_UPDATED_AT:              assetProgress.UpdatedAt,
				}).
				SetDbOpts(dbOpts)

			if _, err := updateGeneric(txCtx, dao, *builderOpts); err != nil {
				return err
			}
		}

		return dao.SyncCourseProgress(txCtx, asset.CourseID)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssetProgress gets a record from the asset progress table based upon the where clause in the options. If
// there is no where clause, it will return the first record in the table
func (dao *DAO) GetAssetProgress(ctx context.Context, dbOpts *database.Options) (*models.AssetProgress, error) {
	builderOpts := newBuilderOptions(models.ASSET_PROGRESS_TABLE).
		WithColumns(models.AssetProgressColumns()...).
		SetDbOpts(dbOpts).
		WithLimit(1)

	return getGeneric[models.AssetProgress](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListAssetProgress gets all records from the asset progress table based upon the where clause and pagination
// in the options
func (dao *DAO) ListAssetProgress(ctx context.Context, dbOpts *database.Options) ([]*models.AssetProgress, error) {
	builderOpts := newBuilderOptions(models.ASSET_PROGRESS_TABLE).
		WithColumns(models.AssetProgressColumns()...).
		SetDbOpts(dbOpts)

	return listGeneric[models.AssetProgress](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// ListAssetProgressIDs returns just the asset progress IDs as a []string
//
// TODO add tests
func (dao *DAO) ListAssetProgressIDs(ctx context.Context, dbOpts *database.Options) ([]string, error) {
	builderOpts := newBuilderOptions(models.ASSET_PROGRESS_TABLE).
		WithColumns(models.ASSET_PROGRESS_TABLE_ID).
		SetDbOpts(dbOpts)

	return pluck[string](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAssetProgress deletes records from the asset progress table
//
// Errors when a where clause is not provided
func (dao *DAO) DeleteAssetProgress(ctx context.Context, dbOpts *database.Options) error {
	if dbOpts == nil || dbOpts.Where == nil {
		return utils.ErrWhere
	}

	builderOpts := newBuilderOptions(models.ASSET_PROGRESS_TABLE).SetDbOpts(dbOpts)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAssetProgressForCourse deletes all asset progress records for a given user that
// belong to a course
func (dao *DAO) DeleteAssetProgressForCourse(ctx context.Context, courseID, userID string) error {
	if courseID == "" {
		return utils.ErrCourseId
	}

	if userID == "" {
		return utils.ErrUserId
	}

	where := squirrel.And{
		squirrel.Eq{models.ASSET_PROGRESS_TABLE_USER_ID: userID},
		squirrel.Expr(
			"EXISTS (SELECT 1 FROM "+models.ASSET_TABLE+
				" WHERE "+models.ASSET_TABLE_ID+" = "+models.ASSET_PROGRESS_TABLE_ASSET_ID+
				" AND "+models.ASSET_TABLE_COURSE_ID+" = ?)",
			courseID,
		),
	}

	dbOpts := database.NewOptions().WithWhere(where)

	return dao.DeleteAssetProgress(ctx, dbOpts)
}
