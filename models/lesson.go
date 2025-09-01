package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Lesson defines the model for a lesson
type Lesson struct {
	Base
	CourseID string        `db:"course_id"` // Immutable
	Title    string        `db:"title"`     // Mutable
	Prefix   sql.NullInt16 `db:"prefix"`    // Mutable
	Module   string        `db:"module"`    // Mutable

	// Relations
	Assets      []*Asset
	Attachments []*Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	LESSON_TABLE            = "lessons"
	LESSON_COURSE_ID        = "course_id"
	LESSON_TITLE            = "title"
	LESSON_PREFIX           = "prefix"
	LESSON_MODULE           = "module"
	LESSON_DESCRIPTION_PATH = "description_path"
	LESSON_DESCRIPTION_TYPE = "description_type"

	LESSON_TABLE_ID               = LESSON_TABLE + "." + BASE_ID
	LESSON_TABLE_CREATED_AT       = LESSON_TABLE + "." + BASE_CREATED_AT
	LESSON_TABLE_UPDATED_AT       = LESSON_TABLE + "." + BASE_UPDATED_AT
	LESSON_TABLE_COURSE_ID        = LESSON_TABLE + "." + LESSON_COURSE_ID
	LESSON_TABLE_TITLE            = LESSON_TABLE + "." + LESSON_TITLE
	LESSON_TABLE_PREFIX           = LESSON_TABLE + "." + LESSON_PREFIX
	LESSON_TABLE_MODULE           = LESSON_TABLE + "." + LESSON_MODULE
	LESSON_TABLE_DESCRIPTION_PATH = LESSON_TABLE + "." + LESSON_DESCRIPTION_PATH
	LESSON_TABLE_DESCRIPTION_TYPE = LESSON_TABLE + "." + LESSON_DESCRIPTION_TYPE
)
