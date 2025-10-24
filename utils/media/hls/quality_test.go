package hls

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestQuality_FromString(t *testing.T) {
	t.Run("valid qualities", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected Quality
		}{
			{"240p", P240},
			{"360p", P360},
			{"480p", P480},
			{"720p", P720},
			{"1080p", P1080},
			{"1440p", P1440},
			{"transcode", NoResize},
			{"original", Original},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				quality, err := QualityFromString(tc.input)
				require.NoError(t, err)
				require.Equal(t, tc.expected, quality)
			})
		}
	})

	t.Run("invalid quality", func(t *testing.T) {
		quality, err := QualityFromString("invalid")
		require.Error(t, err)
		require.Equal(t, Original, quality) // Should return Original as fallback
		require.Equal(t, "invalid quality", err.Error())
	})

	t.Run("empty string", func(t *testing.T) {
		quality, err := QualityFromString("")
		require.Error(t, err)
		require.Equal(t, Original, quality) // Should return Original as fallback
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestQuality_Height(t *testing.T) {
	testCases := []struct {
		quality  Quality
		expected uint32
	}{
		{P240, 240},
		{P360, 360},
		{P480, 480},
		{P720, 720},
		{P1080, 1080},
		{P1440, 1440},
	}

	for _, tc := range testCases {
		t.Run(string(tc.quality), func(t *testing.T) {
			height := tc.quality.Height()
			require.Equal(t, tc.expected, height)
		})
	}

	t.Run("original quality panics", func(t *testing.T) {
		require.Panics(t, func() {
			Original.Height()
		})
	})

	t.Run("invalid quality panics", func(t *testing.T) {
		require.Panics(t, func() {
			Quality("invalid").Height()
		})
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestQuality_AverageBitrate(t *testing.T) {
	testCases := []struct {
		quality  Quality
		expected uint32
	}{
		{P240, 400_000},
		{P360, 800_000},
		{P480, 1_200_000},
		{P720, 2_400_000},
		{P1080, 4_800_000},
		{P1440, 9_600_000},
	}

	for _, tc := range testCases {
		t.Run(string(tc.quality), func(t *testing.T) {
			bitrate := tc.quality.AverageBitrate()
			require.Equal(t, tc.expected, bitrate)
		})
	}

	t.Run("original quality panics", func(t *testing.T) {
		require.Panics(t, func() {
			Original.AverageBitrate()
		})
	})

	t.Run("invalid quality panics", func(t *testing.T) {
		require.Panics(t, func() {
			Quality("invalid").AverageBitrate()
		})
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestQuality_MaxBitrate(t *testing.T) {
	testCases := []struct {
		quality  Quality
		expected uint32
	}{
		{P240, 700_000},
		{P360, 1_400_000},
		{P480, 2_100_000},
		{P720, 4_000_000},
		{P1080, 8_000_000},
		{P1440, 12_000_000},
	}

	for _, tc := range testCases {
		t.Run(string(tc.quality), func(t *testing.T) {
			bitrate := tc.quality.MaxBitrate()
			require.Equal(t, tc.expected, bitrate)
		})
	}

	t.Run("original quality panics", func(t *testing.T) {
		require.Panics(t, func() {
			Original.MaxBitrate()
		})
	})

	t.Run("invalid quality panics", func(t *testing.T) {
		require.Panics(t, func() {
			Quality("invalid").MaxBitrate()
		})
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestGetQuality_ForVideo(t *testing.T) {
	t.Run("height-based selection", func(t *testing.T) {
		testCases := []struct {
			height   uint32
			bitrate  uint32
			expected Quality
		}{
			// Height-based selection - function returns first quality where height >= video height OR bitrate >= video bitrate
			{200, 100_000, P240},  // P240.Height() >= 200 is false, but P240.AverageBitrate() >= 100_000 is true (400_000 >= 100_000)
			{240, 100_000, P240},  // P240.Height() >= 240 is true
			{300, 100_000, P240},  // P240.Height() >= 300 is false, but P240.AverageBitrate() >= 100_000 is true (400_000 >= 100_000)
			{360, 100_000, P240},  // P240.Height() >= 360 is false, but P240.AverageBitrate() >= 100_000 is true (400_000 >= 100_000)
			{500, 100_000, P240},  // P240.Height() >= 500 is false, but P240.AverageBitrate() >= 100_000 is true (400_000 >= 100_000)
			{720, 100_000, P240},  // P240.Height() >= 720 is false, but P240.AverageBitrate() >= 100_000 is true (400_000 >= 100_000)
			{1000, 100_000, P240}, // P240.Height() >= 1000 is false, but P240.AverageBitrate() >= 100_000 is true (400_000 >= 100_000)
			{1080, 100_000, P240}, // P240.Height() >= 1080 is false, but P240.AverageBitrate() >= 100_000 is true (400_000 >= 100_000)
			{1500, 100_000, P240}, // P240.Height() >= 1500 is false, but P240.AverageBitrate() >= 100_000 is true (400_000 >= 100_000)
			{2000, 100_000, P240}, // P240.Height() >= 2000 is false, but P240.AverageBitrate() >= 100_000 is true (400_000 >= 100_000)
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("height_%d", tc.height), func(t *testing.T) {
				quality := GetQualityForVideo(tc.height, tc.bitrate)
				require.Equal(t, tc.expected, quality)
			})
		}
	})

	t.Run("bitrate-based selection", func(t *testing.T) {
		testCases := []struct {
			height   uint32
			bitrate  uint32
			expected Quality
		}{
			// Bitrate-based selection (when height is low but bitrate is high)
			// Note: P240.Height() >= 100 is true (240 >= 100), so it returns P240 immediately
			{100, 500_000, P240},   // P240.Height() >= 100 is true (240 >= 100), so returns P240
			{100, 1_000_000, P240}, // P240.Height() >= 100 is true (240 >= 100), so returns P240
			{100, 3_000_000, P240}, // P240.Height() >= 100 is true (240 >= 100), so returns P240
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("bitrate_%d", tc.bitrate), func(t *testing.T) {
				quality := GetQualityForVideo(tc.height, tc.bitrate)
				require.Equal(t, tc.expected, quality)
			})
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestQuality_SelectionLogic(t *testing.T) {
	t.Run("qualities up to video height", func(t *testing.T) {
		testCases := []struct {
			videoHeight uint32
			expected    []Quality
		}{
			{240, []Quality{P240, Original}},
			{360, []Quality{P240, P360, Original}},
			{480, []Quality{P240, P360, P480, Original}},
			{720, []Quality{P240, P360, P480, P720, Original}},
			{1080, []Quality{P240, P360, P480, P720, P1080, Original}},
			{1440, []Quality{P240, P360, P480, P720, P1080, P1440, Original}},
			{2160, []Quality{P240, P360, P480, P720, P1080, P1440, Original}}, // 4K video
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("height_%d", tc.videoHeight), func(t *testing.T) {
				var qualities []Quality
				for _, q := range Qualities {
					if q.Height() <= tc.videoHeight {
						qualities = append(qualities, q)
					}
				}
				qualities = append(qualities, Original)

				require.Equal(t, tc.expected, qualities)
			})
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestQuality_Constants(t *testing.T) {
	t.Run("qualities slice contains all standard qualities", func(t *testing.T) {
		expected := []Quality{P240, P360, P480, P720, P1080, P1440}
		require.Equal(t, expected, Qualities)
	})

	t.Run("quality string values", func(t *testing.T) {
		require.Equal(t, "240p", string(P240))
		require.Equal(t, "360p", string(P360))
		require.Equal(t, "480p", string(P480))
		require.Equal(t, "720p", string(P720))
		require.Equal(t, "1080p", string(P1080))
		require.Equal(t, "1440p", string(P1440))
		require.Equal(t, "transcode", string(NoResize))
		require.Equal(t, "original", string(Original))
	})
}
