package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"

	"github.com/geerew/off-course/utils/schema"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset defines the model for an asset
type Asset struct {
	Base
	CourseID        string
	Title           string
	Prefix          sql.NullInt16
	SubPrefix       sql.NullInt16
	SubTitle        string
	Chapter         string
	Type            types.Asset
	Path            string
	FileSize        int64
	ModTime         string
	Hash            string
	DescriptionPath string
	DescriptionType types.Description

	// Relations
	VideoMetadata *VideoMetadata
	Progress      *AssetProgress
	Attachments   []*Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	ASSET_TABLE            = "assets"
	ASSET_COURSE_ID        = "course_id"
	ASSET_TITLE            = "title"
	ASSET_PREFIX           = "prefix"
	ASSET_SUB_PREFIX       = "sub_prefix"
	ASSET_SUB_TITLE        = "sub_title"
	ASSET_CHAPTER          = "chapter"
	ASSET_TYPE             = "type"
	ASSET_PATH             = "path"
	ASSET_FILE_SIZE        = "file_size"
	ASSET_MOD_TIME         = "mod_time"
	ASSET_HASH             = "hash"
	ASSET_VIDEO_POSITION   = "video_pos"
	ASSET_COMPLETED        = "completed"
	ASSET_COMPLETED_AT     = "completed_at"
	ASSET_DESCRIPTION_PATH = "description_path"
	ASSET_DESCRIPTION_TYPE = "description_type"

	ASSET_TABLE_ID               = ASSET_TABLE + "." + BASE_ID
	ASSET_TABLE_CREATED_AT       = ASSET_TABLE + "." + BASE_CREATED_AT
	ASSET_TABLE_UPDATED_AT       = ASSET_TABLE + "." + BASE_UPDATED_AT
	ASSET_TABLE_COURSE_ID        = ASSET_TABLE + "." + ASSET_COURSE_ID
	ASSET_TABLE_TITLE            = ASSET_TABLE + "." + ASSET_TITLE
	ASSET_TABLE_PREFIX           = ASSET_TABLE + "." + ASSET_PREFIX
	ASSET_TABLE_SUB_PREFIX       = ASSET_TABLE + "." + ASSET_SUB_PREFIX
	ASSET_TABLE_CHAPTER          = ASSET_TABLE + "." + ASSET_CHAPTER
	ASSET_TABLE_TYPE             = ASSET_TABLE + "." + ASSET_TYPE
	ASSET_TABLE_PATH             = ASSET_TABLE + "." + ASSET_PATH
	ASSET_TABLE_HASH             = ASSET_TABLE + "." + ASSET_HASH
	ASSET_TABLE_VIDEO_POS        = ASSET_TABLE + "." + ASSET_VIDEO_POSITION
	ASSET_TABLE_COMPLETED        = ASSET_TABLE + "." + ASSET_COMPLETED
	ASSET_TABLE_COMPLETED_AT     = ASSET_TABLE + "." + ASSET_COMPLETED_AT
	ASSET_TABLE_DESCRIPTION_PATH = ASSET_TABLE + "." + ASSET_DESCRIPTION_PATH
	ASSET_TABLE_DESCRIPTION_TYPE = ASSET_TABLE + "." + ASSET_DESCRIPTION_TYPE

	ASSET_RELATION_PROGRESS = "Progress"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Table implements the `schema.Modeler` interface by returning the table name
func (a *Asset) Table() string {
	return ASSET_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields implements the `schema.Modeler` interface by defining the model fields
func (a *Asset) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	// Common fields
	s.Field("CourseID").Column(ASSET_COURSE_ID).NotNull()
	s.Field("Title").Column(ASSET_TITLE).NotNull().Mutable()
	s.Field("Prefix").Column(ASSET_PREFIX).Mutable()
	s.Field("SubPrefix").Column(ASSET_SUB_PREFIX).Mutable()
	s.Field("SubTitle").Column(ASSET_SUB_TITLE).Mutable()
	s.Field("Chapter").Column(ASSET_CHAPTER).Mutable()
	s.Field("Type").Column(ASSET_TYPE).NotNull().Mutable()
	s.Field("Path").Column(ASSET_PATH).NotNull().Mutable()
	s.Field("FileSize").Column(ASSET_FILE_SIZE).NotNull().Mutable()
	s.Field("ModTime").Column(ASSET_MOD_TIME).NotNull().Mutable()
	s.Field("Hash").Column(ASSET_HASH).NotNull().Mutable()
	s.Field("DescriptionPath").Column(ASSET_DESCRIPTION_PATH).Mutable()
	s.Field("DescriptionType").Column(ASSET_DESCRIPTION_TYPE).Mutable()

	// Relation fields
	s.Relation("VideoMetadata").MatchOn(VIDEO_METADATA_ASSET_ID)
	s.Relation("Progress").MatchOn(ASSET_PROGRESS_ASSET_ID)
	s.Relation("Attachments").MatchOn(ATTACHMENT_ASSET_ID)
}
