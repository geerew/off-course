package models

import (
	"fmt"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
const (
	LOG_TABLE = "logs"

	LOG_LEVEL   = "level"
	LOG_MESSAGE = "message"
	LOG_DATA    = "data"

	LOG_TABLE_ID         = LOG_TABLE + "." + BASE_ID
	LOG_TABLE_CREATED_AT = LOG_TABLE + "." + BASE_CREATED_AT
	LOG_TABLE_UPDATED_AT = LOG_TABLE + "." + BASE_UPDATED_AT
	LOG_TABLE_LEVEL      = LOG_TABLE + "." + LOG_LEVEL
	LOG_TABLE_MESSAGE    = LOG_TABLE + "." + LOG_MESSAGE
	LOG_TABLE_DATA       = LOG_TABLE + "." + LOG_DATA
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Log defines the model for a log
type Log struct {
	Base
	Level   int           `db:"level"`   // Immutable
	Message string        `db:"message"` // Immutable
	Data    types.JsonMap `db:"data"`    // Immutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// LogColumns returns the list of columns to use when populating `Log`
func LogColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", LOG_TABLE_ID),
		fmt.Sprintf("%s AS created_at", LOG_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS updated_at", LOG_TABLE_UPDATED_AT),
		fmt.Sprintf("%s AS level", LOG_TABLE_LEVEL),
		fmt.Sprintf("%s AS message", LOG_TABLE_MESSAGE),
		fmt.Sprintf("%s AS data", LOG_TABLE_DATA),
	}
}
