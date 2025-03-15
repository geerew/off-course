package models

import "github.com/geerew/off-course/utils/schema"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag defines the model for a tag
type Tag struct {
	Base
	Tag string

	// Aggregate fields
	CourseCount int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	TAG_TABLE        = "tags"
	TAG_TAG          = "tag"
	TAG_COURSE_COUNT = "course_count"

	TAG_TABLE_ID         = TAG_TABLE + "." + BASE_ID
	TAG_TABLE_CREATED_AT = TAG_TABLE + "." + BASE_CREATED_AT
	TAG_TABLE_UPDATED_AT = TAG_TABLE + "." + BASE_UPDATED_AT
	TAG_TABLE_TAG        = TAG_TABLE + "." + TAG_TAG
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (t *Tag) Table() string {
	return TAG_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Define implements the `schema.Modeler` interface by defining the model
func (t *Tag) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("Tag").Column(TAG_TAG).NotNull().Mutable()
	s.Field("CourseCount").AggregateFn("COUNT").JoinTable(COURSE_TAG_TABLE).Column(COURSE_TAG_COURSE_ID).Alias(TAG_COURSE_COUNT)

	// Joins
	s.LeftJoin(COURSE_TAG_TABLE).On(TAG_TABLE_ID + " = " + COURSE_TAG_TABLE_TAG_ID)

	// Group by
	s.GroupBy(TAG_TABLE_ID)
}
