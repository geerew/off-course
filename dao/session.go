package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateOrReplaceSession creates or replaces a session
func (dao *DAO) CreateOrReplaceSession(ctx context.Context, session *models.Session) error {
	if session == nil {
		return utils.ErrNilPtr
	}

	return CreateOrReplace(ctx, dao, session)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSession retrieves a session
//
// When options is nil or options.Where is nil, the models ID will be used
func (dao *DAO) GetSession(ctx context.Context, session *models.Session, options *database.Options) error {
	if session == nil {
		return utils.ErrNilPtr
	}

	if options == nil || options.Where == nil {
		if session.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{
			Where: squirrel.Eq{models.SESSION_TABLE_ID: session.Id()},
		}
	}

	return Get(ctx, dao, session, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListSessions retrieves a list of sessions
func (dao *DAO) ListSessions(ctx context.Context, session *[]*models.Session, options *database.Options) error {
	if session == nil {
		return utils.ErrNilPtr
	}

	return List(ctx, dao, session, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateSession updates a session
func (dao *DAO) UpdateSession(ctx context.Context, session *models.Session) error {
	if session == nil {
		return utils.ErrNilPtr
	}

	_, err := Update(ctx, dao, session)
	return err
}
