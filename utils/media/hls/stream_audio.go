package hls

import (
	"fmt"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AudioStream represents an audio transcoding stream
type AudioStream struct {
	Stream
	index uint32
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAudioStream creates a new audio stream for the given file and index
func NewAudioStream(streamWrapper *StreamWrapper, audioIndex uint32) (*AudioStream, error) {
	streamWrapper.config.Logger.Debug().
		Str("asset_id", streamWrapper.assetID).
		Str("path", streamWrapper.Info.Path).
		Uint32("audio_index", audioIndex).
		Msg("Creating an audio stream")

	audioStream := &AudioStream{
		Stream: Stream{
			streamWrapper: streamWrapper,
			heads:         make([]Head, 0),
		},
		index: audioIndex,
	}

	audioStream.streamer = audioStream
	audioStream.keyframes = getKeyframes(streamWrapper)
	audioStream.initializeSegments()

	return audioStream, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getOutPath returns the output path pattern for segments
func (as *AudioStream) getOutPath(encoderID int) string {
	return fmt.Sprintf("%s/segment-a%d-%d-%%d.ts", as.streamWrapper.Out, as.index, encoderID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getFlags returns the stream flags for audio
func (as *AudioStream) getFlags() Flags {
	return AudioF
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getTranscodeArgs returns the FFmpeg arguments for audio transcoding
//
// TODO: Support 5.1 audio streams
// TODO: Support multi audio qualities
func (as *AudioStream) getTranscodeArgs(_ string) []string {
	return []string{
		"-map", fmt.Sprintf("0:a:%d", as.index),
		"-c:a", "aac",
		"-ac", "2",
		"-b:a", "128k",
	}
}
