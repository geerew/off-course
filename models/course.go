package models

import (
	"database/sql"

	"github.com/geerew/off-course/utils/types"
)

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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgressInfo defines progress information for a course
func CourseProgressJoinColumns() []string {
	return []string{
		COURSE_PROGRESS_TABLE_STARTED + " AS prog_started",
		COURSE_PROGRESS_TABLE_STARTED_AT + " AS prog_started_at",
		COURSE_PROGRESS_TABLE_PERCENT + " AS prog_percent",
		COURSE_PROGRESS_TABLE_COMPLETED_AT + " AS prog_completed_at",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseRow is for use in scanning a full course with optional joins
type CourseRow struct {
	ID          string         `db:"id"`
	Title       string         `db:"title"`
	Path        string         `db:"path"`
	CardPath    string         `db:"card_path"`
	Available   bool           `db:"available"`
	Duration    int            `db:"duration"`
	InitialScan bool           `db:"initial_scan"`
	Maintenance bool           `db:"maintenance"`
	CreatedAt   types.DateTime `db:"created_at"`
	UpdatedAt   types.DateTime `db:"updated_at"`

	// Progress
	ProgStarted     sql.NullBool   `db:"prog_started"`
	ProgStartedAt   types.DateTime `db:"prog_started_at"`
	ProgPercent     sql.NullInt64  `db:"prog_percent"`
	ProgCompletedAt types.DateTime `db:"prog_completed_at"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ToDomain maps the row to the domain model. includeProgress indicates whether
// whether that relation was included in the query, and should be mapped
func (r *CourseRow) ToDomain() *Course {
	c := &Course{
		Base: Base{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		Title:       r.Title,
		Path:        r.Path,
		CardPath:    r.CardPath,
		Available:   r.Available,
		Duration:    r.Duration,
		InitialScan: r.InitialScan,
		Maintenance: r.Maintenance,
	}

	c.Progress = &CourseProgressInfo{
		Started:     r.ProgStarted.Bool,
		StartedAt:   r.ProgStartedAt,
		Percent:     int(r.ProgPercent.Int64),
		CompletedAt: r.ProgCompletedAt,
	}

	return c
}
