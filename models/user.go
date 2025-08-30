package models

import (
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// User defines the model for a user
type User struct {
	Base

	Username     string         `db:"username"`      // Immutable
	DisplayName  string         `db:"display_name"`  // Mutable
	PasswordHash string         `db:"password_hash"` // Mutable
	Role         types.UserRole `db:"role"`          // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	USER_TABLE         = "users"
	USER_USERNAME      = "username"
	USER_DISPLAY_NAME  = "display_name"
	USER_PASSWORD_HASH = "password_hash"
	USER_ROLE          = "role"

	USER_TABLE_ID            = USER_TABLE + "." + BASE_ID
	USER_TABLE_CREATED_AT    = USER_TABLE + "." + BASE_CREATED_AT
	USER_TABLE_UPDATED_AT    = USER_TABLE + "." + BASE_UPDATED_AT
	USER_TABLE_USERNAME      = USER_TABLE + "." + USER_USERNAME
	USER_TABLE_DISPLAY_NAME  = USER_TABLE + "." + USER_DISPLAY_NAME
	USER_TABLE_PASSWORD_HASH = USER_TABLE + "." + USER_PASSWORD_HASH
	USER_TABLE_ROLE          = USER_TABLE + "." + USER_ROLE
)
