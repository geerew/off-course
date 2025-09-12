package models

import "fmt"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	SESSION_TABLE = "sessions"

	SESSION_ID      = "id"
	SESSION_USER_ID = "user_id"
	SESSION_DATA    = "data"
	SESSION_EXPIRES = "expires"

	SESSION_TABLE_ID      = SESSION_TABLE + "." + SESSION_ID
	SESSION_TABLE_USER_ID = SESSION_TABLE + "." + SESSION_USER_ID
	SESSION_TABLE_DATA    = SESSION_TABLE + "." + SESSION_DATA
	SESSION_TABLE_EXPIRES = SESSION_TABLE + "." + SESSION_EXPIRES
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgress defines the model for an asset progress
type Session struct {
	ID      string `db:"id"`      // Immutable
	UserId  string `db:"user_id"` // Immutable
	Data    []byte `db:"data"`    // Mutable
	Expires int64  `db:"expires"` // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Id returns the model ID
func (s *Session) Id() string {
	return s.ID
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshId noop
func (s *Session) RefreshId() {}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SessionColumns returns the list of columns in the session table
func SessionColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", SESSION_TABLE_ID),
		fmt.Sprintf("%s AS user_id", SESSION_TABLE_USER_ID),
		fmt.Sprintf("%s AS data", SESSION_TABLE_DATA),
		fmt.Sprintf("%s AS expires", SESSION_TABLE_EXPIRES),
	}
}
