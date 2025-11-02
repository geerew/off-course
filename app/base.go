package app

import (
	"context"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media"
	"github.com/geerew/off-course/utils/media/hls"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// App represents the application and contains all dependencies
type App struct {
	// Core dependencies
	Logger    *logger.Logger
	AppFs     *appfs.AppFs
	FFmpeg    *media.FFmpeg
	DbManager *database.DatabaseManager

	// Services
	CourseScan *coursescan.CourseScan
	Transcoder *hls.Transcoder

	// Configuration
	Config *Config
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Config holds application configuration
type Config struct {
	HttpAddr     string
	DataDir      string
	IsDev        bool
	EnableSignup bool
	IsDebug      bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new App instance with all dependencies initialized
func New(ctx context.Context, config *Config) (*App, error) {
	// Determine log level
	logLevel := logger.LevelInfo
	if config.IsDebug {
		logLevel = logger.LevelDebug
	}

	// Logger
	appLogger := logger.New(&logger.Config{
		Level:         logLevel,
		ConsoleOutput: true,
	})

	if appLogger == nil {
		return nil, &InitializationError{Message: "Failed to initialize logger"}
	}

	// AppFS (filesystem)
	appFs := appfs.New(afero.NewOsFs())

	// FFmpeg
	ffmpeg, err := media.NewFFmpeg()
	if err != nil {
		return nil, &InitializationError{Message: "Failed to initialize FFmpeg", Err: err}
	}

	// Database manager
	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: config.DataDir,
		AppFs:   appFs,
	})

	if err != nil {
		return nil, &InitializationError{Message: "Failed to create database manager", Err: err}
	}

	// Create app instance first (needed for service initialization)
	application := &App{
		Logger:    appLogger,
		AppFs:     appFs,
		FFmpeg:    ffmpeg,
		DbManager: dbManager,
		Config:    config,
	}

	// Course scanner
	application.CourseScan = coursescan.New(&coursescan.CourseScanConfig{
		Db:     application.DbManager.DataDb,
		AppFs:  application.AppFs,
		Logger: application.Logger.WithCourseScan(),
		FFmpeg: application.FFmpeg,
	})

	// HLS Transcoder
	transcoder, err := hls.NewTranscoder(&hls.TranscoderConfig{
		CachePath: application.Config.DataDir,
		HwAccel:   hls.DetectHardwareAccel(application.Logger),
		AppFs:     application.AppFs,
		Logger:    application.Logger.WithHLS(),
		Dao:       dao.New(application.DbManager.DataDb),
	})

	if err != nil {
		return nil, &InitializationError{Message: "Failed to create HLS transcoder", Err: err}
	}

	application.Transcoder = transcoder

	return application, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InitializationError represents an error during app initialization
type InitializationError struct {
	Message string
	Err     error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Error returns the error message
func (e *InitializationError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Unwrap returns the wrapped error
func (e *InitializationError) Unwrap() error {
	return e.Err
}
