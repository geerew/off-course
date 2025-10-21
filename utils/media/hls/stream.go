// Package hls provides on-demand HLS transcoding and segmentation with
// adaptive prefetching, hardware acceleration, and per-stream tracking
package hls

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Head represents an encoding process
type Head struct {
	segment int32
	end     int32
	command *exec.Cmd
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeletedHead is a marker for a head that has been killed
var DeletedHead = Head{
	segment: -1,
	end:     -1,
	command: nil,
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Segment represents a single HLS segment
type Segment struct {
	channel chan struct{}
	encoder int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Flags represents stream type flags
type Flags int32

const (
	AudioF   Flags = 1 << 0
	VideoF   Flags = 1 << 1
	Transmux Flags = 1 << 3
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoKey uniquely identifies a video stream by index and quality
type VideoKey struct {
	idx     uint32
	quality Quality
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// StreamHandle represents a stream handle


// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Stream represents a transcoding stream with multiple encoding heads
type Stream struct {
	handle    StreamHandle
	wrapper   *StreamWrapper
	keyframes []float64
	segments  []Segment
	heads     []Head
	lock      sync.RWMutex
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initializeSegments creates and initializes segments based on keyframes
func (s *Stream) initializeSegments() {
	length := len(s.keyframes)
	s.segments = make([]Segment, length, max(length, 2000))
	for seg := range s.segments {
		s.segments[seg].channel = make(chan struct{})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isSegmentReady checks if a segment is ready (non-blocking)
// Remember to lock before calling this.
func (s *Stream) isSegmentReady(segment int32) bool {
	select {
	case <-s.segments[segment].channel:
		// if the channel returned, it means it was closed
		return true
	default:
		return false
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isSegmentTranscoding checks if a segment is currently being transcoded
func (s *Stream) isSegmentTranscoding(segment int32) bool {
	for _, head := range s.heads {
		if head.segment == segment {
			return true
		}
	}
	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// run starts transcoding from the given segment
func (s *Stream) run(start int32) error {
	// Start the transcode with adaptive buffer based on video length
	length := len(s.keyframes)

	// Calculate smart buffer size based on video duration
	videoDuration := float64(s.wrapper.Info.Duration)
	var bufferSegments int32

	if videoDuration <= 300 { // 5 minutes or less
		bufferSegments = 15 // ~80 seconds ahead
	} else if videoDuration <= 600 { // 10 minutes or less
		bufferSegments = 20 // ~107 seconds ahead
	} else { // Longer videos
		bufferSegments = 25 // ~133 seconds ahead
	}

	end := min(start+bufferSegments, int32(length))
	// Stop at the first finished segment
	s.lock.Lock()
	for i := start; i < end; i++ {
		if s.isSegmentReady(i) || s.isSegmentTranscoding(i) {
			end = i
			break
		}
	}
	if start >= end {
		// this can happens if the start segment was finished between the check
		// to call run() and the actual call.
		// since most checks are done in a RLock() instead of a Lock() this can
		// happens when two goroutines try to make the same segment ready
		s.lock.Unlock()
		return nil
	}
	encoderID := len(s.heads)
	s.heads = append(s.heads, Head{segment: start, end: end, command: nil})
	s.lock.Unlock()

	utils.Infof(
		"HLS: Starting transcode %d for %s (from %d to %d out of %d segments)\n",
		encoderID,
		s.wrapper.Info.Path,
		start,
		end,
		length,
	)

	// Include both the start and end delimiter because -ss and -to are not accurate
	// Having an extra segment allows us to cut precisely the segments we want with the
	// -f segment that cuts the beginning and the end at the keyframe as requested
	startRef := float64(0)
	startSegment := start
	if start != 0 {
		// we always take on segment before the current one, for different reasons for audio/video:
		//  - Audio: we need context before the starting point, without that ffmpeg doesn't know what to do and leaves ~100ms of silence
		//  - Video: if a segment is really short (between 20 and 100ms), the padding given in the else block below is not enough and
		// the previous segment is played another time. the -segment_times is way more precise so it does not do the same with this one
		startSegment = start - 1
		if s.handle.getFlags()&AudioF != 0 {
			startRef = s.keyframes[startSegment]
		} else {
			// the param for the -ss takes the keyframe before the specified time
			// (if the specified time is a keyframe, it either takes that keyframe or the one before)
			// to prevent this weird behavior, we specify a bit after the keyframe that interest us

			// this can't be used with audio since we need to have context before the start-time
			// without this context, the cut loses a bit of audio (audio gap of ~100ms)
			if startSegment+1 == int32(length) {
				startRef = (s.keyframes[startSegment] + float64(s.wrapper.Info.Duration)) / 2
			} else {
				startRef = (s.keyframes[startSegment] + s.keyframes[startSegment+1]) / 2
			}
		}
	}
	endPadding := int32(1)
	if end == int32(length) {
		endPadding = 0
	}
	segments := s.keyframes[startSegment+1 : end+endPadding]
	if len(segments) == 0 {
		// we can't leave that empty else ffmpeg errors out.
		segments = []float64{9999999}
	}

	outPath := s.handle.getOutPath(encoderID)
	err := os.MkdirAll(filepath.Dir(outPath), 0o755)
	if err != nil {
		return err
	}

	args := []string{
		"-nostats", "-hide_banner", "-loglevel", "warning",
	}

	if s.handle.getFlags()&VideoF != 0 {
		args = append(args, Settings.HwAccel.DecodeFlags...)
	}

	if startRef != 0 {
		if s.handle.getFlags()&VideoF != 0 {
			// This is the default behavior in transmux mode and needed to force pre/post segment to work
			// This must be disabled when processing only audio because it creates gaps in audio
			args = append(args, "-noaccurate_seek")
		}
		args = append(args,
			"-ss", fmt.Sprintf("%.6f", startRef),
		)
	}
	// do not include -to if we want the file to go to the end
	if end+1 < int32(length) {
		// sometimes, the duration is shorter than expected (only during transcode it seems)
		// always include more and use the -f segment to split the file where we want
		end_ref := s.keyframes[end+1]
		// it seems that the -to is confused when -ss seek before the given time (because it searches for a keyframe)
		// add back the time that would be lost otherwise
		// this only happens when -to is before -i but having -to after -i gave a bug (not sure, don't remember)
		end_ref += startRef - s.keyframes[startSegment]
		args = append(args,
			"-to", fmt.Sprintf("%.6f", end_ref),
		)
	}
	args = append(args,
		// some avi files are missing pts, using this flag makes ffmpeg use dts as pts and prevents an error with
		// -c:v copy. Only issue: pts is sometime wrong (+1fps than expected) and this leads to some clients refusing
		// to play the file (they just switch back to the previous quality).
		// since this is better than errorring or not supporting transmux at all, i'll keep it here for now.
		"-fflags", "+genpts",
		"-i", s.wrapper.Info.Path,
		// this makes behaviors consistent between soft and hardware decodes.
		// this also means that after a -ss 50, the output video will start at 50s
		"-start_at_zero",
		// for hls streams, -copyts is mandatory
		"-copyts",
		// this makes output file start at 0s instead of a random delay + the -ss value
		// this also cancel -start_at_zero weird delay.
		// this is not always respected but generally it gives better resuls.
		// even when this is not respected, it does not result in a bugged experience but this is something
		// to keep in mind when debugging
		"-muxdelay", "0",
	)
	args = append(args, s.handle.getTranscodeArgs(toSegmentStr(segments))...)
	args = append(args,
		"-f", "segment",
		// needed for rounding issues when forcing keyframes
		// recommended value is 1/(2*frame_rate), which for a 24fps is ~0.021
		// we take a little bit more than that to be extra safe but too much can be harmful
		// when segments are short (can make the video repeat itself)
		"-segment_time_delta", "0.05",
		"-segment_format", "mpegts",
		"-segment_times", toSegmentStr(utils.Map(segments, func(seg float64) float64 {
			// segment_times want durations, not timestamps so we must subtract the -ss param
			// since we give a greater value to -ss to prevent wrong seeks but -segment_times
			// needs precise segments, we use the keyframe we want to seek to as a reference.
			return seg - s.keyframes[startSegment]
		})),
		"-segment_list_type", "flat",
		"-segment_list", "pipe:1",
		"-segment_start_number", fmt.Sprint(startSegment),
		outPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	utils.Infof("HLS: Running %s\n", strings.Join(cmd.Args, " "))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err = cmd.Start()
	if err != nil {
		return err
	}
	s.lock.Lock()
	s.heads[encoderID].command = cmd
	s.lock.Unlock()

	go func() {
		scanner := bufio.NewScanner(stdout)
		format := filepath.Base(outPath)
		shouldStop := false

		for scanner.Scan() {
			var segment int32
			_, _ = fmt.Sscanf(scanner.Text(), format, &segment)

			if segment < start {
				// This happens because we use -f segment for accurate cutting (since -ss is not)
				// check the comment at the beginning of function for more info
				continue
			}
			s.lock.Lock()
			s.heads[encoderID].segment = segment

			// Determine if this is audio or video stream
			streamType := "unknown"
			if s.handle.getFlags()&AudioF != 0 {
				streamType = "audio"
			} else if s.handle.getFlags()&VideoF != 0 {
				streamType = "video"
			}

			utils.Infof("HLS: %s segment %d is ready (encoder %d)\n", streamType, segment, encoderID)
			if s.isSegmentReady(segment) {
				// the current segment is already marked at done so another process has already gone up to here.
				cmd.Process.Signal(os.Interrupt)
				utils.Infof("HLS: Stopping %s encoder %d because segment %d is already ready\n", streamType, encoderID, segment)
				shouldStop = true
			} else {
				s.segments[segment].encoder = encoderID
				close(s.segments[segment].channel)
				if segment == end-1 {
					// file finished, ffmpeg will finish soon on its own
					shouldStop = true
				} else if s.isSegmentReady(segment + 1) {
					cmd.Process.Signal(os.Interrupt)
					utils.Infof("HLS: Killing ffmpeg because next segment %d is ready\n", segment)
					shouldStop = true
				}
			}
			s.lock.Unlock()
			// we need this and not a return in the condition because we want to unlock
			// the lock (and can't defer since this is a loop)
			if shouldStop {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			utils.Errf("HLS: Error reading stdout of ffmpeg: %v\n", err)
		}
	}()

	go func() {
		err := cmd.Wait()
		if exiterr, ok := err.(*exec.ExitError); ok && exiterr.ExitCode() == 255 {
			utils.Infof("HLS: ffmpeg %d was killed by us\n", encoderID)
		} else if err != nil {
			utils.Errf("HLS: ffmpeg %d occurred an error: %s: %s\n", encoderID, err, stderr.String())
		} else {
			utils.Infof("HLS: ffmpeg %d finished successfully\n", encoderID)
		}

		s.lock.Lock()
		defer s.lock.Unlock()
		// we can't delete the head directly because it would invalidate the others encoderID
		s.heads[encoderID] = DeletedHead
	}()

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetIndex generates the HLS index playlist for the stream
func (s *Stream) GetIndex() (string, error) {
	// Use VOD playlist type since keyframes are always complete (extracted during course scan)
	index := `#EXTM3U
#EXT-X-VERSION:6
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-INDEPENDENT-SEGMENTS
`
	length := len(s.keyframes) // Always complete since keyframes are extracted during course scan

	for segment := int32(0); segment < int32(length)-1; segment++ {
		index += fmt.Sprintf("#EXTINF:%.6f\n", s.keyframes[segment+1]-s.keyframes[segment])
		index += fmt.Sprintf("segment-%d.ts\n", segment)
	}
	// Always add the last segment and ENDLIST since keyframes are always complete
	index += fmt.Sprintf("#EXTINF:%.6f\n", float64(s.wrapper.Info.Duration)-s.keyframes[length-1])
	index += fmt.Sprintf("segment-%d.ts\n", length-1)
	index += `#EXT-X-ENDLIST`
	return index, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegment retrieves a specific segment path, starting transcoding if needed
func (s *Stream) GetSegment(segment int32) (string, error) {
	s.lock.RLock()
	ready := s.isSegmentReady(segment)
	// we want to calculate distance in the same lock else it can be funky
	distance := 0.
	isScheduled := false
	if !ready {
		distance = s.getMinEncoderDistance(segment)
		for _, head := range s.heads {
			if head.segment <= segment && segment < head.end {
				isScheduled = true
				break
			}
		}
	}
	readyChan := s.segments[segment].channel
	s.lock.RUnlock()

	if !ready {
		// Only start a new encode if there is too big a distance between the current encoder and the segment.
		if distance > 60 || !isScheduled {
			utils.Infof("HLS: Creating new head for %d since closest head is %fs away\n", segment, distance)
			err := s.run(segment)
			if err != nil {
				return "", err
			}
		} else {
			utils.Infof("HLS: Waiting for segment %d since encoder head is %fs away\n", segment, distance)
		}

		select {
		case <-readyChan:
		case <-time.After(60 * time.Second):
			return "", errors.New("could not retrieve the selected segment (timeout)")
		}
	}
	s.prepareNextSegments(segment)
	return fmt.Sprintf(s.handle.getOutPath(s.segments[segment].encoder), segment), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// prepareNextSegments starts encoding future segments for video streams
func (s *Stream) prepareNextSegments(segment int32) {
	// Audio is way cheaper to create than video so we don't need to run them in advance
	// Running it in advance might actually slow down the video encode since less compute
	// power can be used so we simply disable that.
	if s.handle.getFlags()&VideoF == 0 {
		return
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	for i := segment + 1; i <= min(segment+10, int32(len(s.segments)-1)); i++ {
		if s.isSegmentReady(i) {
			continue
		}
		// only start encode for segments not planned (getMinEncoderDistance returns Inf for them)
		// or if they are 60s away (assume 5s per segment)
		if s.getMinEncoderDistance(i) < 60+(5*float64(i-segment)) {
			continue
		}
		utils.Infof("HLS: Creating new head for future segment (%d)\n", i)
		go s.run(i)
		return
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getMinEncoderDistance calculates the minimum distance to any active encoder
func (s *Stream) getMinEncoderDistance(segment int32) float64 {
	time := s.keyframes[segment]
	distances := utils.Map(s.heads, func(head Head) float64 {
		// ignore killed heads or heads after the current time
		if head.segment < 0 || s.keyframes[head.segment] > time || segment >= head.end {
			return math.Inf(1)
		}
		return time - s.keyframes[head.segment]
	})
	if len(distances) == 0 {
		return math.Inf(1)
	}
	return slices.Min(distances)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Kill stops all encoding processes associated with the stream
func (s *Stream) Kill() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for id := range s.heads {
		s.KillHead(id)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// KillHead stops a specific encoding process
// Stream is assumed to be locked by the caller.
func (s *Stream) KillHead(encoder_id int) {
	if s.heads[encoder_id] == DeletedHead || s.heads[encoder_id].command == nil {
		return
	}
	s.heads[encoder_id].command.Process.Signal(os.Interrupt)
	s.heads[encoder_id] = DeletedHeadtype StreamHandle interface {
	getTranscodeArgs(segments string) []string
	getOutPath(encoderID int) string
	getFlags() Flags
}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getKeyframes retrieves keyframes from the database, falling back to empty slice on error
func getKeyframes(wrapper *StreamWrapper) []float64 {
	assetKeyframes, err := wrapper.transcoder.dao.GetAssetKeyframes(context.Background(), wrapper.transcoder.assetID)
	if err != nil {
		utils.Errf("HLS: Failed to get keyframes: %v\n", err)
		return []float64{}
	}

	if assetKeyframes != nil && len(assetKeyframes.Keyframes) > 0 {
		return assetKeyframes.Keyframes
	}
	return []float64{}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// toSegmentStr converts keyframe timestamps to a comma-separated string
func toSegmentStr(segments []float64) string {
	return strings.Join(utils.Map(segments, func(seg float64) string {
		return fmt.Sprintf("%.6f", seg)
	}), ",")
}
