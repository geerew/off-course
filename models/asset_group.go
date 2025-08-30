package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetGroups defines the model for an asset group
type AssetGroup struct {
	Base
	CourseID        string            `db:"course_id"`        // Immutable
	Title           string            `db:"title"`            // Mutable
	Prefix          sql.NullInt16     `db:"prefix"`           // Mutable
	Module          string            `db:"module"`           // Mutable
	DescriptionPath string            `db:"description_path"` // Mutable
	DescriptionType types.Description `db:"description_type"` // Mutable

	// Relations
	Assets      []*Asset
	Attachments []*Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	ASSET_GROUP_TABLE            = "asset_groups"
	ASSET_GROUP_COURSE_ID        = "course_id"
	ASSET_GROUP_TITLE            = "title"
	ASSET_GROUP_PREFIX           = "prefix"
	ASSET_GROUP_MODULE           = "module"
	ASSET_GROUP_DESCRIPTION_PATH = "description_path"
	ASSET_GROUP_DESCRIPTION_TYPE = "description_type"

	ASSET_GROUP_TABLE_ID               = ASSET_GROUP_TABLE + "." + BASE_ID
	ASSET_GROUP_TABLE_CREATED_AT       = ASSET_GROUP_TABLE + "." + BASE_CREATED_AT
	ASSET_GROUP_TABLE_UPDATED_AT       = ASSET_GROUP_TABLE + "." + BASE_UPDATED_AT
	ASSET_GROUP_TABLE_COURSE_ID        = ASSET_GROUP_TABLE + "." + ASSET_GROUP_COURSE_ID
	ASSET_GROUP_TABLE_TITLE            = ASSET_GROUP_TABLE + "." + ASSET_GROUP_TITLE
	ASSET_GROUP_TABLE_PREFIX           = ASSET_GROUP_TABLE + "." + ASSET_GROUP_PREFIX
	ASSET_GROUP_TABLE_MODULE           = ASSET_GROUP_TABLE + "." + ASSET_GROUP_MODULE
	ASSET_GROUP_TABLE_DESCRIPTION_PATH = ASSET_GROUP_TABLE + "." + ASSET_GROUP_DESCRIPTION_PATH
	ASSET_GROUP_TABLE_DESCRIPTION_TYPE = ASSET_GROUP_TABLE + "." + ASSET_GROUP_DESCRIPTION_TYPE
)
