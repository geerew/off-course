package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment defines the model for an attachment
type Attachment struct {
	Base
	LessonID string `db:"lesson_id"` // Immutable
	Title    string `db:"title"`     // Mutable
	Path     string `db:"path"`      // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	ATTACHMENT_TABLE     = "attachments"
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
