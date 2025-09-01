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

// CreateOrReplaceSession inserts or replace a session record
func (dao *DAO) CreateOrReplaceSession(ctx context.Context, session *models.Session) error {
	if session == nil {
		return utils.ErrNilPtr
	}

	if session.ID == "" {
		return utils.ErrId
	}

	if session.UserId == "" {
		return utils.ErrUserId
	}

	builderOpts := newBuilderOptions(models.SESSION_TABLE).
		WithData(
			map[string]interface{}{
				models.BASE_ID:         session.ID,
				models.SESSION_USER_ID: session.UserId,
				models.SESSION_DATA:    session.Data,
				models.SESSION_EXPIRES: session.Expires,
			},
		).
		WithReplace()

	return createGeneric(ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSession gets a record from the sessions table based upon the where clause in the options. If
// there is no where clause, it will return the first record in the table
func (dao *DAO) GetSession(ctx context.Context, dbOpts *database.Options) (*models.Session, error) {
	builderOpts := newBuilderOptions(models.SESSION_TABLE).
		WithColumns(
			models.SESSION_TABLE + ".*",
		).
		SetDbOpts(dbOpts).
		WithLimit(1)

	return getGeneric[models.Session](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListSessions gets all records from the sessions table based upon the where clause and pagination
// in the options
func (dao *DAO) ListSessions(ctx context.Context, dbOpts *database.Options) ([]*models.Session, error) {
	builderOpts := newBuilderOptions(models.SESSION_TABLE).
		WithColumns(
			models.SESSION_TABLE + ".*",
		).
		SetDbOpts(dbOpts)

	return listGeneric[models.Session](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateSession updates a session record
func (dao *DAO) UpdateSession(ctx context.Context, session *models.Session) error {
	if session == nil {
		return utils.ErrNilPtr
	}

	if session.ID == "" {
		return utils.ErrId
	}

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.BASE_ID: session.ID})

	builderOpts := newBuilderOptions(models.SESSION_TABLE).
		WithData(
			map[string]interface{}{
				models.SESSION_DATA:    session.Data,
				models.SESSION_EXPIRES: session.Expires,
			},
		).
		SetDbOpts(dbOpts)

	_, err := updateGeneric(ctx, dao, *builderOpts)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateSessionRoleForUser updates the role for all sessions belonging to a user
func (dao *DAO) UpdateSessionRoleForUser(ctx context.Context, userID string, newRole types.UserRole) error {
	if userID == "" {
		return utils.ErrUserId
	}

	opts := database.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_USER_ID: userID})

	sessions, err := dao.ListSessions(ctx, opts)
	if err != nil {
		return err
	}

	var updatedSessions []*models.Session
	for _, session := range sessions {
		var values map[string]interface{}
		buf := bytes.NewBuffer(session.Data)
		if err := gob.NewDecoder(buf).Decode(&values); err != nil {
			continue
		}

		values["role"] = newRole.String()

		var out bytes.Buffer
		if err := gob.NewEncoder(&out).Encode(values); err != nil {
			continue
		}

		session.Data = out.Bytes()
		updatedSessions = append(updatedSessions, session)
	}

	// TODO make bulk update
	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		for _, session := range updatedSessions {
			if err := dao.UpdateSession(txCtx, session); err != nil {
				return err
			}
		}
		return nil
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteSessions deletes records from the sessions table
//
// Errors when a where clause is not provided
func (dao *DAO) DeleteSessions(ctx context.Context, dbOpts *database.Options) error {
	if dbOpts == nil || dbOpts.Where == nil {
		return utils.ErrWhere
	}

	builderOpts := newBuilderOptions(models.SESSION_TABLE).SetDbOpts(dbOpts)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}

// DeleteAllSessions deletes all records from the sessions table
func (dao *DAO) DeleteAllSessions(ctx context.Context) error {
	builderOpts := newBuilderOptions(models.SESSION_TABLE)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}
