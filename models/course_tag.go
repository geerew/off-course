package models

import (
	"github.com/geerew/off-course/utils/schema"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseTag defines the model for a course tag
type CourseTag struct {
	Base
	TagID    string
	CourseID string

	// Joins
	Course string
	Tag    string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	COURSE_TAG_TABLE     = "courses_tags"
	COURSE_TAG_TAG_ID    = "tag_id"
	COURSE_TAG_COURSE_ID = "course_id"

	COURSE_TAG_TABLE_ID         = COURSE_TAG_TABLE + "." + BASE_ID
	COURSE_TAG_TABLE_CREATED_AT = COURSE_TAG_TABLE + "." + BASE_CREATED_AT
	COURSE_TAG_TABLE_UPDATED_AT = COURSE_TAG_TABLE + "." + BASE_UPDATED_AT
	COURSE_TAG_TABLE_TAG_ID     = COURSE_TAG_TABLE + "." + COURSE_TAG_TAG_ID
	COURSE_TAG_TABLE_COURSE_ID  = COURSE_TAG_TABLE + "." + COURSE_TAG_COURSE_ID
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (ct *CourseTag) Table() string {
	return COURSE_TAG_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Define implements the `schema.Modeler` interface by defining the model
func (ct *CourseTag) Define(c *schema.ModelConfig) {
	c.Embedded("Base")

	// Common fields
	c.Field("TagID").Column(COURSE_TAG_TAG_ID).NotNull()
	c.Field("CourseID").Column(COURSE_TAG_COURSE_ID).NotNull()

	// Join field
	c.Field("Course").JoinTable(COURSE_TABLE).Column(COURSE_TITLE).Alias("course_title")
	c.Field("Tag").JoinTable(TAG_TABLE).Column(TAG_TAG).Alias("tag_tag")

	c.LeftJoin(COURSE_TABLE).On(COURSE_TAG_TABLE_COURSE_ID + " = " + COURSE_TABLE_ID)
	c.LeftJoin(TAG_TABLE).On(COURSE_TAG_TABLE_TAG_ID + " = " + TAG_TABLE_ID)

}
