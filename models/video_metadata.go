package models

import (
	"github.com/geerew/off-course/utils/schema"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoMetadata defines video-related metadata for an asset
type VideoMetadata struct {
	Base
	AssetID    string
	Duration   int
	Width      int
	Height     int
	Resolution string
	Codec      string
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

// Table implements the schema.Modeler interface
func (m *VideoMetadata) Table() string {
	return VIDEO_METADATA_TABLE
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Fields defines the fields for the schema
func (m *VideoMetadata) Define(s *schema.ModelConfig) {
	s.Embedded("Base")

	s.Field("AssetID").Column(VIDEO_METADATA_ASSET_ID).NotNull()
	s.Field("Duration").Column(VIDEO_METADATA_DURATION).NotNull().Mutable()
	s.Field("Width").Column(VIDEO_METADATA_WIDTH).NotNull().Mutable()
	s.Field("Height").Column(VIDEO_METADATA_HEIGHT).NotNull().Mutable()
	s.Field("Resolution").Column(VIDEO_METADATA_RESOLUTION).NotNull().Mutable()
	s.Field("Codec").Column(VIDEO_METADATA_CODEC).NotNull().Mutable()
}
