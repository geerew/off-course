package models

import "fmt"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	ATTACHMENT_TABLE = "attachments"

	ATTACHMENT_LESSON_ID = "lesson_id"
	ATTACHMENT_TITLE     = "title"
	ATTACHMENT_PATH      = "path"

	ATTACHMENT_TABLE_ID         = ATTACHMENT_TABLE + "." + BASE_ID
	ATTACHMENT_TABLE_CREATED_AT = ATTACHMENT_TABLE + "." + BASE_CREATED_AT
	ATTACHMENT_TABLE_UPDATED_AT = ATTACHMENT_TABLE + "." + BASE_UPDATED_AT
	ATTACHMENT_TABLE_LESSON_ID  = ATTACHMENT_TABLE + "." + ATTACHMENT_LESSON_ID
	ATTACHMENT_TABLE_TITLE      = ATTACHMENT_TABLE + "." + ATTACHMENT_TITLE
	ATTACHMENT_TABLE_PATH       = ATTACHMENT_TABLE + "." + ATTACHMENT_PATH
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment defines the model for an attachment
type Attachment struct {
	Base
	LessonID string `db:"lesson_id"` // Immutable
	Title    string `db:"title"`     // Mutable
	Path     string `db:"path"`      // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AttachmentColumns returns the list of columns to use when populating `Attachment`
func AttachmentColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", ATTACHMENT_TABLE_ID),
		fmt.Sprintf("%s AS created_at", ATTACHMENT_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS updated_at", ATTACHMENT_TABLE_UPDATED_AT),
		fmt.Sprintf("%s AS lesson_id", ATTACHMENT_TABLE_LESSON_ID),
		fmt.Sprintf("%s AS title", ATTACHMENT_TABLE_TITLE),
		fmt.Sprintf("%s AS path", ATTACHMENT_TABLE_PATH),
	}
}
