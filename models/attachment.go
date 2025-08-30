package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment defines the model for an attachment
type Attachment struct {
	Base
	AssetGroupID string `db:"asset_group_id"` // Immutable
	Title        string `db:"title"`          // Mutable
	Path         string `db:"path"`           // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	ATTACHMENT_TABLE          = "attachments"
	ATTACHMENT_ASSET_GROUP_ID = "asset_group_id"
	ATTACHMENT_TITLE          = "title"
	ATTACHMENT_PATH           = "path"

	ATTACHMENT_TABLE_ID             = ATTACHMENT_TABLE + "." + BASE_ID
	ATTACHMENT_TABLE_CREATED_AT     = ATTACHMENT_TABLE + "." + BASE_CREATED_AT
	ATTACHMENT_TABLE_UPDATED_AT     = ATTACHMENT_TABLE + "." + BASE_UPDATED_AT
	ATTACHMENT_TABLE_ASSET_GROUP_ID = ATTACHMENT_TABLE + "." + ATTACHMENT_ASSET_GROUP_ID
	ATTACHMENT_TABLE_TITLE          = ATTACHMENT_TABLE + "." + ATTACHMENT_TITLE
	ATTACHMENT_TABLE_PATH           = ATTACHMENT_TABLE + "." + ATTACHMENT_PATH
)
