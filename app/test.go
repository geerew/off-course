package app

import (
	"sync"
	"testing"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media"
	"github.com/geerew/off-course/utils/media/hls"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	// Cache FFmpeg instance for tests
	cachedFFmpeg *media.FFmpeg
	ffmpegOnce   sync.Once
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getCachedFFmpeg returns a cached FFmpeg instance for tests
func getCachedFFmpeg(t *testing.T) *media.FFmpeg {
	ffmpegOnce.Do(func() {
		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			// Skip if FFmpeg unavailable
			t.Skip("FFmpeg not available for testing")
		}
		cachedFFmpeg = ffmpeg
	})
	return cachedFFmpeg
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestApp creates a test app instance with test-friendly defaults
func NewTestApp(t *testing.T) *App {
	t.Helper()

	// Create a test logger
	testLogger := logger.New(&logger.Config{
		Level:         logger.LevelInfo,
		ConsoleOutput: false, // Disable console output for tests
	})

	// Create in-memory filesystem
	appFs := appfs.New(afero.NewMemMapFs())

	// Create database manager with test config
	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: "./oc_data",
		AppFs:   appFs,
		Testing: true,
	})
	require.NoError(t, err)
	require.NotNil(t, dbManager)

	// Get FFmpeg instance
	ffmpeg := getCachedFFmpeg(t)

	// Create app instance with test-friendly defaults
	application := &App{
		Logger:    testLogger,
		AppFs:     appFs,
		FFmpeg:    ffmpeg,
		DbManager: dbManager,
		Config: &Config{
			HttpAddr:     "127.0.0.1:9081",
			DataDir:      "./oc_data",
			IsDev:        false,
			EnableSignup: true,
			IsDebug:      false,
		},
	}

	// Initialize CourseScan
	application.CourseScan = coursescan.New(&coursescan.CourseScanConfig{
		Db:     application.DbManager.DataDb,
		AppFs:  application.AppFs,
		Logger: application.Logger.WithCourseScan(),
		FFmpeg: application.FFmpeg,
	})

	// Initialize Transcoder
	transcoder, err := hls.NewTranscoder(&hls.TranscoderConfig{
		CachePath: application.Config.DataDir,
		HwAccel:   hls.DetectHardwareAccel(application.Logger),
		AppFs:     application.AppFs,
		Logger:    application.Logger.WithHLS(),
		Dao:       dao.New(application.DbManager.DataDb),
	})
	require.NoError(t, err)
	application.Transcoder = transcoder

	return application
}
