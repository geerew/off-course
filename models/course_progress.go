package models

import (
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgress defines the model for a course progress
type CourseProgress struct {
	Base
	CourseID string `db:"course_id"` // Immutable
	UserID   string `db:"user_id"`   // Immutable
	CourseProgressInfo
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseProgressInfo defines the progress information for a course
type CourseProgressInfo struct {
	Started     bool           `db:"started"`      // Mutable
	StartedAt   types.DateTime `db:"started_at"`   // Mutable
	Percent     int            `db:"percent"`      // Mutable
	CompletedAt types.DateTime `db:"completed_at"` // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	COURSE_PROGRESS_TABLE        = "courses_progress"
	COURSE_PROGRESS_COURSE_ID    = "course_id"
	COURSE_PROGRESS_USER_ID      = "user_id"
	COURSE_PROGRESS_STARTED      = "started"
	COURSE_PROGRESS_STARTED_AT   = "started_at"
	COURSE_PROGRESS_PERCENT      = "percent"
	COURSE_PROGRESS_COMPLETED_AT = "completed_at"

	COURSE_PROGRESS_TABLE_ID           = COURSE_PROGRESS_TABLE + "." + BASE_ID
	COURSE_PROGRESS_TABLE_CREATED_AT   = COURSE_PROGRESS_TABLE + "." + BASE_CREATED_AT
	COURSE_PROGRESS_TABLE_UPDATED_AT   = COURSE_PROGRESS_TABLE + "." + BASE_UPDATED_AT
	COURSE_PROGRESS_TABLE_COURSE_ID    = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_COURSE_ID
	COURSE_PROGRESS_TABLE_USER_ID      = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_USER_ID
	COURSE_PROGRESS_TABLE_STARTED      = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_STARTED
	COURSE_PROGRESS_TABLE_STARTED_AT   = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_STARTED_AT
	COURSE_PROGRESS_TABLE_PERCENT      = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_PERCENT
	COURSE_PROGRESS_TABLE_COMPLETED_AT = COURSE_PROGRESS_TABLE + "." + COURSE_PROGRESS_COMPLETED_AT

	COURSE_RELATION_PROGRESS = "Progress"
)
