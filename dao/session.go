package dao

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
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

	if options == nil {
		options = &database.Options{}
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateSessionsRoleForUser updates the role for all sessions belonging to a user
func (dao *DAO) UpdateSessionsRoleForUser(ctx context.Context, userID string, newRole types.UserRole) error {
	sessions := []*models.Session{}
	opts := &database.Options{
		Where: squirrel.Eq{models.SESSION_TABLE_USER_ID: userID},
	}
	if err := dao.ListSessions(ctx, &sessions, opts); err != nil {
		return err
	}

	var updatedSessions []*models.Session
	for _, sess := range sessions {
		var values map[string]interface{}
		buf := bytes.NewBuffer(sess.Data)
		if err := gob.NewDecoder(buf).Decode(&values); err != nil {
			continue
		}

		values["role"] = newRole.String()

		var out bytes.Buffer
		if err := gob.NewEncoder(&out).Encode(values); err != nil {
			continue
		}

		sess.Data = out.Bytes()
		updatedSessions = append(updatedSessions, sess)
	}

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		for _, sess := range updatedSessions {
			if err := dao.UpdateSession(txCtx, sess); err != nil {
				return err
			}
		}
		return nil
	})
}
