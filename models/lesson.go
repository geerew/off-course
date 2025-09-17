package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"fmt"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	LESSON_TABLE = "lessons"

	LESSON_COURSE_ID = "course_id"
	LESSON_TITLE     = "title"
	LESSON_PREFIX    = "prefix"
	LESSON_MODULE    = "module"

	LESSON_TABLE_ID         = LESSON_TABLE + "." + BASE_ID
	LESSON_TABLE_CREATED_AT = LESSON_TABLE + "." + BASE_CREATED_AT
	LESSON_TABLE_UPDATED_AT = LESSON_TABLE + "." + BASE_UPDATED_AT
	LESSON_TABLE_COURSE_ID  = LESSON_TABLE + "." + LESSON_COURSE_ID
	LESSON_TABLE_TITLE      = LESSON_TABLE + "." + LESSON_TITLE
	LESSON_TABLE_PREFIX     = LESSON_TABLE + "." + LESSON_PREFIX
	LESSON_TABLE_MODULE     = LESSON_TABLE + "." + LESSON_MODULE
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

// LessonColumns returns the list of columns to use when populating `Lesson`
func LessonColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", LESSON_TABLE_ID),
		fmt.Sprintf("%s AS created_at", LESSON_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS updated_at", LESSON_TABLE_UPDATED_AT),
		fmt.Sprintf("%s AS course_id", LESSON_TABLE_COURSE_ID),
		fmt.Sprintf("%s AS title", LESSON_TABLE_TITLE),
		fmt.Sprintf("%s AS prefix", LESSON_TABLE_PREFIX),
		fmt.Sprintf("%s AS module", LESSON_TABLE_MODULE),
	}
}
