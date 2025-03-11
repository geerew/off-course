package models

import (
	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgress defines the model for a course progress
type CourseProgress struct {
	Base
	CourseID    string
	Started     bool
	StartedAt   types.DateTime
	Percent     int
	CompletedAt types.DateTime
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	COURSE_PROGRESS_TABLE        = "courses_progress"
	COURSE_PROGRESS_COURSE_ID    = "course_id"
	COURSE_PROGRESS_STARTED      = "started"
	COURSE_PROGRESS_STARTED_AT   = "started_at"
	COURSE_PROGRESS_PERCENT      = "percent"
	COURSE_PROGRESS_COMPLETED_AT = "completed_at"

	COURSE_PROGRESS_TABLE_ID           = COURSE_PROGRESS_TABLE + "." + BASE_ID
	COURSE_PROGRESS_TABLE_CREATED_AT   = COURSE_PROGRESS_TABLE + "." + BASE_CREATED_AT
	COURSE_PROGRESS_TABLE_UPDATED_AT   = COURSE_PROGRESS_TABLE + "." + BASE_UPDATED_AT
	COURSE_PROGRESS_TABLE_COURSE_ID    = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_COURSE_ID
	COURSE_PROGRESS_TABLE_STARTED      = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_STARTED
	COURSE_PROGRESS_TABLE_STARTED_AT   = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_STARTED_AT
	COURSE_PROGRESS_TABLE_PERCENT      = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_PERCENT
	COURSE_PROGRESS_TABLE_COMPLETED_AT = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_COMPLETED_AT
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (cp *CourseProgress) Table() string {
	return COURSE_PROGRESS_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (cp *CourseProgress) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("CourseID").Column(COURSE_PROGRESS_COURSE_ID).NotNull()
	s.Field("Started").Column(COURSE_PROGRESS_STARTED).Mutable()
	s.Field("StartedAt").Column(COURSE_PROGRESS_STARTED_AT).Mutable()
	s.Field("Percent").Column(COURSE_PROGRESS_PERCENT).Mutable()
	s.Field("CompletedAt").Column(COURSE_PROGRESS_COMPLETED_AT).Mutable()
}
