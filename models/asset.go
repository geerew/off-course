package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"fmt"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	ASSET_TABLE = "assets"

	ASSET_COURSE_ID  = "course_id"
	ASSET_LESSON_ID  = "lesson_id"
	ASSET_TITLE      = "title"
	ASSET_PREFIX     = "prefix"
	ASSET_SUB_PREFIX = "sub_prefix"
	ASSET_SUB_TITLE  = "sub_title"
	ASSET_MODULE     = "module"
	ASSET_TYPE       = "type"
	ASSET_PATH       = "path"
	ASSET_FILE_SIZE  = "file_size"
	ASSET_MOD_TIME   = "mod_time"
	ASSET_HASH       = "hash"
	ASSET_WEIGHT     = "weight"

	ASSET_TABLE_ID         = ASSET_TABLE + "." + BASE_ID
	ASSET_TABLE_CREATED_AT = ASSET_TABLE + "." + BASE_CREATED_AT
	ASSET_TABLE_UPDATED_AT = ASSET_TABLE + "." + BASE_UPDATED_AT
	ASSET_TABLE_COURSE_ID  = ASSET_TABLE + "." + ASSET_COURSE_ID
	ASSET_TABLE_LESSON_ID  = ASSET_TABLE + "." + ASSET_LESSON_ID
	ASSET_TABLE_TITLE      = ASSET_TABLE + "." + ASSET_TITLE
	ASSET_TABLE_PREFIX     = ASSET_TABLE + "." + ASSET_PREFIX
	ASSET_TABLE_SUB_PREFIX = ASSET_TABLE + "." + ASSET_SUB_PREFIX
	ASSET_TABLE_SUB_TITLE  = ASSET_TABLE + "." + ASSET_SUB_TITLE
	ASSET_TABLE_MODULE     = ASSET_TABLE + "." + ASSET_MODULE
	ASSET_TABLE_TYPE       = ASSET_TABLE + "." + ASSET_TYPE
	ASSET_TABLE_PATH       = ASSET_TABLE + "." + ASSET_PATH
	ASSET_TABLE_FILE_SIZE  = ASSET_TABLE + "." + ASSET_FILE_SIZE
	ASSET_TABLE_MOD_TIME   = ASSET_TABLE + "." + ASSET_MOD_TIME
	ASSET_TABLE_HASH       = ASSET_TABLE + "." + ASSET_HASH
	ASSET_TABLE_WEIGHT     = ASSET_TABLE + "." + ASSET_WEIGHT
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
	Weight    int           `db:"weight"`

	// Relations
	AssetMetadata *AssetMetadata `db:"-"`
	Progress      *AssetProgress `db:"-"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetRow is for use in scanning a full asset with optional relations
type AssetRow struct {
	// Base asset columns (match assets.*)
	Asset

	// Optional progress columns
	AssetProgressRow

	// Optional metadata columns
	AssetMetadataRow
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ToDomain converts AssetRow to Asset. includeProgress and includeMetadata indicate
// whether those relations were included in the query, and thus should be mapped
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
		Weight:    r.Weight,
	}

	// Attach progress if requested
	if includeProgress {
		a.Progress = r.AssetProgressRow.ToDomain()
	}

	// Attach metadata if requested
	if includeMetadata {
		a.AssetMetadata = r.AssetMetadataRow.ToDomain()
	}

	return a
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetColumns returns the list of columns to use when populating `Asset`
func AssetColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", ASSET_TABLE_ID),
		fmt.Sprintf("%s AS created_at", ASSET_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS updated_at", ASSET_TABLE_UPDATED_AT),
		fmt.Sprintf("%s AS course_id", ASSET_TABLE_COURSE_ID),
		fmt.Sprintf("%s AS lesson_id", ASSET_TABLE_LESSON_ID),
		fmt.Sprintf("%s AS title", ASSET_TABLE_TITLE),
		fmt.Sprintf("%s AS prefix", ASSET_TABLE_PREFIX),
		fmt.Sprintf("%s AS sub_prefix", ASSET_TABLE_SUB_PREFIX),
		fmt.Sprintf("%s AS sub_title", ASSET_TABLE_SUB_TITLE),
		fmt.Sprintf("%s AS module", ASSET_TABLE_MODULE),
		fmt.Sprintf("%s AS type", ASSET_TABLE_TYPE),
		fmt.Sprintf("%s AS path", ASSET_TABLE_PATH),
		fmt.Sprintf("%s AS file_size", ASSET_TABLE_FILE_SIZE),
		fmt.Sprintf("%s AS mod_time", ASSET_TABLE_MOD_TIME),
		fmt.Sprintf("%s AS hash", ASSET_TABLE_HASH),
		fmt.Sprintf("%s AS weight", ASSET_TABLE_WEIGHT),
	}
}
