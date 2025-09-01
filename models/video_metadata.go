package models

import (
	"fmt"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoMetadata defines video-related metadata for an asset
type VideoMetadata struct {
	Base
	AssetID string `db:"asset_id"` // Immutable
	VideoMetadataInfo
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type VideoMetadataInfo struct {
	Duration   int    `db:"duration"`   // Mutable
	Width      int    `db:"width"`      // Mutable
	Height     int    `db:"height"`     // Mutable
	Resolution string `db:"resolution"` // Mutable
	Codec      string `db:"codec"`      // Mutable
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	VIDEO_METADATA_TABLE      = "asset_video_metadata"
	VIDEO_METADATA_ASSET_ID   = "asset_id"
	VIDEO_METADATA_DURATION   = "duration"
	VIDEO_METADATA_WIDTH      = "width"
	VIDEO_METADATA_HEIGHT     = "height"
	VIDEO_METADATA_RESOLUTION = "resolution"
	VIDEO_METADATA_CODEC      = "codec"

	VIDEO_METADATA_TABLE_ID         = VIDEO_METADATA_TABLE + "." + BASE_ID
	VIDEO_METADATA_TABLE_CREATED_AT = VIDEO_METADATA_TABLE + "." + BASE_CREATED_AT
	VIDEO_METADATA_TABLE_UPDATED_AT = VIDEO_METADATA_TABLE + "." + BASE_UPDATED_AT
	VIDEO_METADATA_TABLE_ASSET_ID   = VIDEO_METADATA_TABLE + "." + VIDEO_METADATA_ASSET_ID
	VIDEO_METADATA_TABLE_DURATION   = VIDEO_METADATA_TABLE + "." + VIDEO_METADATA_DURATION
	VIDEO_METADATA_TABLE_WIDTH      = VIDEO_METADATA_TABLE + "." + VIDEO_METADATA_WIDTH
	VIDEO_METADATA_TABLE_HEIGHT     = VIDEO_METADATA_TABLE + "." + VIDEO_METADATA_HEIGHT
	VIDEO_METADATA_TABLE_RESOLUTION = VIDEO_METADATA_TABLE + "." + VIDEO_METADATA_RESOLUTION
	VIDEO_METADATA_TABLE_CODEC      = VIDEO_METADATA_TABLE + "." + VIDEO_METADATA_CODEC
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoMetadataJoinColumns returns the columns to join video metadata
func VideoMetadataJoinColumns() []string {
	return []string{
		fmt.Sprintf("%s AS duration", VIDEO_METADATA_TABLE_DURATION),
		fmt.Sprintf("%s AS width", VIDEO_METADATA_TABLE_WIDTH),
		fmt.Sprintf("%s AS height", VIDEO_METADATA_TABLE_HEIGHT),
		fmt.Sprintf("%s AS resolution", VIDEO_METADATA_TABLE_RESOLUTION),
		fmt.Sprintf("%s AS codec", VIDEO_METADATA_TABLE_CODEC),
	}
}
