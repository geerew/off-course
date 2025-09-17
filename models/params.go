package models

import "fmt"

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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Param defines the model for a parameter
type Param struct {
	Base
	Key   string `db:"key"`   // Immutable
	Value string `db:"value"` // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ParamColumns returns the list of columns to use when populating `Param`
func ParamColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", PARAM_TABLE_ID),
		fmt.Sprintf("%s AS created_at", PARAM_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS updated_at", PARAM_TABLE_UPDATED_AT),
		fmt.Sprintf("%s AS key", PARAM_TABLE_KEY),
		fmt.Sprintf("%s AS value", PARAM_TABLE_VALUE),
	}
}
