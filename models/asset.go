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
	AssetMetadata *AssetMetadata
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetRow is for use in scanning a full asset with optional joins
type AssetRow struct {
	// Base asset columns (match assets.*)
	ID        string         `db:"id"`
	CourseID  string         `db:"course_id"`
	LessonID  string         `db:"lesson_id"`
	Title     string         `db:"title"`
	Prefix    sql.NullInt16  `db:"prefix"`
	SubPrefix sql.NullInt16  `db:"sub_prefix"`
	SubTitle  string         `db:"sub_title"`
	Module    string         `db:"module"`
	Type      types.Asset    `db:"type"`
	Path      string         `db:"path"`
	FileSize  int64          `db:"file_size"`
	ModTime   string         `db:"mod_time"`
	Hash      string         `db:"hash"`
	CreatedAt types.DateTime `db:"created_at"`
	UpdatedAt types.DateTime `db:"updated_at"`

	// Optional progress columns (nullable, appear only if you include the join)
	ProgressVideoPos    sql.NullInt64  `db:"video_pos"`
	ProgressCompleted   sql.NullBool   `db:"completed"`
	ProgressCompletedAt types.DateTime `db:"completed_at"`

	// Optional metadata columns (nullable, appear only if you include the joins)
	// Reuse the existing scan row for metadata.
	AssetMetadataRow
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ToDomain maps the row to the domain model. includeProgress and includeMetadata indicate
// whether those relations were included in the query, and thus should be mapped.
func (r *AssetRow) ToDomain(includeProgress, includeMetadata bool) *Asset {
	a := &Asset{
		Base: Base{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		CourseID:  r.CourseID,
		LessonID:  r.LessonID,
		Title:     r.Title,
		Prefix:    r.Prefix,
		SubPrefix: r.SubPrefix,
		SubTitle:  r.SubTitle,
		Module:    r.Module,
		Type:      r.Type,
		Path:      r.Path,
		FileSize:  r.FileSize,
		ModTime:   r.ModTime,
		Hash:      r.Hash,
	}

	// Attach progress if requested
	if includeProgress {
		a.Progress = &AssetProgressInfo{
			VideoPos:    int(r.ProgressVideoPos.Int64),
			Completed:   r.ProgressCompleted.Bool,
			CompletedAt: r.ProgressCompletedAt,
		}
	}

	// Attach metadata if requested
	if includeMetadata {
		meta := r.AssetMetadataRow.ToDomain()
		// always attach an object, even if empty
		a.AssetMetadata = meta
	}

	return a
}
