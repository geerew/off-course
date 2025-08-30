package session

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqlite_Set(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, time.Hour)

		// Build a Fiber-like session map and gob-encode it.
		payload := map[string]any{
			"id":   "user-123",
			"role": "admin",
		}
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(payload)
		require.NoError(t, err)

		// Call Set with gob bytes and a short expiration (e.g., 1s)
		err = storage.Set("key", buf.Bytes(), time.Second)
		require.NoError(t, err)

		records, err := storage.dao.ListSessions(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)

		require.Equal(t, "key", records[0].ID)
		require.Equal(t, "user-123", records[0].UserId)             // Set() extracted from gob and stored
		require.Greater(t, records[0].Expires, time.Now().Unix()-1) // expires is set
		require.NotEmpty(t, records[0].Data)
	})

	t.Run("replace", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, time.Hour)

		// First payload
		p1 := map[string]any{
			"id":   "user-123",
			"role": "user",
			"msg":  "value",
		}
		var b1 bytes.Buffer
		require.NoError(t, gob.NewEncoder(&b1).Encode(p1))

		err := storage.Set("key", b1.Bytes(), time.Second)
		require.NoError(t, err)

		// Second payload (same key, different content)
		p2 := map[string]any{
			"id":   "user-123",
			"role": "admin",
			"msg":  "new value",
		}
		var b2 bytes.Buffer
		require.NoError(t, gob.NewEncoder(&b2).Encode(p2))

		err = storage.Set("key", b2.Bytes(), time.Second)
		require.NoError(t, err)

		// Optional: verify the row content was replaced
		records, err := storage.dao.ListSessions(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, "key", records[0].ID)
		require.Equal(t, "user-123", records[0].UserId)

		// Data should match the second payload (gob bytes differ)
		require.NotEqual(t, b1.Bytes(), records[0].Data)
		require.Equal(t, b2.Bytes(), records[0].Data)
	})

	t.Run("no key", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, time.Millisecond)

		err := storage.Set("", []byte("value"), time.Second)
		require.NoError(t, err)

		records, err := storage.dao.ListSessions(ctx, nil)
		require.NoError(t, err)
		require.Zero(t, records)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqlite_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, time.Hour)

		payload := map[string]any{
			"id":   "user-123",
			"role": "admin",
		}
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(payload)
		require.NoError(t, err)

		err = storage.Set("key", buf.Bytes(), time.Second)
		require.NoError(t, err)

		_, err = storage.Get("key")
		require.NoError(t, err)
	})

	t.Run("no key", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, time.Hour)

		res, err := storage.Get("")
		require.NoError(t, err)
		require.Nil(t, res)
	})

	t.Run("expired", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, time.Hour)

		payload := map[string]any{
			"id":   "user-123",
			"role": "admin",
		}
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(payload)
		require.NoError(t, err)

		err = storage.Set("key", buf.Bytes(), time.Millisecond)
		require.NoError(t, err)

		time.Sleep(2 * time.Millisecond)

		res, err := storage.Get("key")
		require.NoError(t, err)
		require.Nil(t, res)
	})

	t.Run("gc", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, time.Millisecond)

		payload := map[string]any{
			"id":   "user-123",
			"role": "admin",
		}
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(payload)
		require.NoError(t, err)

		err = storage.Set("key", buf.Bytes(), time.Millisecond)
		require.NoError(t, err)

		time.Sleep(5 * time.Millisecond)

		res, err := storage.Get("key")
		require.NoError(t, err)
		require.Nil(t, res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqlite_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, time.Hour)

		payload := map[string]any{
			"id":   "user-123",
			"role": "admin",
		}
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(payload)
		require.NoError(t, err)

		err = storage.Set("key", buf.Bytes(), time.Second)
		require.NoError(t, err)

		err = storage.Delete("key")
		require.NoError(t, err)

		records, err := storage.dao.ListSessions(ctx, nil)
		require.NoError(t, err)
		require.Zero(t, len(records))
	})

	t.Run("no key", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, time.Hour)

		err := storage.Delete("")
		require.NoError(t, err)
	})

	t.Run("gc", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, time.Millisecond)

		payload := map[string]any{
			"id":   "user-123",
			"role": "admin",
		}
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(payload)
		require.NoError(t, err)

		err = storage.Set("key", buf.Bytes(), time.Millisecond)
		require.NoError(t, err)

		time.Sleep(2 * time.Millisecond)

		err = storage.Delete("key")
		require.NoError(t, err)
	})
}

func TestSqlite_DeleteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, ctx := setup(t)

		storage := NewSqliteStorage(db, time.Millisecond)

		// Set two sessions for the same user
		payload := map[string]any{
			"id":   "user-123",
			"role": "admin",
		}
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(payload)
		require.NoError(t, err)

		err = storage.Set("key1", buf.Bytes(), time.Second)
		require.NoError(t, err)

		err = storage.Set("key2", buf.Bytes(), time.Second)
		require.NoError(t, err)

		// Set a session for a different user
		payload = map[string]any{
			"id":   "user-456",
			"role": "admin",
		}
		err = gob.NewEncoder(&buf).Encode(payload)
		require.NoError(t, err)

		err = storage.Set("key3", buf.Bytes(), time.Second)
		require.NoError(t, err)

		err = storage.DeleteUser("user-123")
		require.NoError(t, err)

		records, err := storage.dao.ListSessions(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)

		require.Equal(t, "key3", records[0].ID)
		require.Equal(t, "user-123", records[0].UserId)             // Set() extracted from gob and stored
		require.Greater(t, records[0].Expires, time.Now().Unix()-1) // expires is set
		require.NotEmpty(t, records[0].Data)
	})

	t.Run("no id", func(t *testing.T) {
		db, _ := setup(t)

		storage := NewSqliteStorage(db, time.Hour)

		err := storage.DeleteUser("")
		require.NoError(t, err)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSqlite_Reset(t *testing.T) {
	db, ctx := setup(t)

	storage := NewSqliteStorage(db, time.Hour)

	payload := map[string]any{
		"id":   "user-123",
		"role": "admin",
	}
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(payload)
	require.NoError(t, err)

	err = storage.Set("key 1", buf.Bytes(), time.Second)
	require.NoError(t, err)

	err = storage.Set("key 2", buf.Bytes(), time.Second)
	require.NoError(t, err)

	err = storage.Set("key 3", buf.Bytes(), time.Second)
	require.NoError(t, err)

	err = storage.Reset()
	require.NoError(t, err)

	records, err := storage.dao.ListSessions(ctx, nil)
	require.NoError(t, err)
	require.Zero(t, len(records))
}
