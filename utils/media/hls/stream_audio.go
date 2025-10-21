package hls

import (
	"context"
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

// NewAudioStream creates a new audio stream for the given file and index
func NewAudioStream(wrapper *StreamWrapper, audioIndex uint32) (*AudioStream, error) {
	utils.Infof("HLS: Creating an audio stream %d for %s\n", audioIndex, wrapper.Info.Path)

	ret := &AudioStream{index: audioIndex}

	// Get keyframes from database
	assetKeyframes, err := wrapper.transcoder.dao.GetAssetKeyframes(context.Background(), wrapper.transcoder.assetID)
	if err != nil {
		utils.Errf("HLS: Failed to get keyframes: %v\n", err)

		NewStream(wrapper, []float64{}, ret, &ret.Stream)
		return ret, nil
	}

	keyframes := []float64{}
	if assetKeyframes != nil && len(assetKeyframes.Keyframes) > 0 {
		keyframes = assetKeyframes.Keyframes
	}

	NewStream(wrapper, keyframes, ret, &ret.Stream)
	return ret, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getOutPath returns the output path pattern for segments
func (as *AudioStream) getOutPath(encoderID int) string {
	return fmt.Sprintf("%s/segment-a%d-%d-%%d.ts", as.wrapper.Out, as.index, encoderID)
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
		// TODO: Support 5.1 audio streams
		"-ac", "2",
		// TODO: Support multi audio qualities
		"-b:a", "128k",
	}
}
