package models

import (
	"fmt"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	TAG_TABLE = "tags"

	TAG_TAG          = "tag"
	TAG_COURSE_COUNT = "course_count"

	TAG_TABLE_ID         = TAG_TABLE + "." + BASE_ID
	TAG_TABLE_CREATED_AT = TAG_TABLE + "." + BASE_CREATED_AT
	TAG_TABLE_UPDATED_AT = TAG_TABLE + "." + BASE_UPDATED_AT
	TAG_TABLE_TAG        = TAG_TABLE + "." + TAG_TAG
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag defines the model for a tag
type Tag struct {
	Base
	Tag string `db:"tag"` // Mutable

	// Aggregate fields
	CourseCount int `db:"course_count"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TagColumns returns the list of columns to use when populating `Tag`
func TagColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", TAG_TABLE_ID),
		fmt.Sprintf("%s AS created_at", TAG_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS updated_at", TAG_TABLE_UPDATED_AT),
		fmt.Sprintf("%s AS tag", TAG_TABLE_TAG),
		// Aggregate fields
		fmt.Sprintf("COUNT(%s) as course_count", COURSE_TAG_TABLE_COURSE_ID),
	}
}
