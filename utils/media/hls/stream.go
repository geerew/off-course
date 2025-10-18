package hls

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Stream represents a base stream for HLS transcoding
type Stream struct {
	ID          string
	FilePath    string
	Quality     Quality
	OutputDir   string
	Keyframes   []float64
	HwAccel     *HwAccelConfig
	FFmpegPath  string
	FFProbePath string

	// State
	mu         sync.RWMutex
	segments   map[int]*Segment
	lastAccess time.Time
	active     bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Segment represents a single HLS segment
type Segment struct {
	Index      int
	StartTime  float64
	Duration   float64
	FilePath   string
	Exists     bool
	Generating bool
	mu         sync.RWMutex
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewStream creates a new stream
func NewStream(id, filePath string, quality Quality, outputDir string, keyframes []float64, hwAccel *HwAccelConfig, ffmpegPath, ffprobePath string) *Stream {
	return &Stream{
		ID:          id,
		FilePath:    filePath,
		Quality:     quality,
		OutputDir:   outputDir,
		Keyframes:   keyframes,
		HwAccel:     hwAccel,
		FFmpegPath:  ffmpegPath,
		FFProbePath: ffprobePath,
		segments:    make(map[int]*Segment),
		lastAccess:  time.Now(),
		active:      true,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegment returns a segment by index
func (s *Stream) GetSegment(index int) *Segment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if segment, exists := s.segments[index]; exists {
		return segment
	}

	// Create new segment if it doesn't exist
	segment := &Segment{
		Index:      index,
		StartTime:  s.getSegmentStartTime(index),
		Duration:   s.getSegmentDuration(index),
		FilePath:   s.getSegmentFilePath(index),
		Exists:     false,
		Generating: false,
	}

	s.segments[index] = segment
	return segment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GenerateSegment generates a segment if it doesn't exist
func (s *Stream) GenerateSegment(ctx context.Context, index int) error {
	segment := s.GetSegment(index)

	segment.mu.Lock()
	defer segment.mu.Unlock()

	// Check if already exists
	if segment.Exists {
		return nil
	}

	// Check if already generating
	if segment.Generating {
		// Wait for generation to complete
		for segment.Generating && !segment.Exists {
			time.Sleep(100 * time.Millisecond)
		}
		return nil
	}

	// Mark as generating
	segment.Generating = true

	// Generate the segment
	err := s.generateSegmentFile(ctx, segment)

	segment.Generating = false
	if err == nil {
		segment.Exists = true
	}

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// generateSegmentFile creates the actual segment file using FFmpeg
func (s *Stream) generateSegmentFile(ctx context.Context, segment *Segment) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(segment.FilePath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command
	args := s.buildFFmpegArgs(segment)

	cmd := exec.CommandContext(ctx, s.FFmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// buildFFmpegArgs builds the FFmpeg command arguments for segment generation
func (s *Stream) buildFFmpegArgs(segment *Segment) []string {
	args := []string{
		"-nostats", "-hide_banner", "-loglevel", "warning",
	}

	// Add hardware acceleration decode flags BEFORE input file (like Kyoo)
	if s.HwAccel != nil && s.HwAccel.IsHardwareAccelerated() {
		args = append(args, s.HwAccel.DecodeFlags...)
	}

	// Add input file and timing
	args = append(args,
		"-ss", fmt.Sprintf("%.3f", segment.StartTime),
		"-t", fmt.Sprintf("%.3f", segment.Duration),
		"-avoid_negative_ts", "make_zero",
		"-fflags", "+genpts",
		"-i", s.FilePath,
		"-start_at_zero",
		"-copyts",
		"-muxdelay", "0",
	)

	// Add encoding flags AFTER input file
	if s.HwAccel != nil && s.HwAccel.IsHardwareAccelerated() {
		args = append(args, s.HwAccel.EncodeFlags...)
		// TODO: Add scale filter for hardware acceleration when needed
		// For now, skip the scale filter to avoid the "No filters specified" error
	} else {
		// Software fallback with proper flags
		args = append(args, "-c:v", "libx264", "-preset", "fast", "-crf", "23", "-sc_threshold", "0", "-pix_fmt", "yuv420p")
	}

	// Add audio codec
	args = append(args, "-c:a", "aac", "-b:a", "128k")

	// Output format
	args = append(args, "-f", "mpegts", segment.FilePath)

	return args
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getSegmentStartTime returns the start time for a segment
func (s *Stream) getSegmentStartTime(index int) float64 {
	if index < 0 || index >= len(s.Keyframes) {
		return 0
	}
	return s.Keyframes[index]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getSegmentDuration returns the duration for a segment
func (s *Stream) getSegmentDuration(index int) float64 {
	if index < 0 || index >= len(s.Keyframes) {
		return 0
	}

	// If this is the last segment, we can't determine duration
	// For now, use a default duration
	if index == len(s.Keyframes)-1 {
		return 4.0 // Default 4 seconds
	}

	return s.Keyframes[index+1] - s.Keyframes[index]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getSegmentFilePath returns the file path for a segment
func (s *Stream) getSegmentFilePath(index int) string {
	return filepath.Join(s.OutputDir, fmt.Sprintf("segment-%d.ts", index))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentCount returns the total number of segments
func (s *Stream) GetSegmentCount() int {
	return len(s.Keyframes)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateLastAccess updates the last access time
func (s *Stream) UpdateLastAccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastAccess = time.Now()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetLastAccess returns the last access time
func (s *Stream) GetLastAccess() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastAccess
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsActive returns true if the stream is active
func (s *Stream) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetActive sets the active state
func (s *Stream) SetActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = active
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Cleanup removes old segments
func (s *Stream) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove all segment files
	for _, segment := range s.segments {
		if segment.Exists {
			os.Remove(segment.FilePath)
		}
	}

	// Clear segments map
	s.segments = make(map[int]*Segment)
	s.active = false

	return nil
}
