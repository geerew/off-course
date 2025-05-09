package dao

import (
	"context"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAsset creates an asset and refreshes course progress
func (dao *DAO) CreateAsset(ctx context.Context, asset *models.Asset) error {
	if asset == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, asset)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAsset retrieves an asset
func (dao *DAO) GetAsset(ctx context.Context, asset *models.Asset, options *database.Options) error {
	if asset == nil {
		return utils.ErrNilPtr
	}

	userId, ok := ctx.Value(types.UserContextKey).(string)
	if !ok || userId == "" {
		return utils.ErrMissingUserId
	}

	if options == nil {
		options = &database.Options{}
	}

	options.AddRelationFilter("Progress", models.ASSET_PROGRESS_USER_ID, userId)

	return dao.Get(ctx, asset, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListAssets retrieves a list of assets
func (dao *DAO) ListAssets(ctx context.Context, assets *[]*models.Asset, options *database.Options) error {
	if assets == nil {
		return utils.ErrNilPtr
	}

	userId, ok := ctx.Value(types.UserContextKey).(string)
	if !ok || userId == "" {
		return utils.ErrMissingUserId
	}

	if options == nil {
		options = &database.Options{}
	}

	options.AddRelationFilter("Progress", models.ASSET_PROGRESS_USER_ID, userId)

	return dao.List(ctx, assets, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAsset updates an asset
func (dao *DAO) UpdateAsset(ctx context.Context, asset *models.Asset) error {
	if asset == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, asset)
	return err
}
