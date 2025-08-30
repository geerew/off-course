package models

import (
	"fmt"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgress defines the model for an asset progress
type AssetProgress struct {
	Base
	AssetID string `db:"asset_id"` // Immutable
	UserID  string `db:"user_id"`  // Immutable
	AssetProgressInfo
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgressInfo defines a minified AssetProgress model
type AssetProgressInfo struct {
	VideoPos    int            `db:"video_pos"`    // Mutable
	Completed   bool           `db:"completed"`    // Mutable
	CompletedAt types.DateTime `db:"completed_at"` // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	ASSET_PROGRESS_TABLE        = "assets_progress"
	ASSET_PROGRESS_ASSET_ID     = "asset_id"
	ASSET_PROGRESS_USER_ID      = "user_id"
	ASSET_PROGRESS_VIDEO_POS    = "video_pos"
	ASSET_PROGRESS_COMPLETED    = "completed"
	ASSET_PROGRESS_COMPLETED_AT = "completed_at"

	ASSET_PROGRESS_TABLE_ID           = ASSET_PROGRESS_TABLE + "." + BASE_ID
	ASSET_PROGRESS_TABLE_CREATED_AT   = ASSET_PROGRESS_TABLE + "." + BASE_CREATED_AT
	ASSET_PROGRESS_TABLE_UPDATED_AT   = ASSET_PROGRESS_TABLE + "." + BASE_UPDATED_AT
	ASSET_PROGRESS_TABLE_ASSET_ID     = ASSET_PROGRESS_TABLE + "." + ASSET_PROGRESS_ASSET_ID
	ASSET_PROGRESS_TABLE_USER_ID      = ASSET_PROGRESS_TABLE + "." + ASSET_PROGRESS_USER_ID
	ASSET_PROGRESS_TABLE_VIDEO_POS    = ASSET_PROGRESS_TABLE + "." + ASSET_PROGRESS_VIDEO_POS
	ASSET_PROGRESS_TABLE_COMPLETED    = ASSET_PROGRESS_TABLE + "." + ASSET_PROGRESS_COMPLETED
	ASSET_PROGRESS_TABLE_COMPLETED_AT = ASSET_PROGRESS_TABLE + "." + ASSET_PROGRESS_COMPLETED_AT
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgressJoinColumns returns the columns to join asset progress
func AssetProgressJoinColumns() []string {
	return []string{
		fmt.Sprintf("%s AS video_pos", ASSET_PROGRESS_TABLE_VIDEO_POS),
		fmt.Sprintf("%s AS completed", ASSET_PROGRESS_TABLE_COMPLETED),
		fmt.Sprintf("%s AS completed_at", ASSET_PROGRESS_TABLE_COMPLETED_AT),
	}
}
