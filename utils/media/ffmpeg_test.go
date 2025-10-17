package media

import (
	"os/exec"
	"testing"

	"github.com/geerew/off-course/utils"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func ffmpegAvailable(t *testing.T) {
	t.Helper()
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed; skipping test")
	}
	_, err = exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe not installed; skipping test")
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestNewFFmpeg(t *testing.T) {
	ffmpegAvailable(t)

	t.Run("success", func(t *testing.T) {
		ffmpeg, err := NewFFmpeg()
		require.NoError(t, err)
		require.NotNil(t, ffmpeg)
		require.NotEmpty(t, ffmpeg.GetFFmpegPath())
		require.NotEmpty(t, ffmpeg.GetFFProbePath())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestNewFFmpegWithPaths(t *testing.T) {
	ffmpegAvailable(t)

	t.Run("success with system paths", func(t *testing.T) {
		ffmpeg, err := NewFFmpegWithPaths("ffmpeg", "ffprobe")
		require.NoError(t, err)
		require.NotNil(t, ffmpeg)
		require.Equal(t, "ffmpeg", ffmpeg.GetFFmpegPath())
		require.Equal(t, "ffprobe", ffmpeg.GetFFProbePath())
	})

	t.Run("error with empty ffmpeg path", func(t *testing.T) {
		_, err := NewFFmpegWithPaths("", "ffprobe")
		require.Error(t, err)
		require.Equal(t, utils.ErrFFmpegPathEmpty, err)
	})

	t.Run("error with empty ffprobe path", func(t *testing.T) {
		_, err := NewFFmpegWithPaths("ffmpeg", "")
		require.Error(t, err)
		require.Equal(t, utils.ErrFFProbePathEmpty, err)
	})

	t.Run("error with invalid ffmpeg path", func(t *testing.T) {
		_, err := NewFFmpegWithPaths("/nonexistent/ffmpeg", "ffprobe")
		require.Error(t, err)
		require.Equal(t, utils.ErrInvalidFFmpegPath, err)
	})

	t.Run("error with invalid ffprobe path", func(t *testing.T) {
		_, err := NewFFmpegWithPaths("ffmpeg", "/nonexistent/ffprobe")
		require.Error(t, err)
		require.Equal(t, utils.ErrInvalidFFProbePath, err)
	})
}
