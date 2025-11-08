package session

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SqliteStorage is a sqlite storage
type SqliteStorage struct {
	dao        *dao.DAO
	gcInterval time.Duration
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewSqliteStorage creates a new sqlite storage
func NewSqliteStorage(db database.Database, gcInterval time.Duration) *SqliteStorage {
	storage := &SqliteStorage{
		dao:        dao.New(db),
		gcInterval: gcInterval,
	}

	// Start garbage collector
	go storage.gcTicker()

	return storage
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get value by key
func (s *SqliteStorage) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, nil
	}

	dbOpts := dao.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_ID: key})
	session, err := s.dao.GetSession(context.Background(), dbOpts)
	if err != nil {
		return nil, err
	}

	if session == nil || (session.Expires != 0 && session.Expires <= time.Now().Unix()) {
		return nil, nil
	}

	return session.Data, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Set key with value
func (s *SqliteStorage) Set(key string, data []byte, exp time.Duration) error {
	if key == "" || len(data) <= 0 {
		return nil
	}

	var expSeconds int64
	if exp != 0 {
		expSeconds = time.Now().Add(exp).Unix()
	}

	// Try to extract "id" (userId) from the serialized session payload.
	// Fiber session uses encoding/gob by default for the map[string]any.
	userID := ""
	if uid, ok := extractUserIDFromSessionBytes(data); ok {
		userID = uid
	}

	session := &models.Session{
		ID:      key,
		Data:    data,
		Expires: expSeconds,
		UserId:  userID,
	}

	return s.dao.CreateOrReplaceSession(context.Background(), session)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes an entry by ID (key)
func (s *SqliteStorage) Delete(key string) error {
	if key == "" {
		return nil
	}

	dbOpts := dao.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_ID: key})
	return s.dao.DeleteSessions(context.Background(), dbOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteUser deletes all entries for a user
func (s *SqliteStorage) DeleteUser(id string) error {
	if id == "" {
		return nil
	}

	dbOpts := dao.NewOptions().WithWhere(squirrel.Eq{models.SESSION_TABLE_USER_ID: id})
	return s.dao.DeleteSessions(context.Background(), dbOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Reset resets all entries, including unexpired
func (s *SqliteStorage) Reset() error {
	return s.dao.DeleteAllSessions(context.Background())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Close closes the database
func (s *SqliteStorage) Close() error {
	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// gcTicker starts a garbage collector that deletes expired sessions
func (s *SqliteStorage) gcTicker() {
	ticker := time.NewTicker(s.gcInterval)
	ctx := context.Background()
	defer ticker.Stop()

	for t := range ticker.C {
		dbOpts := dao.NewOptions().
			WithWhere(squirrel.And{
				squirrel.LtOrEq{models.SESSION_TABLE_EXPIRES: t.Unix()},
				squirrel.NotEq{models.SESSION_TABLE_EXPIRES: 0},
			})
		s.dao.DeleteSessions(ctx, dbOpts)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// helper: decode gob into map and read "id"
func extractUserIDFromSessionBytes(b []byte) (string, bool) {
	var m map[string]any
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&m); err != nil {
		return "", false
	}
	// SessionManager sets: session.Set("id", userId)
	if v, ok := m["id"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s, true
		}
	}
	return "", false
}
