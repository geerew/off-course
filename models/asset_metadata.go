package models

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	// Tables
	MEDIA_VIDEO_TABLE = "asset_media_video"
	MEDIA_AUDIO_TABLE = "asset_media_audio"

	// Shared columns
	META_ASSET_ID = "asset_id"

	// Video table columns
	MEDIA_VIDEO_DURATION    = "duration_sec"
	MEDIA_VIDEO_CONTAINER   = "container"
	MEDIA_VIDEO_MIME_TYPE   = "mime_type"
	MEDIA_VIDEO_SIZE_BYTES  = "size_bytes"
	MEDIA_VIDEO_OVERALL_BPS = "overall_bps"
	MEDIA_VIDEO_CODEC       = "video_codec"
	MEDIA_VIDEO_WIDTH       = "width"
	MEDIA_VIDEO_HEIGHT      = "height"
	MEDIA_VIDEO_FPS_NUM     = "fps_num"
	MEDIA_VIDEO_FPS_DEN     = "fps_den"

	// Audio table columns
	MEDIA_AUDIO_LANGUAGE       = "language"
	MEDIA_AUDIO_CODEC          = "codec"
	MEDIA_AUDIO_PROFILE        = "profile"
	MEDIA_AUDIO_CHANNELS       = "channels"
	MEDIA_AUDIO_CHANNEL_LAYOUT = "channel_layout"
	MEDIA_AUDIO_SAMPLE_RATE    = "sample_rate"
	MEDIA_AUDIO_BIT_RATE       = "bit_rate"

	// Qualified video columns
	MEDIA_VIDEO_TABLE_ID          = MEDIA_VIDEO_TABLE + "." + BASE_ID
	MEDIA_VIDEO_TABLE_ASSET_ID    = MEDIA_VIDEO_TABLE + "." + META_ASSET_ID
	MEDIA_VIDEO_TABLE_DURATION    = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_DURATION
	MEDIA_VIDEO_TABLE_CONTAINER   = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_CONTAINER
	MEDIA_VIDEO_TABLE_MIME_TYPE   = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_MIME_TYPE
	MEDIA_VIDEO_TABLE_SIZE_BYTES  = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_SIZE_BYTES
	MEDIA_VIDEO_TABLE_OVERALL_BPS = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_OVERALL_BPS
	MEDIA_VIDEO_TABLE_CODEC       = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_CODEC
	MEDIA_VIDEO_TABLE_WIDTH       = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_WIDTH
	MEDIA_VIDEO_TABLE_HEIGHT      = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_HEIGHT
	MEDIA_VIDEO_TABLE_FPS_NUM     = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_FPS_NUM
	MEDIA_VIDEO_TABLE_FPS_DEN     = MEDIA_VIDEO_TABLE + "." + MEDIA_VIDEO_FPS_DEN
	MEDIA_VIDEO_TABLE_CREATED_AT  = MEDIA_VIDEO_TABLE + "." + BASE_CREATED_AT
	MEDIA_VIDEO_TABLE_UPDATED_AT  = MEDIA_VIDEO_TABLE + "." + BASE_UPDATED_AT

	// Qualified audio columns
	MEDIA_AUDIO_TABLE_ID             = MEDIA_AUDIO_TABLE + "." + BASE_ID
	MEDIA_AUDIO_TABLE_ASSET_ID       = MEDIA_AUDIO_TABLE + "." + META_ASSET_ID
	MEDIA_AUDIO_TABLE_LANGUAGE       = MEDIA_AUDIO_TABLE + "." + MEDIA_AUDIO_LANGUAGE
	MEDIA_AUDIO_TABLE_CODEC          = MEDIA_AUDIO_TABLE + "." + MEDIA_AUDIO_CODEC
	MEDIA_AUDIO_TABLE_PROFILE        = MEDIA_AUDIO_TABLE + "." + MEDIA_AUDIO_PROFILE
	MEDIA_AUDIO_TABLE_CHANNELS       = MEDIA_AUDIO_TABLE + "." + MEDIA_AUDIO_CHANNELS
	MEDIA_AUDIO_TABLE_CHANNEL_LAYOUT = MEDIA_AUDIO_TABLE + "." + MEDIA_AUDIO_CHANNEL_LAYOUT
	MEDIA_AUDIO_TABLE_SAMPLE_RATE    = MEDIA_AUDIO_TABLE + "." + MEDIA_AUDIO_SAMPLE_RATE
	MEDIA_AUDIO_TABLE_BIT_RATE       = MEDIA_AUDIO_TABLE + "." + MEDIA_AUDIO_BIT_RATE
	MEDIA_AUDIO_TABLE_CREATED_AT     = MEDIA_AUDIO_TABLE + "." + BASE_CREATED_AT
	MEDIA_AUDIO_TABLE_UPDATED_AT     = MEDIA_AUDIO_TABLE + "." + BASE_UPDATED_AT
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetMetadata defines metadata for an asset
type AssetMetadata struct {
	AssetID string `db:"asset_id"` // Immutable

	// Joins
	VideoMetadata *VideoMetadata
	AudioMetadata *AudioMetadata
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoMetadata defines video metadata for an asset
type VideoMetadata struct {
	Base
	DurationSec int

	// Container
	Container  string
	MIMEType   string
	SizeBytes  int64
	OverallBPS int

	// Video
	VideoCodec string
	Width      int
	Height     int
	FPSNum     int
	FPSDen     int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoMetaJoinedRow is for use in scanning joined video metadata rows
type VideoMetaJoinedRow struct {
	VideoID      sql.NullString `db:"video_id"`
	DurationSec  sql.NullInt64  `db:"duration_sec"`
	Container    sql.NullString `db:"container"`
	MIMEType     sql.NullString `db:"mime_type"`
	SizeBytes    sql.NullInt64  `db:"size_bytes"`
	OverallBPS   sql.NullInt64  `db:"overall_bps"`
	VideoCodec   sql.NullString `db:"video_codec"`
	Width        sql.NullInt64  `db:"width"`
	Height       sql.NullInt64  `db:"height"`
	FPSNum       sql.NullInt64  `db:"fps_num"`
	FPSDen       sql.NullInt64  `db:"fps_den"`
	VideoCreated types.DateTime `db:"video_created_at"`
	VideoUpdated types.DateTime `db:"video_updated_at"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AudioMetadata defines audio metadata for an asset
type AudioMetadata struct {
	Base
	Language      string
	Codec         string
	Profile       string
	Channels      int
	ChannelLayout string
	SampleRate    int
	BitRate       int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AudioMetaJoinedRow is for use in scanning joined audio metadata rows
type AudioMetaJoinedRow struct {
	AudioID            sql.NullString `db:"audio_id"`
	AudioLanguage      sql.NullString `db:"audio_language"`
	AudioCodec         sql.NullString `db:"audio_codec"`
	AudioProfile       sql.NullString `db:"audio_profile"`
	AudioChannels      sql.NullInt64  `db:"audio_channels"`
	AudioChannelLayout sql.NullString `db:"audio_channel_layout"`
	AudioSampleRate    sql.NullInt64  `db:"audio_sample_rate"`
	AudioBitRate       sql.NullInt64  `db:"audio_bit_rate"`
	AudioCreated       types.DateTime `db:"audio_created_at"`
	AudioUpdated       types.DateTime `db:"audio_updated_at"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetMetadataRow is for use in scanning joined asset metadata rows
type AssetMetadataRow struct {
	AssetID string `db:"asset_id"`
	VideoMetaJoinedRow
	AudioMetaJoinedRow
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ToDomain converts AssetMetadataRow to AssetMetadata
func (r *AssetMetadataRow) ToDomain() *AssetMetadata {
	out := &AssetMetadata{
		AssetID: r.AssetID,
	}

	if r.VideoID.Valid {
		out.VideoMetadata = &VideoMetadata{
			Base: Base{
				ID:        r.VideoID.String,
				CreatedAt: r.VideoCreated,
				UpdatedAt: r.VideoUpdated,
			},
			DurationSec: int(r.DurationSec.Int64),
			Container:   r.Container.String,
			MIMEType:    r.MIMEType.String,
			SizeBytes:   r.SizeBytes.Int64,
			OverallBPS:  int(r.OverallBPS.Int64),
			VideoCodec:  strings.ToLower(r.VideoCodec.String),
			Width:       int(r.Width.Int64),
			Height:      int(r.Height.Int64),
			FPSNum:      int(r.FPSNum.Int64),
			FPSDen:      int(r.FPSDen.Int64),
		}
	}

	if r.AudioID.Valid {
		ch := int(r.AudioChannels.Int64)
		if ch == 0 {
			switch strings.ToLower(r.AudioChannelLayout.String) {
			case "mono":
				ch = 1
			case "stereo", "2.0":
				ch = 2
			case "5.1":
				ch = 6
			case "7.1":
				ch = 8
			}
		}
		out.AudioMetadata = &AudioMetadata{
			Base: Base{
				ID:        r.AudioID.String,
				CreatedAt: r.AudioCreated,
				UpdatedAt: r.AudioUpdated,
			},
			Language:      strings.ToLower(r.AudioLanguage.String),
			Codec:         strings.ToLower(r.AudioCodec.String),
			Profile:       r.AudioProfile.String,
			Channels:      ch,
			ChannelLayout: r.AudioChannelLayout.String,
			SampleRate:    int(r.AudioSampleRate.Int64),
			BitRate:       int(r.AudioBitRate.Int64),
		}
	}

	return out
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetMetadataRowColumns returns the list of columns to use when populating `AssetMetadataRow`
func AssetMetadataRowColumns() []string {
	return []string{
		fmt.Sprintf("%s AS asset_id", ASSET_TABLE_ID),

		// Video
		fmt.Sprintf("%s AS video_id", MEDIA_VIDEO_TABLE_ID),
		fmt.Sprintf("%s AS duration_sec", MEDIA_VIDEO_TABLE_DURATION),
		fmt.Sprintf("%s AS container", MEDIA_VIDEO_TABLE_CONTAINER),
		fmt.Sprintf("%s AS mime_type", MEDIA_VIDEO_TABLE_MIME_TYPE),
		fmt.Sprintf("%s AS size_bytes", MEDIA_VIDEO_TABLE_SIZE_BYTES),
		fmt.Sprintf("%s AS overall_bps", MEDIA_VIDEO_TABLE_OVERALL_BPS),
		fmt.Sprintf("%s AS video_codec", MEDIA_VIDEO_TABLE_CODEC),
		fmt.Sprintf("%s AS width", MEDIA_VIDEO_TABLE_WIDTH),
		fmt.Sprintf("%s AS height", MEDIA_VIDEO_TABLE_HEIGHT),
		fmt.Sprintf("%s AS fps_num", MEDIA_VIDEO_TABLE_FPS_NUM),
		fmt.Sprintf("%s AS fps_den", MEDIA_VIDEO_TABLE_FPS_DEN),
		fmt.Sprintf("%s AS video_created_at", MEDIA_VIDEO_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS video_updated_at", MEDIA_VIDEO_TABLE_UPDATED_AT),

		// Audio
		fmt.Sprintf("%s AS audio_id", MEDIA_AUDIO_TABLE_ID),
		fmt.Sprintf("%s AS audio_language", MEDIA_AUDIO_TABLE_LANGUAGE),
		fmt.Sprintf("%s AS audio_codec", MEDIA_AUDIO_TABLE_CODEC),
		fmt.Sprintf("%s AS audio_profile", MEDIA_AUDIO_TABLE_PROFILE),
		fmt.Sprintf("%s AS audio_channels", MEDIA_AUDIO_TABLE_CHANNELS),
		fmt.Sprintf("%s AS audio_channel_layout", MEDIA_AUDIO_TABLE_CHANNEL_LAYOUT),
		fmt.Sprintf("%s AS audio_sample_rate", MEDIA_AUDIO_TABLE_SAMPLE_RATE),
		fmt.Sprintf("%s AS audio_bit_rate", MEDIA_AUDIO_BIT_RATE),
		fmt.Sprintf("%s AS audio_created_at", MEDIA_AUDIO_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS audio_updated_at", MEDIA_AUDIO_TABLE_UPDATED_AT),
	}
}
