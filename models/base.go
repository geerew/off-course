package models

import (
	"time"

	"github.com/geerew/off-course/utils/security"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// BaseModel defines the base model for all models
type BaseModel struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshId generates and sets a new model ID
func (b *BaseModel) RefreshId() {
	b.ID = security.PseudorandomString(10)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetId sets the model ID
func (b *BaseModel) SetId(id string) {
	b.ID = id
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshCreatedAt updates the Created At field to the current date/time
func (b *BaseModel) RefreshCreatedAt() {
	b.CreatedAt = time.Now()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshUpdatedAt updates the Updated At field to the current date/time
func (b *BaseModel) RefreshUpdatedAt() {
	b.UpdatedAt = time.Now()
}
