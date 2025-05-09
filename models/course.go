package models

import (
	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course defines the model for a course
type Course struct {
	Base
	Title     string
	Path      string
	CardPath  string
	Available bool
	Duration  int

	// Joins
	ScanStatus types.ScanStatus

	// Relations
	Progress *CourseProgress
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	COURSE_TABLE       = "courses"
	COURSE_TITLE       = "title"
	COURSE_PATH        = "path"
	COURSE_CARD_PATH   = "card_path"
	COURSE_AVAILABLE   = "available"
	COURSE_DURATION    = "duration"
	COURSE_SCAN_STATUS = "status"

	COURSE_TABLE_ID          = COURSE_TABLE + "." + BASE_ID
	COURSE_TABLE_CREATED_AT  = COURSE_TABLE + "." + BASE_CREATED_AT
	COURSE_TABLE_UPDATED_AT  = COURSE_TABLE + "." + BASE_UPDATED_AT
	COURSE_TABLE_TITLE       = COURSE_TABLE + "." + COURSE_TITLE
	COURSE_TABLE_PATH        = COURSE_TABLE + "." + COURSE_PATH
	COURSE_TABLE_CARD_PATH   = COURSE_TABLE + "." + COURSE_CARD_PATH
	COURSE_TABLE_AVAILABLE   = COURSE_TABLE + "." + COURSE_AVAILABLE
	COURSE_TABLE_DURATION    = COURSE_TABLE + "." + COURSE_DURATION
	COURSE_TABLE_SCAN_STATUS = COURSE_TABLE + "." + COURSE_SCAN_STATUS
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (c *Course) Table() string {
	return COURSE_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (c *Course) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("Title").Column(COURSE_TITLE).NotNull()
	s.Field("Path").Column(COURSE_PATH).NotNull()
	s.Field("CardPath").Column(COURSE_CARD_PATH).Mutable()
	s.Field("Available").Column(COURSE_AVAILABLE).Mutable()
	s.Field("Duration").Column(COURSE_DURATION).Mutable()

	// Join fields
	s.Field("ScanStatus").JoinTable(SCAN_TABLE).Column(COURSE_SCAN_STATUS).Alias("scan_status")

	// Relation fields
	s.Relation("Progress").MatchOn(COURSE_PROGRESS_COURSE_ID)

	// Joins
	s.LeftJoin(SCAN_TABLE).On(COURSE_TABLE_ID + " = " + SCAN_TABLE_COURSE_ID)
}
