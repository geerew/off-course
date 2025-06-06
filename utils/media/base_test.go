package media

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/geerew/off-course/utils"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func ffprobeAvailable(t *testing.T) {
	t.Helper()
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe not installed; skipping test")
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestProbeVideo(t *testing.T) {
	ffprobeAvailable(t)

	t.Run("valid video", func(t *testing.T) {
		testVideo := filepath.Join("testdata", "sample.mp4")

		mp := MediaProbe{}
		info, err := mp.ProbeVideo(testVideo)
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, 1, info.Duration)
		require.Equal(t, 1280, info.Width)
		require.Equal(t, 720, info.Height)
		require.Equal(t, "h264", info.Codec)
		require.Equal(t, "720p", info.Resolution)
	})

	t.Run("invalid video", func(t *testing.T) {
		mp := MediaProbe{}
		_, err := mp.ProbeVideo("testdata/does_not_exist.mp4")
		require.Error(t, err)
		require.Contains(t, err.Error(), "error running ffprobe")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestResolveFFProbePath(t *testing.T) {
	t.Run("ffprobe valid", func(t *testing.T) {
		ffprobeAvailable(t)

		mp := MediaProbe{FFProbePath: ""}
		path, err := mp.resolveFFProbePath()
		require.NoError(t, err)
		require.Equal(t, "ffprobe", path)

		mp = MediaProbe{FFProbePath: "ffprobe"}
		path, err = mp.resolveFFProbePath()
		require.NoError(t, err)
		require.NotEmpty(t, "ffprobe", path)
	})

	t.Run("ffprobe not found", func(t *testing.T) {
		mp := MediaProbe{FFProbePath: "nonexistent"}
		_, err := mp.resolveFFProbePath()
		require.ErrorIs(t, err, utils.ErrFFProbeNotFound)
	})

	t.Run("ffprobe invalid path", func(t *testing.T) {
		mp := MediaProbe{FFProbePath: "nonexistent/ffprobe"}
		_, err := mp.resolveFFProbePath()
		require.ErrorIs(t, err, utils.ErrInvalidFFProbePath)
	})

}
