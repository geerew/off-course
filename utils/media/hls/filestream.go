package hls

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// FileStream represents a file stream that manages video streams for different qualities
type FileStream struct {
	ID          string
	FilePath    string
	OutputDir   string
	Keyframes   []float64
	HwAccel     *HwAccelConfig
	FFmpegPath  string
	FFProbePath string

	// Video streams by quality
	videoStreams map[Quality]*VideoStream

	// Metadata
	Duration  float64
	Width     int
	Height    int
	Framerate float64
	Codec     string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewFileStream creates a new file stream
func NewFileStream(id, filePath, outputDir string, keyframes []float64, hwAccel *HwAccelConfig, ffmpegPath, ffprobePath string) *FileStream {
	return &FileStream{
		ID:           id,
		FilePath:     filePath,
		OutputDir:    outputDir,
		Keyframes:    keyframes,
		HwAccel:      hwAccel,
		FFmpegPath:   ffmpegPath,
		FFProbePath:  ffprobePath,
		videoStreams: make(map[Quality]*VideoStream),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoStream returns a video stream for the given quality
func (fs *FileStream) GetVideoStream(quality Quality) *VideoStream {
	if stream, exists := fs.videoStreams[quality]; exists {
		return stream
	}

	// Create new video stream
	stream := fs.createVideoStream(quality)
	fs.videoStreams[quality] = stream
	return stream
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// createVideoStream creates a new video stream for the given quality
func (fs *FileStream) createVideoStream(quality Quality) *VideoStream {
	// Create base stream
	baseStream := NewStream(
		fmt.Sprintf("%s-%s", fs.ID, quality),
		fs.FilePath,
		quality,
		filepath.Join(fs.OutputDir, string(quality)),
		fs.Keyframes,
		fs.HwAccel,
		fs.FFmpegPath,
		fs.FFProbePath,
	)

	// Get quality-specific parameters
	width, height, bitrate := fs.getQualityParams(quality)

	return NewVideoStream(
		baseStream,
		width,
		height,
		bitrate,
		fs.Framerate,
		fs.Codec,
	)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getQualityParams returns the parameters for a given quality
func (fs *FileStream) getQualityParams(quality Quality) (width, height, bitrate int) {
	if quality.IsOriginal() {
		return fs.Width, fs.Height, 0 // Use original bitrate
	}

	// Get target resolution
	targetHeight := quality.Height()
	if targetHeight == 0 {
		return fs.Width, fs.Height, 0
	}

	// Calculate aspect ratio
	aspectRatio := float64(fs.Width) / float64(fs.Height)
	targetWidth := int(float64(targetHeight) * aspectRatio)

	// Get bitrate
	bitrate = int(quality.AverageBitrate())

	return targetWidth, int(targetHeight), bitrate
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenerateMasterPlaylist generates the master M3U8 playlist
func (fs *FileStream) GenerateMasterPlaylist() (string, error) {
	masterPath := filepath.Join(fs.OutputDir, "master.m3u8")

	// Ensure output directory exists
	if err := os.MkdirAll(fs.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create master playlist content
	content := fs.buildMasterPlaylistContent()

	// Write to file
	if err := os.WriteFile(masterPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write master playlist: %w", err)
	}

	return masterPath, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// buildMasterPlaylistContent builds the master playlist content
func (fs *FileStream) buildMasterPlaylistContent() string {
	var builder strings.Builder

	// M3U8 header
	builder.WriteString("#EXTM3U\n")
	builder.WriteString("#EXT-X-VERSION:3\n")

	// Add video streams
	for _, quality := range Qualities {
		stream := fs.GetVideoStream(quality)
		width, height, bitrate, _ := stream.GetQualityInfo()

		// Stream info
		builder.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d,NAME=\"%s\"\n",
			bitrate, width, height, quality))
		builder.WriteString(fmt.Sprintf("%s/index.m3u8\n", quality))
	}

	return builder.String()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenerateSegment generates a segment for a specific quality
func (fs *FileStream) GenerateSegment(ctx context.Context, quality Quality, index int) error {
	stream := fs.GetVideoStream(quality)
	return stream.GenerateSegment(ctx, index)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentPath returns the path to a segment file
func (fs *FileStream) GetSegmentPath(quality Quality, index int) string {
	stream := fs.GetVideoStream(quality)
	segment := stream.GetSegment(index)
	return segment.FilePath
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentInfo returns information about a segment
func (fs *FileStream) GetSegmentInfo(quality Quality, index int) (startTime, duration float64, exists bool) {
	stream := fs.GetVideoStream(quality)
	return stream.GetSegmentInfo(index)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentCount returns the total number of segments
func (fs *FileStream) GetSegmentCount() int {
	return len(fs.Keyframes)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsReady checks if the file stream is ready for playback
func (fs *FileStream) IsReady() bool {
	// Check if we have at least the first segment for original quality
	stream := fs.GetVideoStream(Original)
	return stream.IsReady()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetReadySegments returns the number of ready segments for a quality
func (fs *FileStream) GetReadySegments(quality Quality) int {
	stream := fs.GetVideoStream(quality)
	return stream.GetReadySegments()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Cleanup removes all segments and streams
func (fs *FileStream) Cleanup() error {
	// Cleanup all video streams
	for _, stream := range fs.videoStreams {
		stream.Cleanup()
	}

	// Clear streams map
	fs.videoStreams = make(map[Quality]*VideoStream)

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetLastAccess returns the last access time
func (fs *FileStream) GetLastAccess() time.Time {
	// Return the most recent access time from all streams
	var lastAccess time.Time
	for _, stream := range fs.videoStreams {
		access := stream.GetLastAccess()
		if access.After(lastAccess) {
			lastAccess = access
		}
	}
	return lastAccess
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateLastAccess updates the last access time for all streams
func (fs *FileStream) UpdateLastAccess() {
	for _, stream := range fs.videoStreams {
		stream.UpdateLastAccess()
	}
}
