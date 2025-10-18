package hls

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoStream represents a video stream for HLS transcoding
type VideoStream struct {
	*Stream
	Width     int
	Height    int
	Bitrate   int
	Framerate float64
	Codec     string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewVideoStream creates a new video stream
func NewVideoStream(stream *Stream, width, height, bitrate int, framerate float64, codec string) *VideoStream {
	return &VideoStream{
		Stream:    stream,
		Width:     width,
		Height:    height,
		Bitrate:   bitrate,
		Framerate: framerate,
		Codec:     codec,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenerateM3U8Index generates the M3U8 index file for this stream
func (vs *VideoStream) GenerateM3U8Index() (string, error) {
	indexPath := filepath.Join(vs.OutputDir, "index.m3u8")

	// Ensure output directory exists
	if err := os.MkdirAll(vs.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create M3U8 content
	content := vs.buildM3U8Content()

	// Write to file
	if err := os.WriteFile(indexPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write M3U8 index: %w", err)
	}

	return indexPath, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// buildM3U8Content builds the M3U8 playlist content
func (vs *VideoStream) buildM3U8Content() string {
	var builder strings.Builder

	// M3U8 header
	builder.WriteString("#EXTM3U\n")
	builder.WriteString("#EXT-X-VERSION:3\n")
	builder.WriteString("#EXT-X-TARGETDURATION:4\n")
	builder.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	// Add segments
	for i := 0; i < vs.GetSegmentCount(); i++ {
		duration := vs.getSegmentDuration(i)

		builder.WriteString(fmt.Sprintf("#EXTINF:%.3f,\n", duration))
		builder.WriteString(fmt.Sprintf("segment-%d.ts\n", i))
	}

	// End of playlist
	builder.WriteString("#EXT-X-ENDLIST\n")

	return builder.String()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenerateSegment generates a specific segment
func (vs *VideoStream) GenerateSegment(ctx context.Context, index int) error {
	return vs.Stream.GenerateSegment(ctx, index)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentInfo returns information about a segment
func (vs *VideoStream) GetSegmentInfo(index int) (startTime, duration float64, exists bool) {
	segment := vs.GetSegment(index)
	return segment.StartTime, vs.getSegmentDuration(index), segment.Exists
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getSegmentDuration returns the duration for a segment
func (vs *VideoStream) getSegmentDuration(index int) float64 {
	if index < 0 || index >= len(vs.Keyframes) {
		return 0
	}

	// If this is the last segment, use a default duration
	if index == len(vs.Keyframes)-1 {
		return 4.0 // Default 4 seconds
	}

	return vs.Keyframes[index+1] - vs.Keyframes[index]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetQualityInfo returns quality information
func (vs *VideoStream) GetQualityInfo() (width, height, bitrate int, codec string) {
	return vs.Width, vs.Height, vs.Bitrate, vs.Codec
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsReady checks if the stream is ready for playback
func (vs *VideoStream) IsReady() bool {
	// Check if we have at least the first segment
	segment := vs.GetSegment(0)
	return segment.Exists
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetReadySegments returns the number of ready segments
func (vs *VideoStream) GetReadySegments() int {
	count := 0
	for i := 0; i < vs.GetSegmentCount(); i++ {
		segment := vs.GetSegment(i)
		if segment.Exists {
			count++
		}
	}
	return count
}
