package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateUser creates a user
func (dao *DAO) CreateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, user)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetUser retrieves a tag
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetUser(ctx context.Context, user *models.User, options *database.Options) error {
	if user == nil {
		return utils.ErrNilPtr
	}

	if options == nil || options.Where == nil {
		if user.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{
			Where: squirrel.Eq{user.Table() + "." + models.BASE_ID: user.Id()},
		}
	}

	if options.Where == nil {
	}

	return dao.Get(ctx, user, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListUsers retrieves a list of users
func (dao *DAO) ListUsers(ctx context.Context, users *[]*models.User, options *database.Options) error {
	if users == nil {
		return utils.ErrNilPtr
	}

	return dao.List(ctx, users, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateUser updates a user
func (dao *DAO) UpdateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, user)
	return err
}
