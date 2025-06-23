package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"

	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetGroups defines the model for an asset group
type AssetGroup struct {
	Base
	CourseID        string
	Title           string
	Prefix          sql.NullInt16
	Module          string
	DescriptionPath string
	DescriptionType types.Description

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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (ag *AssetGroup) Table() string {
	return ASSET_GROUP_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (ag *AssetGroup) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("CourseID").Column(ASSET_GROUP_COURSE_ID).NotNull()
	s.Field("Title").Column(ASSET_GROUP_TITLE).NotNull().Mutable()
	s.Field("Prefix").Column(ASSET_GROUP_PREFIX).Mutable()
	s.Field("Module").Column(ASSET_GROUP_MODULE).Mutable()
	s.Field("DescriptionPath").Column(ASSET_GROUP_DESCRIPTION_PATH).Mutable()
	s.Field("DescriptionType").Column(ASSET_GROUP_DESCRIPTION_TYPE).Mutable()

	// Relation fields
	s.Relation("Assets").MatchOn(ASSET_ASSET_GROUP_ID)
	s.Relation("Attachments").MatchOn(ATTACHMENT_ASSET_GROUP_ID)
}
