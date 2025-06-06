package models

import "github.com/geerew/off-course/utils/schema"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Param defines the model for a parameter
type Param struct {
	Base
	Key   string
	Value string
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (p *Param) Table() string {
	return PARAM_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (p *Param) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("Key").Column(PARAM_KEY).NotNull()
	s.Field("Value").Column(PARAM_VALUE).NotNull().Mutable()
}
