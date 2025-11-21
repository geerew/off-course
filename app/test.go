package app

import (
	"sync"
	"testing"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/cardcache"
	"github.com/geerew/off-course/utils/coursemetadata"
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
	testLogger := logger.NilLogger()

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
	app := &App{
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

	// Initialize Transcoder
	transcoder, err := hls.NewTranscoder(&hls.TranscoderConfig{
		CachePath: app.Config.DataDir,
		HwAccel:   hls.DetectHardwareAccel(app.Logger),
		AppFs:     app.AppFs,
		Logger:    app.Logger.WithHLS(),
		Dao:       dao.New(app.DbManager.DataDb),
	})
	require.NoError(t, err)
	app.Transcoder = transcoder

	// Initialize CardCache
	cardCache, err := cardcache.NewCardCache(&cardcache.CardCacheConfig{
		CachePath: app.Config.DataDir,
		AppFs:     app.AppFs,
		Logger:    app.Logger.WithCardCache(),
		FFmpeg:    app.FFmpeg,
	})
	require.NoError(t, err)
	app.CardCache = cardCache

	// Ensure fallback card exists (required for tests)
	fallbackPath := cardCache.GetFallbackPath()
	err = cardCache.EnsureFallbackCard(fallbackPath)
	require.NoError(t, err)

	// Initialize CourseScan
	app.CourseScan = coursescan.New(&coursescan.CourseScanConfig{
		Db:        app.DbManager.DataDb,
		AppFs:     app.AppFs,
		Logger:    app.Logger.WithCourseScan(),
		FFmpeg:    app.FFmpeg,
		CardCache: cardCache,
	})

	// Initialize MetadataWriter
	app.MetadataWriter = coursemetadata.NewMetadataWriter(app.AppFs.Fs, app.Logger)

	return app
}
