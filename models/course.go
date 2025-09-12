package models

import (
	"fmt"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	COURSE_TABLE = "courses"

	COURSE_TITLE        = "title"
	COURSE_PATH         = "path"
	COURSE_CARD_PATH    = "card_path"
	COURSE_AVAILABLE    = "available"
	COURSE_DURATION     = "duration"
	COURSE_ASSETS_COUNT = "assets_count"
	COURSE_TOTAL_WEIGHT = "total_weight"
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
	COURSE_TABLE_ASSETS_COUNT = COURSE_TABLE + "." + COURSE_ASSETS_COUNT
	COURSE_TABLE_TOTAL_WEIGHT = COURSE_TABLE + "." + COURSE_TOTAL_WEIGHT
	COURSE_TABLE_INITIAL_SCAN = COURSE_TABLE + "." + COURSE_INITIAL_SCAN
	COURSE_TABLE_MAINTENANCE  = COURSE_TABLE + "." + COURSE_MAINTENANCE
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
	AssetsCount int    `db:"assets_count"` // Mutable
	TotalWeight int    `db:"total_weight"` // Mutable
	InitialScan bool   `db:"initial_scan"` // Mutable
	Maintenance bool   `db:"maintenance"`  // Mutable

	// Relation
	Progress *CourseProgress `db:"-"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseColumns returns the list of columns to use when populating `Course`
func CourseColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", COURSE_TABLE_ID),
		fmt.Sprintf("%s AS created_at", COURSE_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS updated_at", COURSE_TABLE_UPDATED_AT),
		fmt.Sprintf("%s AS title", COURSE_TABLE_TITLE),
		fmt.Sprintf("%s AS path", COURSE_TABLE_PATH),
		fmt.Sprintf("%s AS card_path", COURSE_TABLE_CARD_PATH),
		fmt.Sprintf("%s AS available", COURSE_TABLE_AVAILABLE),
		fmt.Sprintf("%s AS duration", COURSE_TABLE_DURATION),
		fmt.Sprintf("%s AS assets_count", COURSE_TABLE_ASSETS_COUNT),
		fmt.Sprintf("%s AS total_weight", COURSE_TABLE_TOTAL_WEIGHT),
		fmt.Sprintf("%s AS initial_scan", COURSE_TABLE_INITIAL_SCAN),
		fmt.Sprintf("%s AS maintenance", COURSE_TABLE_MAINTENANCE),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseRow is for use in scanning a full course with optional relations
type CourseRow struct {
	Course

	// Progress
	CourseProgressRow
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ToDomain converts CourseRow to Course
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
		AssetsCount: r.AssetsCount,
		TotalWeight: r.TotalWeight,
		InitialScan: r.InitialScan,
		Maintenance: r.Maintenance,
	}

	c.Progress = r.CourseProgressRow.ToDomain()

	return c
}
