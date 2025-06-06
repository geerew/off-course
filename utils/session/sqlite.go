package session

import (
	"context"
	"database/sql"
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

	session := &models.Session{ID: key}
	err := s.dao.GetSession(context.Background(), session, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	// If the expiration time has already passed, then return nil
	if session.Expires != 0 && session.Expires <= time.Now().Unix() {
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

	session := &models.Session{
		ID:      key,
		Data:    data,
		Expires: expSeconds,
	}

	return s.dao.CreateOrReplaceSession(context.Background(), session)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetUser sets the user ID for a session
func (s *SqliteStorage) SetUser(key, userId string) error {
	if key == "" || userId == "" {
		return nil
	}

	session := &models.Session{ID: key}
	err := s.dao.GetSession(context.Background(), session, nil)
	if err != nil {
		return err
	}

	session.UserId = userId

	err = s.dao.UpdateSession(context.Background(), session)
	if err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes an entry by ID (key)
func (s *SqliteStorage) Delete(key string) error {
	if key == "" {
		return nil
	}

	options := &database.Options{Where: squirrel.Eq{models.SESSION_TABLE_ID: key}}
	return dao.Delete(context.Background(), s.dao, &models.Session{}, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteUser deletes all entries for a user
func (s *SqliteStorage) DeleteUser(id string) error {
	if id == "" {
		return nil
	}

	options := &database.Options{Where: squirrel.Eq{models.SESSION_TABLE_USER_ID: id}}
	return dao.Delete(context.Background(), s.dao, &models.Session{}, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Reset resets all entries, including unexpired
func (s *SqliteStorage) Reset() error {
	return dao.DeleteAll(context.Background(), s.dao, &models.Session{})
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
		o := &database.Options{
			Where: squirrel.And{
				squirrel.LtOrEq{models.SESSION_TABLE_EXPIRES: t.Unix()},
				squirrel.NotEq{models.SESSION_TABLE_EXPIRES: 0}},
		}
		dao.Delete(ctx, s.dao, &models.Session{}, o)
	}
}
