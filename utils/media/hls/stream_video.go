package hls

import (
	"fmt"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoStream represents a video transcoding stream
type VideoStream struct {
	Stream
	video   *Video
	quality Quality
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewVideoStream creates a new video stream for the given file, index and quality
func NewVideoStream(sw *StreamWrapper, videoIndex uint32, quality Quality) (*VideoStream, error) {
	Settings.Logger.Debug().
		Str("path", sw.Info.Path).
		Uint32("video_index", videoIndex).
		Str("quality", string(quality)).
		Msg("Creating a new video stream")

	// Find the video metadata from the file's info
	var video *Video
	for _, v := range sw.Info.Videos {
		if v.Index == videoIndex {
			video = &v
			break
		}
	}

	if video == nil {
		return nil, fmt.Errorf("video stream %d not found", videoIndex)
	}

	videoStream := &VideoStream{
		Stream: Stream{
			streamWrapper: sw,
			heads:         make([]Head, 0),
			logger:        Settings.Logger,
		},
		quality: quality,
		video:   video,
	}

	videoStream.streamer = videoStream
	videoStream.keyframes = getKeyframes(sw)
	videoStream.initializeSegments()

	return videoStream, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getFlags returns the stream flags for video
func (vs *VideoStream) getFlags() Flags {
	if vs.quality == Original {
		return VideoF | Transmux
	}
	return VideoF
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getOutPath returns the output path pattern for segments
func (vs *VideoStream) getOutPath(encoderID int) string {
	return fmt.Sprintf("%s/segment-%s-%d-%%d.ts", vs.streamWrapper.Out, vs.quality, encoderID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getTranscodeArgs returns the FFmpeg arguments for transcoding
func (vs *VideoStream) getTranscodeArgs(segments string) []string {
	args := []string{
		"-map", fmt.Sprintf("0:V:%d", vs.video.Index),
	}

	// In original mode, we don't need to transcode the video
	if vs.quality == Original {
		args = append(args, "-c:v", "copy")
		return args
	}

	args = append(args, Settings.HwAccel.EncodeFlags...)

	quality := vs.quality
	if vs.quality != NoResize {
		width := int32(float64(vs.quality.Height()) / float64(vs.video.Height) * float64(vs.video.Width))
		// force a width that is a multiple of two else some apps behave badly
		width = closestMultiple(width, 2)
		args = append(args,
			"-vf", fmt.Sprintf(Settings.HwAccel.ScaleFilter, width, vs.quality.Height()),
		)
	} else {
		// Only add video filter if NoResizeFilter is defined (not empty)
		if Settings.HwAccel.NoResizeFilter != "" {
			args = append(args, "-vf", Settings.HwAccel.NoResizeFilter)
		}

		// NoResize doesn't have bitrate info, fallback to a know quality higher or equal
		for _, q := range Qualities {
			if q.Height() >= vs.video.Height {
				quality = q
				break
			}
		}
	}

	args = append(args,
		// Even less sure but bufsize are 5x the average bitrate since the average bitrate is only
		// useful for hls segments
		"-bufsize", fmt.Sprint(quality.MaxBitrate()*5),
		"-b:v", fmt.Sprint(quality.AverageBitrate()),
		"-maxrate", fmt.Sprint(quality.MaxBitrate()),
		// Force segments to be split exactly on keyframes (only works when transcoding)
		// forced-idr is needed to force keyframes to be an idr-frame (by default it can be any i frames)
		// without this option, some hardware encoders uses others i-frames and the -f segment can't cut at them
		"-forced-idr", "1",
		"-force_key_frames", segments,
		// make ffmpeg globally less buggy
		"-strict", "-2",
	)

	return args
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// closestMultiple finds the closest multiple of x that is >= n
func closestMultiple(n int32, x int32) int32 {
	if x > n {
		return x
	}

	n = n + x/2
	n = n - (n % x)
	return n
}
