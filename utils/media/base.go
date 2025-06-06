package media

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoInfo holds the metadata of a video file
type VideoInfo struct {
	Duration   int
	Width      int
	Height     int
	Codec      string
	Resolution string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type MediaProbe struct {
	FFProbePath string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ProbeVideo uses ffprobe to extract metadata from a video file
func (mp MediaProbe) ProbeVideo(filepath string) (*VideoInfo, error) {
	ffprobePath, err := mp.resolveFFProbePath()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filepath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running ffprobe: %w", err)
	}

	// Parse JSON
	var probe probeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("error parsing ffprobe output: %w", err)
	}

	// Find the first video stream with dimensions
	var videoStream *stream
	for _, stream := range probe.Streams {
		if stream.Width > 0 && stream.Height > 0 {
			videoStream = &stream
			break
		}
	}

	if videoStream == nil {
		return nil, fmt.Errorf("no valid video stream found in: %s", filepath)
	}

	// Use stream duration if available, else fallback to format duration
	durationStr := videoStream.Duration
	if durationStr == "" {
		durationStr = probe.Format.Duration
	}

	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid duration: %q", durationStr)
	}

	duration := int(math.Round(durationFloat))

	return &VideoInfo{
		Duration:   duration,
		Width:      videoStream.Width,
		Height:     videoStream.Height,
		Resolution: resolutionLabel(videoStream.Height),
		Codec:      videoStream.CodecName,
	}, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// resolveProbePath checks if the ffprobe path is valid
func (mp MediaProbe) resolveFFProbePath() (string, error) {
	// Default to system ffprobe and check if it's available
	if mp.FFProbePath == "" || mp.FFProbePath == "ffprobe" {
		if _, err := exec.LookPath("ffprobe"); err == nil {
			return "ffprobe", nil
		}

		return "", utils.ErrFFProbeUnavailable
	}

	// If user provided a full path to ffprobe
	if strings.HasSuffix(mp.FFProbePath, "ffprobe") || strings.HasSuffix(mp.FFProbePath, "ffprobe.exe") {
		if _, err := os.Stat(mp.FFProbePath); err == nil {
			return mp.FFProbePath, nil
		}
		return "", utils.ErrInvalidFFProbePath
	}

	// If user provided a directory, check for ffprobe inside
	candidate := filepath.Join(mp.FFProbePath, "ffprobe")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	candidateExe := candidate + ".exe"
	if _, err := os.Stat(candidateExe); err == nil {
		return candidateExe, nil
	}

	return "", utils.ErrFFProbeNotFound
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// resolutionLabel returns a human-readable label for the video resolution
func resolutionLabel(height int) string {
	switch {
	case height >= 4320:
		return "8K"
	case height >= 2160:
		return "4K"
	case height >= 1440:
		return "1440p"
	case height >= 1080:
		return "1080p"
	case height >= 720:
		return "720p"
	case height >= 480:
		return "480p"
	default:
		return fmt.Sprintf("%dp", height)
	}
}
