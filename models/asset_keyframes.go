package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	// Table
	ASSET_KEYFRAMES_TABLE = "asset_keyframes"

	// Columns
	KEYFRAMES_ASSET_ID    = "asset_id"
	KEYFRAMES_DATA        = "keyframes"
	KEYFRAMES_IS_COMPLETE = "is_complete"

	// Qualified columns
	KEYFRAMES_TABLE_ID          = ASSET_KEYFRAMES_TABLE + "." + BASE_ID
	KEYFRAMES_TABLE_ASSET_ID    = ASSET_KEYFRAMES_TABLE + "." + KEYFRAMES_ASSET_ID
	KEYFRAMES_TABLE_DATA        = ASSET_KEYFRAMES_TABLE + "." + KEYFRAMES_DATA
	KEYFRAMES_TABLE_IS_COMPLETE = ASSET_KEYFRAMES_TABLE + "." + KEYFRAMES_IS_COMPLETE
	KEYFRAMES_TABLE_CREATED_AT  = ASSET_KEYFRAMES_TABLE + "." + BASE_CREATED_AT
	KEYFRAMES_TABLE_UPDATED_AT  = ASSET_KEYFRAMES_TABLE + "." + BASE_UPDATED_AT
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetKeyframes defines keyframe data for an asset
type AssetKeyframes struct {
	Base
	AssetID       string `db:"asset_id"`    // Immutable
	KeyframesJSON string `db:"keyframes"`   // Raw JSON in DB
	IsComplete    bool   `db:"is_complete"` // Mutable

	Keyframes []float64 `db:"-"` // Populated from JSON
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MarshalKeyframes converts the Keyframes slice to JSON and stores it in KeyframesJSON
func (ak *AssetKeyframes) MarshalKeyframes() error {
	if ak.Keyframes == nil {
		ak.KeyframesJSON = "[]"
		return nil
	}

	data, err := json.Marshal(ak.Keyframes)
	if err != nil {
		return fmt.Errorf("failed to marshal keyframes: %w", err)
	}

	ak.KeyframesJSON = string(data)
	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UnmarshalKeyframes parses the KeyframesJSON and populates the Keyframes slice
func (ak *AssetKeyframes) UnmarshalKeyframes() error {
	if ak.KeyframesJSON == "" {
		ak.Keyframes = []float64{}
		return nil
	}

	err := json.Unmarshal([]byte(ak.KeyframesJSON), &ak.Keyframes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal keyframes: %w", err)
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetKeyframesColumns returns the list of columns to use when populating `AssetKeyframes`
func AssetKeyframesColumns() []string {
	return []string{
		fmt.Sprintf("%s AS id", KEYFRAMES_TABLE_ID),
		fmt.Sprintf("%s AS created_at", KEYFRAMES_TABLE_CREATED_AT),
		fmt.Sprintf("%s AS updated_at", KEYFRAMES_TABLE_UPDATED_AT),
		fmt.Sprintf("%s AS asset_id", KEYFRAMES_TABLE_ASSET_ID),
		fmt.Sprintf("%s AS keyframes", KEYFRAMES_TABLE_DATA),
		fmt.Sprintf("%s AS is_complete", KEYFRAMES_TABLE_IS_COMPLETE),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetKeyframesRow is for use in scanning joined asset keyframes rows
type AssetKeyframesRow struct {
	AssetID       string `db:"asset_id"`
	KeyframesJSON string `db:"keyframes"`
	IsComplete    bool   `db:"is_complete"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ToDomain converts AssetKeyframesRow to AssetKeyframes
func (r *AssetKeyframesRow) ToDomain() *AssetKeyframes {
	ak := &AssetKeyframes{
		AssetID:       r.AssetID,
		KeyframesJSON: r.KeyframesJSON,
		IsComplete:    r.IsComplete,
	}

	// Parse the JSON keyframes
	if err := ak.UnmarshalKeyframes(); err != nil {
		// If parsing fails, set empty slice
		ak.Keyframes = []float64{}
	}

	return ak
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetKeyframesRowColumns returns the list of columns to use when populating `AssetKeyframesRow`
func AssetKeyframesRowColumns() []string {
	return []string{
		fmt.Sprintf("%s AS asset_id", KEYFRAMES_TABLE_ASSET_ID),
		fmt.Sprintf("%s AS keyframes", KEYFRAMES_TABLE_DATA),
		fmt.Sprintf("%s AS is_complete", KEYFRAMES_TABLE_IS_COMPLETE),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ValidateKeyframes validates that the keyframes are in ascending order and non-negative
func (ak *AssetKeyframes) ValidateKeyframes() error {
	if len(ak.Keyframes) == 0 {
		return nil
	}

	// Check for negative timestamps
	for i, timestamp := range ak.Keyframes {
		if timestamp < 0 {
			return fmt.Errorf("keyframe at index %d has negative timestamp: %f", i, timestamp)
		}
	}

	// Check for ascending order
	for i := 1; i < len(ak.Keyframes); i++ {
		if ak.Keyframes[i] <= ak.Keyframes[i-1] {
			return fmt.Errorf("keyframes not in ascending order: %f <= %f at indices %d, %d",
				ak.Keyframes[i], ak.Keyframes[i-1], i, i-1)
		}
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentCount returns the number of segments that would be generated from these keyframes
func (ak *AssetKeyframes) GetSegmentCount() int {
	if len(ak.Keyframes) == 0 {
		return 0
	}

	// Number of segments = number of keyframes (each keyframe starts a new segment)
	return len(ak.Keyframes)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentDuration returns the duration of a specific segment
func (ak *AssetKeyframes) GetSegmentDuration(segmentIndex int) float64 {
	if segmentIndex < 0 || segmentIndex >= len(ak.Keyframes) {
		return 0
	}

	// If this is the last segment, we can't determine duration without total video duration
	// For now, return 0 for the last segment
	if segmentIndex == len(ak.Keyframes)-1 {
		return 0
	}

	// Duration is the difference between this keyframe and the next
	return ak.Keyframes[segmentIndex+1] - ak.Keyframes[segmentIndex]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String returns a string representation of the keyframes for debugging
func (ak *AssetKeyframes) String() string {
	if len(ak.Keyframes) == 0 {
		return "AssetKeyframes{asset_id=" + ak.AssetID + ", keyframes=[], is_complete=" + fmt.Sprintf("%t", ak.IsComplete) + "}"
	}

	// Show first few and last few keyframes
	var parts []string
	if len(ak.Keyframes) <= 6 {
		// Show all if 6 or fewer
		for _, kf := range ak.Keyframes {
			parts = append(parts, fmt.Sprintf("%.2f", kf))
		}
	} else {
		// Show first 3, ..., last 3
		for i := 0; i < 3; i++ {
			parts = append(parts, fmt.Sprintf("%.2f", ak.Keyframes[i]))
		}
		parts = append(parts, "...")
		for i := len(ak.Keyframes) - 3; i < len(ak.Keyframes); i++ {
			parts = append(parts, fmt.Sprintf("%.2f", ak.Keyframes[i]))
		}
	}

	return fmt.Sprintf("AssetKeyframes{asset_id=%s, keyframes=[%s], is_complete=%t}",
		ak.AssetID, strings.Join(parts, ", "), ak.IsComplete)
}
