package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset defines the model for an asset
type Asset struct {
	Base
	CourseID  string        `db:"course_id"`  // Immutable
	LessonID  string        `db:"lesson_id"`  // Mutable
	Title     string        `db:"title"`      // Mutable
	Prefix    sql.NullInt16 `db:"prefix"`     // Mutable
	SubPrefix sql.NullInt16 `db:"sub_prefix"` // Mutable
	SubTitle  string        `db:"sub_title"`  // Mutable
	Module    string        `db:"module"`     // Mutable
	Type      types.Asset   `db:"type"`       // Mutable
	Path      string        `db:"path"`       // Mutable
	FileSize  int64         `db:"file_size"`  // Mutable
	ModTime   string        `db:"mod_time"`   // Mutable
	Hash      string        `db:"hash"`       // Mutable

	// Relations
	VideoMetadata *VideoMetadataInfo
	Progress      *AssetProgressInfo
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	ASSET_TABLE            = "assets"
	ASSET_COURSE_ID        = "course_id"
	ASSET_LESSON_ID        = "lesson_id"
	ASSET_TITLE            = "title"
	ASSET_PREFIX           = "prefix"
	ASSET_SUB_PREFIX       = "sub_prefix"
	ASSET_SUB_TITLE        = "sub_title"
	ASSET_MODULE           = "module"
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
	ASSET_TABLE_LESSON_ID        = ASSET_TABLE + "." + ASSET_LESSON_ID
	ASSET_TABLE_TITLE            = ASSET_TABLE + "." + ASSET_TITLE
	ASSET_TABLE_PREFIX           = ASSET_TABLE + "." + ASSET_PREFIX
	ASSET_TABLE_SUB_PREFIX       = ASSET_TABLE + "." + ASSET_SUB_PREFIX
	ASSET_TABLE_MODULE           = ASSET_TABLE + "." + ASSET_MODULE
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
