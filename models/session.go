package models

import (
	"github.com/geerew/off-course/utils/schema"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgress defines the model for an asset progress
type Session struct {
	ID      string
	Data    []byte
	Expires int64
	UserId  string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	SESSION_TABLE   = "sessions"
	SESSION_ID      = "id"
	SESSION_DATA    = "data"
	SESSION_EXPIRES = "expires"
	SESSION_USER_ID = "user_id"

	SESSION_TABLE_ID      = SESSION_TABLE + "." + SESSION_ID
	SESSION_TABLE_DATA    = SESSION_TABLE + "." + SESSION_DATA
	SESSION_TABLE_EXPIRES = SESSION_TABLE + "." + SESSION_EXPIRES
	SESSION_TABLE_USER_ID = SESSION_TABLE + "." + SESSION_USER_ID
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Id returns the model ID
func (s *Session) Id() string {
	return s.ID
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshId noop
func (s *Session) RefreshId() {}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshCreatedAt noop
func (s *Session) RefreshCreatedAt() {}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshUpdatedAt noop
func (s *Session) RefreshUpdatedAt() {}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (s *Session) Table() string {
	return SESSION_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (a *Session) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("ID").Column(SESSION_ID).NotNull()
	s.Field("Data").Column(SESSION_DATA).NotNull().Mutable()
	s.Field("Expires").Column(SESSION_EXPIRES).NotNull().Mutable()
	s.Field("UserId").Column(SESSION_USER_ID).NotNull().Mutable()
}
