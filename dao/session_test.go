package dao

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateOrReplaceSession(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "session-1",
			UserId:  "user-123",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.NoError(t, dao.CreateOrReplaceSession(ctx, session))
	})

	t.Run("replace", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "session-1",
			UserId:  "user-123",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.NoError(t, dao.CreateOrReplaceSession(ctx, session))

		// Replace the session
		session.Data = []byte("updated session data")
		require.NoError(t, dao.CreateOrReplaceSession(ctx, session))

		// Verify the session was updated
		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_USER_ID: session.UserId})
		record, err := dao.GetSession(ctx, dbOpts)
		require.NoError(t, err)
		require.Equal(t, session.UserId, record.UserId)
		require.Equal(t, session.Data, record.Data)
	})

	t.Run("nil pointer", func(t *testing.T) {
		dao, ctx := setup(t)

		require.ErrorIs(t, dao.CreateOrReplaceSession(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("invalid ID", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "",
			UserId:  "user-123",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.ErrorIs(t, dao.CreateOrReplaceSession(ctx, session), utils.ErrId)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "session-1",
			UserId:  "",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.ErrorIs(t, dao.CreateOrReplaceSession(ctx, session), utils.ErrUserId)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetSession(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "session-1",
			UserId:  "user-123",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.NoError(t, dao.CreateOrReplaceSession(ctx, session))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_ID: session.ID})
		record, err := dao.GetSession(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, session.ID, record.ID)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		record, err := dao.GetSession(ctx, nil)
		require.Nil(t, err)
		require.Nil(t, record)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListSessions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		sessions := []*models.Session{}
		for i := range 3 {
			session := &models.Session{
				ID:      fmt.Sprintf("session-%d", i),
				UserId:  fmt.Sprintf("user-%d", i),
				Data:    []byte(fmt.Sprintf("session data %d", i)),
				Expires: time.Now().Add(24 * time.Hour).Unix(),
			}
			sessions = append(sessions, session)
			require.NoError(t, dao.CreateOrReplaceSession(ctx, session))
			time.Sleep(1 * time.Millisecond)
		}

		records, err := dao.ListSessions(ctx, nil)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, sessions[i].ID, record.ID)
		}
	})

	t.Run("empty", func(t *testing.T) {
		dao, ctx := setup(t)

		records, err := dao.ListSessions(ctx, nil)
		require.Nil(t, err)
		require.Empty(t, records)
	})

	t.Run("order by", func(t *testing.T) {
		dao, ctx := setup(t)

		sessions := []*models.Session{}
		for i := range 3 {
			session := &models.Session{
				ID:      fmt.Sprintf("session-%d", i),
				UserId:  fmt.Sprintf("user-%d", i),
				Data:    []byte(fmt.Sprintf("session data %d", i)),
				Expires: time.Now().Add(24 * time.Hour).Unix(),
			}
			sessions = append(sessions, session)
			require.NoError(t, dao.CreateOrReplaceSession(ctx, session))
			time.Sleep(1 * time.Millisecond)
		}

		// Descending order by created_at
		opts := database.NewOptions().WithOrderBy(models.SESSION_TABLE_USER_ID + " DESC")

		records, err := dao.ListSessions(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, sessions[2-i].ID, record.ID)
		}

		// Ascending order by created_at
		opts = database.NewOptions().WithOrderBy(models.SESSION_TABLE_USER_ID + " ASC")

		records, err = dao.ListSessions(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, sessions[i].ID, record.ID)
		}
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "session-1",
			UserId:  "users-1",
			Data:    []byte("session data 1"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.NoError(t, dao.CreateOrReplaceSession(ctx, session))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_ID: session.ID})
		records, err := dao.ListSessions(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 1)
		require.Equal(t, session.ID, records[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, ctx := setup(t)

		sessions := []*models.Session{}
		for i := range 17 {
			session := &models.Session{
				ID:      fmt.Sprintf("session-%d", i),
				UserId:  fmt.Sprintf("user-%d", i),
				Data:    []byte(fmt.Sprintf("session data %d", i)),
				Expires: time.Now().Add(24 * time.Hour).Unix(),
			}
			sessions = append(sessions, session)
			require.NoError(t, dao.CreateOrReplaceSession(ctx, session))
			time.Sleep(1 * time.Millisecond)
		}

		// First page with 10 records
		p := database.NewOptions().WithPagination(pagination.New(1, 10))
		records, err := dao.ListSessions(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 10)
		require.Equal(t, sessions[0].ID, records[0].ID)
		require.Equal(t, sessions[9].ID, records[9].ID)

		// Second page with remaining 7 records
		p = database.NewOptions().WithPagination(pagination.New(2, 10))
		records, err = dao.ListSessions(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 7)
		require.Equal(t, sessions[10].ID, records[0].ID)
		require.Equal(t, sessions[16].ID, records[6].ID)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateSession(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		originalSession := &models.Session{
			ID:      "session-1",
			UserId:  "user-123",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.NoError(t, dao.CreateOrReplaceSession(ctx, originalSession))

		updatedSession := &models.Session{
			ID:      originalSession.ID,
			UserId:  originalSession.UserId,                // Immutable
			Data:    []byte("updated session data"),        // Mutable
			Expires: time.Now().Add(48 * time.Hour).Unix(), // Mutable
		}

		time.Sleep(1 * time.Millisecond)
		require.NoError(t, dao.UpdateSession(ctx, updatedSession))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_ID: originalSession.ID})
		record, err := dao.GetSession(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, originalSession.ID, record.ID)          // No change
		require.Equal(t, originalSession.UserId, record.UserId)  // Immutable
		require.Equal(t, updatedSession.Data, record.Data)       // Changed
		require.Equal(t, updatedSession.Expires, record.Expires) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "session-1",
			UserId:  "user-123",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.NoError(t, dao.CreateOrReplaceSession(ctx, session))

		// Empty ID
		session.ID = ""
		require.ErrorIs(t, dao.UpdateSession(ctx, session), utils.ErrId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateSession(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
func Test_UpdateSessionRoleForUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		for i := range 3 {
			session := &models.Session{
				ID:      fmt.Sprintf("session-%d", i),
				UserId:  "user-123",
				Expires: time.Now().Add(24 * time.Hour).Unix(),
			}

			// Set the user role in the session data
			values := map[string]interface{}{
				"role": types.UserRoleUser.String(),
			}

			var out bytes.Buffer
			require.NoError(t, gob.NewEncoder(&out).Encode(values))
			session.Data = out.Bytes()

			require.NoError(t, dao.CreateOrReplaceSession(ctx, session))

			time.Sleep(1 * time.Millisecond)
		}

		require.NoError(t, dao.UpdateSessionRoleForUser(ctx, "user-123", types.UserRoleAdmin))

		records, err := dao.ListSessions(ctx, nil)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for _, record := range records {

			buf := bytes.NewBuffer(record.Data)
			var values map[string]interface{}
			require.NoError(t, gob.NewDecoder(buf).Decode(&values))

			require.Equal(t, types.UserRoleAdmin.String(), values["role"], "Role should be updated to admin")
		}
	})

	t.Run("invalid user ID", func(t *testing.T) {
		dao, ctx := setup(t)

		require.ErrorIs(t, dao.UpdateSessionRoleForUser(ctx, "", "admin"), utils.ErrUserId)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteSessions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "session-1",
			UserId:  "user-123",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.NoError(t, dao.CreateOrReplaceSession(ctx, session))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_ID: session.ID})
		require.Nil(t, dao.DeleteSessions(ctx, opts))

		records, err := dao.ListSessions(ctx, opts)
		require.NoError(t, err)
		require.Empty(t, records)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "session-1",
			UserId:  "user-123",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.NoError(t, dao.CreateOrReplaceSession(ctx, session))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_ID: "non-existent"})
		require.Nil(t, dao.DeleteSessions(ctx, opts))

		records, err := dao.ListSessions(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, session.ID, records[0].ID)
	})

	t.Run("missing where", func(t *testing.T) {
		dao, ctx := setup(t)

		session := &models.Session{
			ID:      "session-1",
			UserId:  "user-123",
			Data:    []byte("session data"),
			Expires: time.Now().Add(24 * time.Hour).Unix(),
		}
		require.NoError(t, dao.CreateOrReplaceSession(ctx, session))

		require.ErrorIs(t, dao.DeleteSessions(ctx, nil), utils.ErrWhere)

		records, err := dao.ListSessions(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, session.ID, records[0].ID)
	})
}
