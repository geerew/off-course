package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course defines the model for a course
type Course struct {
	Base
	Title       string `db:"title"`        // Mutable
	Path        string `db:"path"`         // Mutable
	CardPath    string `db:"card_path"`    // Mutable
	Available   bool   `db:"available"`    // Mutable
	Duration    int    `db:"duration"`     // Mutable
	InitialScan bool   `db:"initial_scan"` // Mutable
	Maintenance bool   `db:"maintenance"`  // Mutable

	// Relation
	Progress *CourseProgressInfo
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	COURSE_TABLE        = "courses"
	COURSE_TITLE        = "title"
	COURSE_PATH         = "path"
	COURSE_CARD_PATH    = "card_path"
	COURSE_AVAILABLE    = "available"
	COURSE_DURATION     = "duration"
	COURSE_INITIAL_SCAN = "initial_scan"
	COURSE_MAINTENANCE  = "maintenance"

	COURSE_TABLE_ID           = COURSE_TABLE + "." + BASE_ID
	COURSE_TABLE_CREATED_AT   = COURSE_TABLE + "." + BASE_CREATED_AT
	COURSE_TABLE_UPDATED_AT   = COURSE_TABLE + "." + BASE_UPDATED_AT
	COURSE_TABLE_TITLE        = COURSE_TABLE + "." + COURSE_TITLE
	COURSE_TABLE_PATH         = COURSE_TABLE + "." + COURSE_PATH
	COURSE_TABLE_CARD_PATH    = COURSE_TABLE + "." + COURSE_CARD_PATH
	COURSE_TABLE_AVAILABLE    = COURSE_TABLE + "." + COURSE_AVAILABLE
	COURSE_TABLE_DURATION     = COURSE_TABLE + "." + COURSE_DURATION
	COURSE_TABLE_INITIAL_SCAN = COURSE_TABLE + "." + COURSE_INITIAL_SCAN
	COURSE_TABLE_MAINTENANCE  = COURSE_TABLE + "." + COURSE_MAINTENANCE
)
