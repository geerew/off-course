package models

import (
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Modeler interface {
	Table() string
	Id() string
	RefreshId()
	RefreshCreatedAt()
	RefreshUpdatedAt()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Base defines the base model for all models
type Base struct {
	ID        string         `db:"id"`         // Immutable
	CreatedAt types.DateTime `db:"created_at"` // Immutable
	UpdatedAt types.DateTime `db:"updated_at"` // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	BASE_ID         = "id"
	BASE_CREATED_AT = "created_at"
	BASE_UPDATED_AT = "updated_at"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Id returns the model ID
func (b *Base) Id() string {
	return b.ID
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshId generates and sets a new model ID
func (b *Base) RefreshId() {
	b.ID = security.PseudorandomString(10)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetId sets the model ID
func (b *Base) SetId(id string) {
	b.ID = id
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshCreatedAt updates the Created At field to the current date/time
func (b *Base) RefreshCreatedAt() {
	b.CreatedAt = types.NowDateTime()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshUpdatedAt updates the Updated At field to the current date/time
func (b *Base) RefreshUpdatedAt() {
	b.UpdatedAt = types.NowDateTime()
}
