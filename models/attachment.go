package models

import "github.com/geerew/off-course/utils/schema"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment defines the model for an attachment
type Attachment struct {
	Base
	AssetID string
	Title   string
	Path    string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	ATTACHMENT_TABLE    = "attachments"
	ATTACHMENT_ASSET_ID = "asset_id"
	ATTACHMENT_TITLE    = "title"
	ATTACHMENT_PATH     = "path"

	ATTACHMENT_TABLE_ID         = ATTACHMENT_TABLE + "." + BASE_ID
	ATTACHMENT_TABLE_CREATED_AT = ATTACHMENT_TABLE + "." + BASE_CREATED_AT
	ATTACHMENT_TABLE_UPDATED_AT = ATTACHMENT_TABLE + "." + BASE_UPDATED_AT
	ATTACHMENT_TABLE_ASSET_ID   = ATTACHMENT_TABLE + "." + ATTACHMENT_ASSET_ID
	ATTACHMENT_TABLE_TITLE      = ATTACHMENT_TABLE + "." + ATTACHMENT_TITLE
	ATTACHMENT_TABLE_PATH       = ATTACHMENT_TABLE + "." + ATTACHMENT_PATH
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (a *Attachment) Table() string {
	return ATTACHMENT_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (a *Attachment) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	s.Field("AssetID").Column(ATTACHMENT_ASSET_ID).NotNull()
	s.Field("Title").Column(ATTACHMENT_TITLE).NotNull().Mutable()
	s.Field("Path").Column(ATTACHMENT_PATH).NotNull().Mutable()
}
