package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetKeyframes_MarshalKeyframes(t *testing.T) {
	t.Run("empty keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{},
		}

		err := ak.MarshalKeyframes()
		require.NoError(t, err)
		assert.Equal(t, "[]", ak.KeyframesJSON)
	})

	t.Run("nil keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: nil,
		}

		err := ak.MarshalKeyframes()
		require.NoError(t, err)
		assert.Equal(t, "[]", ak.KeyframesJSON)
	})

	t.Run("valid keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{0.0, 2.5, 5.0, 7.5, 10.0},
		}

		err := ak.MarshalKeyframes()
		require.NoError(t, err)
		assert.Equal(t, "[0,2.5,5,7.5,10]", ak.KeyframesJSON)
	})

	t.Run("single keyframe", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{0.0},
		}

		err := ak.MarshalKeyframes()
		require.NoError(t, err)
		assert.Equal(t, "[0]", ak.KeyframesJSON)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetKeyframes_UnmarshalKeyframes(t *testing.T) {
	t.Run("empty JSON", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:       "test-asset",
			KeyframesJSON: "",
		}

		err := ak.UnmarshalKeyframes()
		require.NoError(t, err)
		assert.Equal(t, []float64{}, ak.Keyframes)
	})

	t.Run("empty array JSON", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:       "test-asset",
			KeyframesJSON: "[]",
		}

		err := ak.UnmarshalKeyframes()
		require.NoError(t, err)
		assert.Equal(t, []float64{}, ak.Keyframes)
	})

	t.Run("valid JSON", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:       "test-asset",
			KeyframesJSON: "[0,2.5,5,7.5,10]",
		}

		err := ak.UnmarshalKeyframes()
		require.NoError(t, err)
		assert.Equal(t, []float64{0.0, 2.5, 5.0, 7.5, 10.0}, ak.Keyframes)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:       "test-asset",
			KeyframesJSON: "invalid json",
		}

		err := ak.UnmarshalKeyframes()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal keyframes")
	})

	t.Run("wrong type JSON", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:       "test-asset",
			KeyframesJSON: `{"not": "array"}`,
		}

		err := ak.UnmarshalKeyframes()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal keyframes")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetKeyframes_ValidateKeyframes(t *testing.T) {
	t.Run("empty keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{},
		}

		err := ak.ValidateKeyframes()
		require.NoError(t, err)
	})

	t.Run("single keyframe", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{0.0},
		}

		err := ak.ValidateKeyframes()
		require.NoError(t, err)
	})

	t.Run("valid ascending keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{0.0, 2.5, 5.0, 7.5, 10.0},
		}

		err := ak.ValidateKeyframes()
		require.NoError(t, err)
	})

	t.Run("negative timestamp", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{-1.0, 2.5, 5.0},
		}

		err := ak.ValidateKeyframes()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "negative timestamp")
	})

	t.Run("not ascending order", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{0.0, 5.0, 2.5, 10.0},
		}

		err := ak.ValidateKeyframes()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not in ascending order")
	})

	t.Run("duplicate timestamps", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{0.0, 2.5, 2.5, 5.0},
		}

		err := ak.ValidateKeyframes()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not in ascending order")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetKeyframes_GetSegmentCount(t *testing.T) {
	t.Run("empty keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{},
		}

		count := ak.GetSegmentCount()
		assert.Equal(t, 0, count)
	})

	t.Run("single keyframe", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{0.0},
		}

		count := ak.GetSegmentCount()
		assert.Equal(t, 1, count)
	})

	t.Run("multiple keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:   "test-asset",
			Keyframes: []float64{0.0, 2.5, 5.0, 7.5, 10.0},
		}

		count := ak.GetSegmentCount()
		assert.Equal(t, 5, count)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetKeyframes_GetSegmentDuration(t *testing.T) {
	ak := &AssetKeyframes{
		AssetID:   "test-asset",
		Keyframes: []float64{0.0, 2.5, 5.0, 7.5, 10.0},
	}

	t.Run("negative index", func(t *testing.T) {
		duration := ak.GetSegmentDuration(-1)
		assert.Equal(t, 0.0, duration)
	})

	t.Run("index out of bounds", func(t *testing.T) {
		duration := ak.GetSegmentDuration(10)
		assert.Equal(t, 0.0, duration)
	})

	t.Run("valid segments", func(t *testing.T) {
		duration0 := ak.GetSegmentDuration(0)
		assert.Equal(t, 2.5, duration0)

		duration1 := ak.GetSegmentDuration(1)
		assert.Equal(t, 2.5, duration1)

		duration2 := ak.GetSegmentDuration(2)
		assert.Equal(t, 2.5, duration2)

		duration3 := ak.GetSegmentDuration(3)
		assert.Equal(t, 2.5, duration3)
	})

	t.Run("last segment", func(t *testing.T) {
		// Last segment duration can't be determined without total video duration
		duration := ak.GetSegmentDuration(4)
		assert.Equal(t, 0.0, duration)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetKeyframes_String(t *testing.T) {
	t.Run("empty keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:    "test-asset",
			Keyframes:  []float64{},
			IsComplete: false,
		}

		result := ak.String()
		assert.Contains(t, result, "asset_id=test-asset")
		assert.Contains(t, result, "keyframes=[]")
		assert.Contains(t, result, "is_complete=false")
	})

	t.Run("few keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:    "test-asset",
			Keyframes:  []float64{0.0, 2.5, 5.0},
			IsComplete: true,
		}

		result := ak.String()
		assert.Contains(t, result, "asset_id=test-asset")
		assert.Contains(t, result, "keyframes=[0.00, 2.50, 5.00]")
		assert.Contains(t, result, "is_complete=true")
	})

	t.Run("many keyframes", func(t *testing.T) {
		ak := &AssetKeyframes{
			AssetID:    "test-asset",
			Keyframes:  []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0},
			IsComplete: true,
		}

		result := ak.String()
		assert.Contains(t, result, "asset_id=test-asset")
		assert.Contains(t, result, "keyframes=[0.00, 1.00, 2.00, ..., 7.00, 8.00, 9.00]")
		assert.Contains(t, result, "is_complete=true")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetKeyframesColumns(t *testing.T) {
	columns := AssetKeyframesColumns()

	expectedColumns := []string{
		"asset_keyframes.id AS id",
		"asset_keyframes.created_at AS created_at",
		"asset_keyframes.updated_at AS updated_at",
		"asset_keyframes.asset_id AS asset_id",
		"asset_keyframes.keyframes AS keyframes",
		"asset_keyframes.is_complete AS is_complete",
	}

	assert.Equal(t, expectedColumns, columns)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetKeyframesRow_ToDomain(t *testing.T) {
	t.Run("valid row", func(t *testing.T) {
		row := &AssetKeyframesRow{
			AssetID:       "test-asset",
			KeyframesJSON: "[0,2.5,5,7.5,10]",
			IsComplete:    true,
		}

		ak := row.ToDomain()
		require.NotNil(t, ak)
		assert.Equal(t, "test-asset", ak.AssetID)
		assert.Equal(t, []float64{0.0, 2.5, 5.0, 7.5, 10.0}, ak.Keyframes)
		assert.True(t, ak.IsComplete)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		row := &AssetKeyframesRow{
			AssetID:       "test-asset",
			KeyframesJSON: "invalid json",
			IsComplete:    false,
		}

		ak := row.ToDomain()
		require.NotNil(t, ak)
		assert.Equal(t, "test-asset", ak.AssetID)
		assert.Equal(t, []float64{}, ak.Keyframes) // Should default to empty slice
		assert.False(t, ak.IsComplete)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssetKeyframesRowColumns(t *testing.T) {
	columns := AssetKeyframesRowColumns()

	expectedColumns := []string{
		"asset_keyframes.asset_id AS asset_id",
		"asset_keyframes.keyframes AS keyframes",
		"asset_keyframes.is_complete AS is_complete",
	}

	assert.Equal(t, expectedColumns, columns)
}
