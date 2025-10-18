package hls

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTranscoder(t *testing.T) {
	// Create a test transcoder
	config := &TranscoderConfig{
		OutputDir:       "/tmp/hls_test",
		HwAccel:         &HwAccelConfig{Type: HwAccelNone, Available: false},
		FFmpegPath:      "ffmpeg",
		FFProbePath:     "ffprobe",
		CleanupInterval: 5 * time.Minute,
		InactiveTimeout: 30 * time.Minute,
	}

	transcoder := NewTranscoder(config)

	t.Run("create file stream", func(t *testing.T) {
		assetID := "test-asset-1"
		filePath := "/test/video.mp4"
		keyframes := []float64{0.0, 2.5, 5.0, 7.5, 10.0}

		stream := transcoder.GetFileStream(assetID, filePath, keyframes)
		require.NotNil(t, stream)
		assert.Equal(t, assetID, stream.ID)
		assert.Equal(t, filePath, stream.FilePath)
		assert.Equal(t, keyframes, stream.Keyframes)
	})

	t.Run("get master playlist", func(t *testing.T) {
		assetID := "test-asset-2"
		filePath := "/test/video.mp4"
		keyframes := []float64{0.0, 2.5, 5.0, 7.5, 10.0}

		// Create file stream
		stream := transcoder.GetFileStream(assetID, filePath, keyframes)

		// Set metadata for the stream
		stream.Duration = 10.0
		stream.Width = 1920
		stream.Height = 1080
		stream.Framerate = 30.0
		stream.Codec = "h264"

		// Get master playlist
		playlist, err := transcoder.GetMasterPlaylist(assetID)
		require.NoError(t, err)
		assert.NotEmpty(t, playlist)
		// Check that the file was created
		assert.FileExists(t, playlist)
	})

	t.Run("get video playlist", func(t *testing.T) {
		assetID := "test-asset-3"
		filePath := "/test/video.mp4"
		keyframes := []float64{0.0, 2.5, 5.0, 7.5, 10.0}

		// Create file stream
		stream := transcoder.GetFileStream(assetID, filePath, keyframes)

		// Set metadata for the stream
		stream.Duration = 10.0
		stream.Width = 1920
		stream.Height = 1080
		stream.Framerate = 30.0
		stream.Codec = "h264"

		// Get video playlist
		playlist, err := transcoder.GetVideoPlaylist(assetID, Original)
		require.NoError(t, err)
		assert.NotEmpty(t, playlist)
		// Check that the file was created
		assert.FileExists(t, playlist)
	})

	t.Run("get segment count", func(t *testing.T) {
		assetID := "test-asset-4"
		filePath := "/test/video.mp4"
		keyframes := []float64{0.0, 2.5, 5.0, 7.5, 10.0}

		// Create file stream
		transcoder.GetFileStream(assetID, filePath, keyframes)

		// Get segment count
		count, err := transcoder.GetSegmentCount(assetID)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("get segment info", func(t *testing.T) {
		assetID := "test-asset-5"
		filePath := "/test/video.mp4"
		keyframes := []float64{0.0, 2.5, 5.0, 7.5, 10.0}

		// Create file stream
		transcoder.GetFileStream(assetID, filePath, keyframes)

		// Get segment info
		startTime, duration, exists, err := transcoder.GetSegmentInfo(assetID, Original, 0)
		require.NoError(t, err)
		assert.Equal(t, 0.0, startTime)
		assert.Equal(t, 2.5, duration)
		assert.False(t, exists) // Segment doesn't exist yet
	})

	t.Run("active streams", func(t *testing.T) {
		// Create a new transcoder for this test to avoid interference
		config := &TranscoderConfig{
			OutputDir:       "/tmp/hls_test_2",
			HwAccel:         &HwAccelConfig{Type: HwAccelNone, Available: false},
			FFmpegPath:      "ffmpeg",
			FFProbePath:     "ffprobe",
			CleanupInterval: 5 * time.Minute,
			InactiveTimeout: 30 * time.Minute,
		}

		testTranscoder := NewTranscoder(config)

		// Initially no active streams
		assert.Equal(t, 0, testTranscoder.GetActiveStreams())

		// Create some streams
		testTranscoder.GetFileStream("asset1", "/test/video1.mp4", []float64{0.0, 2.5})
		testTranscoder.GetFileStream("asset2", "/test/video2.mp4", []float64{0.0, 2.5, 5.0})

		// Should have 2 active streams
		assert.Equal(t, 2, testTranscoder.GetActiveStreams())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestQuality(t *testing.T) {
	t.Run("original quality", func(t *testing.T) {
		assert.True(t, Original.IsOriginal())
		assert.Equal(t, uint32(0), Original.Height())
		assert.Equal(t, uint32(0), Original.AverageBitrate())
		assert.True(t, Original.IsValid())
	})

	t.Run("quality from height", func(t *testing.T) {
		assert.Equal(t, P240, GetQualityFromHeight(240))
		assert.Equal(t, P720, GetQualityFromHeight(720))
		assert.Equal(t, P1080, GetQualityFromHeight(1080))
		assert.Equal(t, P4k, GetQualityFromHeight(2160))
		assert.Equal(t, Original, GetQualityFromHeight(100))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestHwAccel(t *testing.T) {
	t.Run("no hardware acceleration", func(t *testing.T) {
		config := &HwAccelConfig{
			Type:      HwAccelNone,
			Available: false,
			EncodeFlags: []string{
				"-c:v", "libx264",
				"-preset", "fast",
				"-sc_threshold", "0",
				"-pix_fmt", "yuv420p",
			},
		}

		assert.False(t, config.IsHardwareAccelerated())
		assert.Equal(t, "none", config.String())

		// Test encode flags
		assert.Contains(t, config.EncodeFlags, "-c:v")
		assert.Contains(t, config.EncodeFlags, "libx264")
	})

	t.Run("hardware acceleration", func(t *testing.T) {
		config := &HwAccelConfig{
			Type:           HwAccelVAAPI,
			Available:      true,
			DecodeFlags:    []string{"-hwaccel", "vaapi", "-hwaccel_device", "/dev/dri/renderD128"},
			EncodeFlags:    []string{"-c:v", "h264_vaapi"},
			ScaleFilter:    "format=nv12|vaapi,hwupload,scale_vaapi=%d:%d:format=nv12",
			NoResizeFilter: "format=nv12|vaapi,hwupload,scale_vaapi=format=nv12",
			Preset:         "fast",
		}

		assert.True(t, config.IsHardwareAccelerated())
		assert.Equal(t, "vaapi", config.String())

		// Test decode and encode flags
		assert.Contains(t, config.DecodeFlags, "-hwaccel")
		assert.Contains(t, config.DecodeFlags, "vaapi")
		assert.Contains(t, config.EncodeFlags, "-c:v")
		assert.Contains(t, config.EncodeFlags, "h264_vaapi")

		// Test scaling filters
		scaleFilter := config.GetScaleFilter(1920, 1080)
		assert.Contains(t, scaleFilter, "scale_vaapi=1920:1080")

		noResizeFilter := config.GetNoResizeFilter()
		assert.Contains(t, noResizeFilter, "scale_vaapi=format=nv12")
	})
}
