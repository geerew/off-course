package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag defines the model for a tag
type Tag struct {
	Base
	Tag string `db:"tag"` // Mutable

	// Aggregate fields
	CourseCount int `db:"course_count"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	TAG_TABLE        = "tags"
	TAG_TAG          = "tag"
	TAG_COURSE_COUNT = "course_count"

	TAG_TABLE_ID         = TAG_TABLE + "." + BASE_ID
	TAG_TABLE_CREATED_AT = TAG_TABLE + "." + BASE_CREATED_AT
	TAG_TABLE_UPDATED_AT = TAG_TABLE + "." + BASE_UPDATED_AT
	TAG_TABLE_TAG        = TAG_TABLE + "." + TAG_TAG
)
