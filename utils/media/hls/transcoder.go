package hls

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Transcoder manages all active file streams for HLS transcoding
type Transcoder struct {
	OutputDir   string
	HwAccel     *HwAccelConfig
	FFmpegPath  string
	FFProbePath string

	// File streams by asset ID
	fileStreams CMap[string, *FileStream]

	// Cleanup settings
	CleanupInterval time.Duration
	InactiveTimeout time.Duration
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TranscoderConfig holds configuration for the transcoder
type TranscoderConfig struct {
	OutputDir       string
	HwAccel         *HwAccelConfig
	FFmpegPath      string
	FFProbePath     string
	CleanupInterval time.Duration
	InactiveTimeout time.Duration
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTranscoder creates a new transcoder
func NewTranscoder(config *TranscoderConfig) *Transcoder {
	return &Transcoder{
		OutputDir:       config.OutputDir,
		HwAccel:         config.HwAccel,
		FFmpegPath:      config.FFmpegPath,
		FFProbePath:     config.FFProbePath,
		fileStreams:     NewCMap[string, *FileStream](),
		CleanupInterval: config.CleanupInterval,
		InactiveTimeout: config.InactiveTimeout,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetFileStream returns a file stream for the given asset ID
func (t *Transcoder) GetFileStream(assetID, filePath string, keyframes []float64) *FileStream {
	return t.fileStreams.GetOrCreate(assetID, func() *FileStream {
		outputDir := filepath.Join(t.OutputDir, assetID)
		return NewFileStream(assetID, filePath, outputDir, keyframes, t.HwAccel, t.FFmpegPath, t.FFProbePath)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetMasterPlaylist returns the master playlist for an asset
func (t *Transcoder) GetMasterPlaylist(assetID string) (string, error) {
	stream, exists := t.fileStreams.Get(assetID)
	if !exists {
		return "", fmt.Errorf("file stream not found for asset %s", assetID)
	}

	return stream.GenerateMasterPlaylist()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoPlaylist returns the video playlist for an asset and quality
func (t *Transcoder) GetVideoPlaylist(assetID string, quality Quality) (string, error) {
	stream, exists := t.fileStreams.Get(assetID)
	if !exists {
		return "", fmt.Errorf("file stream not found for asset %s", assetID)
	}

	videoStream := stream.GetVideoStream(quality)
	return videoStream.GenerateM3U8Index()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegment returns a segment for an asset, quality, and index
func (t *Transcoder) GetSegment(ctx context.Context, assetID string, quality Quality, index int) (string, error) {
	stream, exists := t.fileStreams.Get(assetID)
	if !exists {
		return "", fmt.Errorf("file stream not found for asset %s", assetID)
	}

	// Generate the segment if it doesn't exist
	if err := stream.GenerateSegment(ctx, quality, index); err != nil {
		return "", fmt.Errorf("failed to generate segment: %w", err)
	}

	// Update last access time
	stream.UpdateLastAccess()

	return stream.GetSegmentPath(quality, index), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentInfo returns information about a segment
func (t *Transcoder) GetSegmentInfo(assetID string, quality Quality, index int) (startTime, duration float64, exists bool, err error) {
	stream, exists := t.fileStreams.Get(assetID)
	if !exists {
		return 0, 0, false, fmt.Errorf("file stream not found for asset %s", assetID)
	}

	startTime, duration, exists = stream.GetSegmentInfo(quality, index)
	return startTime, duration, exists, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentCount returns the total number of segments for an asset
func (t *Transcoder) GetSegmentCount(assetID string) (int, error) {
	stream, exists := t.fileStreams.Get(assetID)
	if !exists {
		return 0, fmt.Errorf("file stream not found for asset %s", assetID)
	}

	return stream.GetSegmentCount(), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsReady checks if an asset is ready for playback
func (t *Transcoder) IsReady(assetID string) bool {
	stream, exists := t.fileStreams.Get(assetID)
	if !exists {
		return false
	}

	return stream.IsReady()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetReadySegments returns the number of ready segments for an asset and quality
func (t *Transcoder) GetReadySegments(assetID string, quality Quality) int {
	stream, exists := t.fileStreams.Get(assetID)
	if !exists {
		return 0
	}

	return stream.GetReadySegments(quality)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// StartCleanup starts the cleanup goroutine
func (t *Transcoder) StartCleanup(ctx context.Context) {
	go t.cleanupLoop(ctx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// cleanupLoop runs the cleanup process
func (t *Transcoder) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(t.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.cleanupInactiveStreams()
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// cleanupInactiveStreams removes streams that have been inactive for too long
func (t *Transcoder) cleanupInactiveStreams() {
	now := time.Now()

	t.fileStreams.ForEach(func(assetID string, stream *FileStream) {
		lastAccess := stream.GetLastAccess()
		if now.Sub(lastAccess) > t.InactiveTimeout {
			// Cleanup the stream
			stream.Cleanup()
			t.fileStreams.Remove(assetID)
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CleanupAll removes all streams and cleans up the output directory
func (t *Transcoder) CleanupAll() error {
	// Cleanup all streams
	t.fileStreams.ForEach(func(assetID string, stream *FileStream) {
		stream.Cleanup()
	})

	// Clear streams map
	t.fileStreams.Clear()

	// Remove output directory
	return os.RemoveAll(t.OutputDir)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetActiveStreams returns the number of active streams
func (t *Transcoder) GetActiveStreams() int {
	return t.fileStreams.Len()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetStreamInfo returns information about a stream
func (t *Transcoder) GetStreamInfo(assetID string) (filePath string, segmentCount int, ready bool, err error) {
	stream, exists := t.fileStreams.Get(assetID)
	if !exists {
		return "", 0, false, fmt.Errorf("file stream not found for asset %s", assetID)
	}

	return stream.FilePath, stream.GetSegmentCount(), stream.IsReady(), nil
}
