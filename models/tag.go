package models

import "github.com/geerew/off-course/utils/schema"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag defines the model for a tag
type Tag struct {
	Base
	Tag string

	// Relations
	CourseTags []*CourseTag
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	TAG_TABLE = "tags"
	TAG_TAG   = "tag"

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

	// Relation fields
	s.Relation("CourseTags").MatchOn(COURSE_TAG_TAG_ID)
}
