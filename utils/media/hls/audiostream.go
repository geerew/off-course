package hls

import (
	"fmt"

	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AudioStream represents an audio transcoding stream
type AudioStream struct {
	Stream
	index uint32
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAudioStream creates a new audio stream
func NewAudioStream(file *FileStream, idx uint32) (*AudioStream, error) {
	utils.Infof("Creating a audio stream %d for %s\n", idx, file.Info.Path)

	// Create keyframes for audio (every 4 seconds as per Kyoo)
	// For now, we'll create a simple keyframe structure
	// In the real implementation, this will come from the database
	keyframes := NewKeyframeFromSlice([]float64{0.0, 4.0, 8.0, 12.0, 16.0, 20.0, 24.0, 28.0, 32.0, 36.0}, true)

	ret := &AudioStream{
		index: idx,
	}
	NewStream(file, keyframes, ret, &ret.Stream)
	return ret, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getOutPath returns the output path pattern for segments
func (as *AudioStream) getOutPath(encoderID int) string {
	return fmt.Sprintf("%s/segment-a%d-%d-%%d.ts", as.file.Out, as.index, encoderID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getFlags returns the stream flags for audio
func (as *AudioStream) getFlags() Flags {
	return AudioF
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getTranscodeArgs returns the FFmpeg arguments for audio transcoding
func (as *AudioStream) getTranscodeArgs(segments string) []string {
	return []string{
		"-map", fmt.Sprintf("0:a:%d", as.index),
		"-c:a", "aac",
		// TODO: Support 5.1 audio streams.
		"-ac", "2",
		// TODO: Support multi audio qualities.
		"-b:a", "128k",
	}
}
