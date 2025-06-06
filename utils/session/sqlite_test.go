package session

import (
	"database/sql"
	"testing"
	"time"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/models"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqlite_Set(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.Set("key", []byte("value"), 1*time.Second)
		require.NoError(t, err)

		count, err := dao.Count(ctx, storage.dao, &models.Session{}, nil)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("replace", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.Set("key", []byte("value"), 1*time.Second)
		require.NoError(t, err)

		err = storage.Set("key", []byte("new value"), 1*time.Second)
		require.NoError(t, err)

		count, err := dao.Count(ctx, storage.dao, &models.Session{}, nil)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("no key", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.Set("", []byte("value"), 1*time.Second)
		require.NoError(t, err)

		count, err := dao.Count(ctx, storage.dao, &models.Session{}, nil)
		require.NoError(t, err, sql.ErrNoRows)
		require.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqlite_SetUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.Set("key", []byte("value"), 1*time.Second)
		require.NoError(t, err)

		err = storage.SetUser("key", "1234")
		require.NoError(t, err)

		session := &models.Session{ID: "key"}
		err = storage.dao.GetSession(ctx, session, nil)
		require.NoError(t, err)
		require.Equal(t, "1234", session.UserId)
	})

	t.Run("no session", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.SetUser("key", "1234")
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("no key or user", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.SetUser("", "1234")
		require.NoError(t, err)

		err = storage.SetUser("key", "")
		require.NoError(t, err)

		count, err := dao.Count(ctx, storage.dao, &models.Session{}, nil)
		require.NoError(t, err, sql.ErrNoRows)
		require.Equal(t, 0, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqlite_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.Set("key", []byte("value"), 1*time.Second)
		require.NoError(t, err)

		res, err := storage.Get("key")
		require.NoError(t, err)
		require.Equal(t, []byte("value"), res)
	})

	t.Run("no key", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		res, err := storage.Get("")
		require.NoError(t, err)
		require.Nil(t, res)
	})

	t.Run("expired", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, 1*time.Second)

		err := storage.Set("key", []byte("value"), 1*time.Millisecond)
		require.NoError(t, err)

		time.Sleep(2 * time.Millisecond)

		res, err := storage.Get("key")
		require.NoError(t, err)
		require.Nil(t, res)
	})

	t.Run("gc", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.Set("key", []byte("value"), 1*time.Millisecond)
		require.NoError(t, err)

		time.Sleep(2 * time.Millisecond)

		res, err := storage.Get("key")
		require.NoError(t, err)
		require.Nil(t, res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqlite_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.Set("key", []byte("value"), 1*time.Second)
		require.NoError(t, err)

		err = storage.Delete("key")
		require.NoError(t, err)

		count, err := dao.Count(ctx, storage.dao, &models.Session{}, nil)
		require.NoError(t, err, sql.ErrNoRows)
		require.Equal(t, 0, count)
	})

	t.Run("no key", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.Delete("")
		require.NoError(t, err)
	})

	t.Run("gc", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.Set("key", []byte("value"), 1*time.Millisecond)
		require.NoError(t, err)

		time.Sleep(2 * time.Millisecond)

		err = storage.Delete("key")
		require.NoError(t, err)
	})
}

func TestSqlite_DeleteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		// Set two sessions for the same user
		err := storage.Set("key1", []byte("value"), 1*time.Second)
		require.NoError(t, err)

		err = storage.SetUser("key1", "1234")
		require.NoError(t, err)

		err = storage.Set("key2", []byte("value"), 1*time.Second)
		require.NoError(t, err)

		err = storage.SetUser("key2", "1234")
		require.NoError(t, err)

		// Set a session for a different user
		err = storage.Set("key3", []byte("value"), 1*time.Second)
		require.NoError(t, err)

		err = storage.SetUser("key3", "4567")
		require.NoError(t, err)

		err = storage.DeleteUser("1234")
		require.NoError(t, err)

		count, err := dao.Count(ctx, storage.dao, &models.Session{}, nil)
		require.NoError(t, err, sql.ErrNoRows)
		require.Equal(t, 1, count)
	})

	t.Run("no id", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, 1*time.Millisecond)

		err := storage.DeleteUser("")
		require.NoError(t, err)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqlite_Reset(t *testing.T) {
	db, ctx := setup(t)

	storage := NewSqliteStorage(db, 1*time.Millisecond)

	err := storage.Set("key 1", []byte("value"), 1*time.Second)
	require.NoError(t, err)
	err = storage.Set("key 2", []byte("value"), 1*time.Second)
	require.NoError(t, err)
	err = storage.Set("key 3", []byte("value"), 1*time.Second)
	require.NoError(t, err)

	err = storage.Reset()
	require.NoError(t, err)

	count, err := dao.Count(ctx, storage.dao, &models.Session{}, nil)
	require.NoError(t, err, sql.ErrNoRows)
	require.Equal(t, 0, count)
}
