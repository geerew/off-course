package media

import (
	"os/exec"
	"testing"

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
