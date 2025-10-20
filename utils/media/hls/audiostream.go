package hls

import (
	"context"
	"fmt"

	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AudioStream represents an audio transcoding stream.
type AudioStream struct {
	Stream
	index uint32
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAudioStream creates a new audio stream for the given file and index.
func NewAudioStream(file *FileStream, idx uint32) (*AudioStream, error) {
	utils.Infof("HLS: Creating an audio stream %d for %s\n", idx, file.Info.Path)

	ret := &AudioStream{
		index: idx,
	}

	// Get keyframes from database (like original Kyoo)
	assetKeyframes, err := file.transcoder.dao.GetAssetKeyframes(context.Background(), file.transcoder.assetID)
	if err != nil {
		utils.Errf("HLS: Failed to get keyframes: %v\n", err)
		// Fallback to empty keyframes
		keyframes := NewKeyframeFromSlice([]float64{}, false)
		NewStream(file, keyframes, ret, &ret.Stream)
		return ret, nil
	}

	// Convert database keyframes to HLS keyframes (same as video stream)
	var keyframeTimes []float64
	if assetKeyframes != nil && len(assetKeyframes.Keyframes) > 0 {
		keyframeTimes = assetKeyframes.Keyframes
	}
	keyframes := NewKeyframeFromSlice(keyframeTimes, assetKeyframes.IsComplete)
	NewStream(file, keyframes, ret, &ret.Stream)
	return ret, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getOutPath returns the output path pattern for segments.
func (as *AudioStream) getOutPath(encoderID int) string {
	return fmt.Sprintf("%s/segment-a%d-%d-%%d.ts", as.file.Out, as.index, encoderID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getFlags returns the stream flags for audio.
func (as *AudioStream) getFlags() Flags {
	return AudioF
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getTranscodeArgs returns the FFmpeg arguments for audio transcoding.
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
