package dao

import (
	"context"
	"slices"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAsset creates an asset and refreshes course progress
func (dao *DAO) CreateAsset(ctx context.Context, asset *models.Asset) error {
	if asset == nil {
		return utils.ErrNilPtr
	}

	return Create(ctx, dao, asset)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAsset retrieves an asset
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetAsset(ctx context.Context, asset *models.Asset, options *database.Options) error {
	if asset == nil {
		return utils.ErrNilPtr
	}

	if options == nil {
		options = &database.Options{}
	}

	// When there is no where clause, use the ID
	if options.Where == nil {
		if asset.Id() == "" {
			return utils.ErrInvalidId
		}

		options.Where = squirrel.Eq{models.ASSET_TABLE_ID: asset.Id()}
	}

	principal, err := principalFromCtx(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(options.ExcludeRelations, models.ASSET_RELATION_PROGRESS) {
		options.AddRelationFilter(models.ASSET_RELATION_PROGRESS, models.ASSET_PROGRESS_USER_ID, principal.UserID)
	}

	return Get(ctx, dao, asset, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListAssets retrieves a list of assets
func (dao *DAO) ListAssets(ctx context.Context, assets *[]*models.Asset, options *database.Options) error {
	if assets == nil {
		return utils.ErrNilPtr
	}

	if options == nil {
		options = &database.Options{}
	}

	principal, err := principalFromCtx(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(options.ExcludeRelations, models.ASSET_RELATION_PROGRESS) {
		options.AddRelationFilter(models.ASSET_RELATION_PROGRESS, models.ASSET_PROGRESS_USER_ID, principal.UserID)
	}

	return List(ctx, dao, assets, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAsset updates an asset
func (dao *DAO) UpdateAsset(ctx context.Context, asset *models.Asset) error {
	if asset == nil {
		return utils.ErrNilPtr
	}

	_, err := Update(ctx, dao, asset)
	return err
}
