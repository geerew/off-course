package media

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// FFmpeg holds the paths to ffmpeg and ffprobe executables
type FFmpeg struct {
	FFmpegPath  string
	FFProbePath string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewFFmpeg creates a new FFmpeg instance by resolving the paths to ffmpeg and ffprobe
// It will error if either executable is not found on the system PATH
func NewFFmpeg() (*FFmpeg, error) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, utils.ErrFFmpegUnavailable
	}

	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, utils.ErrFFProbeUnavailable
	}

	return &FFmpeg{
		FFmpegPath:  ffmpegPath,
		FFProbePath: ffprobePath,
	}, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewFFmpegWithPaths creates a new FFmpeg instance with custom paths
// It will validate that the provided paths exist and are executable
func NewFFmpegWithPaths(ffmpegPath, ffprobePath string) (*FFmpeg, error) {
	// Validate ffmpeg path
	if ffmpegPath == "" {
		return nil, utils.ErrFFmpegPathEmpty
	}

	// Check if it's a system command or full path
	if ffmpegPath == "ffmpeg" {
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			return nil, utils.ErrFFmpegUnavailable
		}
	} else {
		// Check if it's a full path to ffmpeg
		if strings.HasSuffix(ffmpegPath, "ffmpeg") || strings.HasSuffix(ffmpegPath, "ffmpeg.exe") {
			if _, err := exec.LookPath(ffmpegPath); err != nil {
				return nil, utils.ErrInvalidFFmpegPath
			}
		} else {
			// Check if it's a directory containing ffmpeg
			candidate := filepath.Join(ffmpegPath, "ffmpeg")
			if _, err := exec.LookPath(candidate); err == nil {
				ffmpegPath = candidate
			} else {
				candidateExe := candidate + ".exe"
				if _, err := exec.LookPath(candidateExe); err == nil {
					ffmpegPath = candidateExe
				} else {
					return nil, utils.ErrFFmpegNotFound
				}
			}
		}
	}

	// Validate ffprobe path
	if ffprobePath == "" {
		return nil, utils.ErrFFProbePathEmpty
	}

	// Check if it's a system command or full path
	if ffprobePath == "ffprobe" {
		if _, err := exec.LookPath("ffprobe"); err != nil {
			return nil, utils.ErrFFProbeUnavailable
		}
	} else {
		// Check if it's a full path to ffprobe
		if strings.HasSuffix(ffprobePath, "ffprobe") || strings.HasSuffix(ffprobePath, "ffprobe.exe") {
			if _, err := exec.LookPath(ffprobePath); err != nil {
				return nil, utils.ErrInvalidFFProbePath
			}
		} else {
			// Check if it's a directory containing ffprobe
			candidate := filepath.Join(ffprobePath, "ffprobe")
			if _, err := exec.LookPath(candidate); err == nil {
				ffprobePath = candidate
			} else {
				candidateExe := candidate + ".exe"
				if _, err := exec.LookPath(candidateExe); err == nil {
					ffprobePath = candidateExe
				} else {
					return nil, utils.ErrFFProbeNotFound
				}
			}
		}
	}

	return &FFmpeg{
		FFmpegPath:  ffmpegPath,
		FFProbePath: ffprobePath,
	}, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetFFmpegPath returns the resolved ffmpeg path
func (f *FFmpeg) GetFFmpegPath() string {
	return f.FFmpegPath
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetFFProbePath returns the resolved ffprobe path
func (f *FFmpeg) GetFFProbePath() string {
	return f.FFProbePath
}
