package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Param defines the model for a parameter
type Param struct {
	Base
	Key   string `db:"key"`   // Immutable
	Value string `db:"value"` // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	PARAM_TABLE = "params"
	PARAM_KEY   = "key"
	PARAM_VALUE = "value"

	PARAM_TABLE_ID         = PARAM_TABLE + "." + BASE_ID
	PARAM_TABLE_CREATED_AT = PARAM_TABLE + "." + BASE_CREATED_AT
	PARAM_TABLE_UPDATED_AT = PARAM_TABLE + "." + BASE_UPDATED_AT
	PARAM_TABLE_KEY        = PARAM_TABLE + "." + PARAM_KEY
	PARAM_TABLE_VALUE      = PARAM_TABLE + "." + PARAM_VALUE
)
