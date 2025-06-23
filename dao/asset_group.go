package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAssetGroup creates an asset group
func (dao *DAO) CreateAssetGroup(ctx context.Context, assetGRoup *models.AssetGroup) error {
	if assetGRoup == nil {
		return utils.ErrNilPtr
	}

	return Create(ctx, dao, assetGRoup)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssetGroup retrieves an asset group
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetAssetGroup(ctx context.Context, assetGroup *models.AssetGroup, options *database.Options) error {
	if assetGroup == nil {
		return utils.ErrNilPtr
	}

	if options == nil {
		options = &database.Options{}
	}

	// When there is no where clause, use the ID
	if options.Where == nil {
		if assetGroup.Id() == "" {
			return utils.ErrInvalidId
		}

		options.Where = squirrel.Eq{models.ASSET_GROUP_TABLE_ID: assetGroup.Id()}
	}

	return Get(ctx, dao, assetGroup, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListAssetGroups retrieves a list of asset groups
func (dao *DAO) ListAssetGroups(ctx context.Context, assetGroups *[]*models.AssetGroup, options *database.Options) error {
	if assetGroups == nil {
		return utils.ErrNilPtr
	}

	if options == nil {
		options = &database.Options{}
	}

	return List(ctx, dao, assetGroups, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAssetGroup updates an asset group
func (dao *DAO) UpdateAssetGroup(ctx context.Context, assetGroup *models.AssetGroup) error {
	if assetGroup == nil {
		return utils.ErrNilPtr
	}

	_, err := Update(ctx, dao, assetGroup)
	return err
}
