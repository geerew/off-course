package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateParam creates a parameter
func (dao *DAO) CreateParam(ctx context.Context, param *models.Param) error {
	if param == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, param)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetParam retrieves a parameter
//
// When options is nil or options.Where is nil, the models Key will be used
func (dao *DAO) GetParam(ctx context.Context, param *models.Param, options *database.Options) error {
	if param == nil {
		return utils.ErrNilPtr
	}

	if options == nil || options.Where == nil {
		if param.Key == "" {
			return utils.ErrInvalidKey
		}

		options = &database.Options{
			Where: squirrel.Eq{models.PARAM_TABLE_KEY: param.Key},
		}
	}

	return dao.Get(ctx, param, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListParams retrieves a list of params
func (dao *DAO) ListParams(ctx context.Context, params *[]*models.Param, options *database.Options) error {
	if params == nil {
		return utils.ErrNilPtr
	}

	return dao.List(ctx, params, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateParam updates a parameter
func (dao *DAO) UpdateParam(ctx context.Context, param *models.Param) error {
	if param == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, param)
	return err
}
