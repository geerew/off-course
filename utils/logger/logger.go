package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// LogLevel represents the log level
type LogLevel int

const (
	LevelDebug LogLevel = -1
	LevelInfo  LogLevel = 0
	LevelWarn  LogLevel = 2
	LevelError LogLevel = 1
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Config holds the logger configuration
type Config struct {
	// Level sets the minimum log level (Debug, Info, Error)
	Level LogLevel

	// ConsoleOutput enables pretty console output
	ConsoleOutput bool

	// DbWriter is an optional writer for database logging
	DbWriter io.Writer

	// AdditionalWriters for any custom outputs
	AdditionalWriters []io.Writer
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Logger wraps zerolog.Logger with component support
type Logger struct {
	zlog      zerolog.Logger
	component string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new Logger with the specified configuration
func New(config *Config) *Logger {
	if config == nil {
		config = &Config{
			Level:         LevelInfo,
			ConsoleOutput: true,
		}
	}

	// Collect all writers
	writers := []io.Writer{}

	// Add console output if enabled
	if config.ConsoleOutput {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
			NoColor:    false,
		}
		writers = append(writers, consoleWriter)
	}

	// Add database writer if provided
	if config.DbWriter != nil {
		writers = append(writers, config.DbWriter)
	}

	// Add any additional writers
	writers = append(writers, config.AdditionalWriters...)

	// Create multi-writer
	var output io.Writer
	if len(writers) == 0 {
		output = io.Discard
	} else if len(writers) == 1 {
		output = writers[0]
	} else {
		output = zerolog.MultiLevelWriter(writers...)
	}

	// Create zerolog logger
	zlog := zerolog.New(output).With().Timestamp().Logger()

	// Set log level
	switch config.Level {
	case LevelDebug:
		zlog = zlog.Level(zerolog.DebugLevel)
	case LevelInfo:
		zlog = zlog.Level(zerolog.InfoLevel)
	case LevelWarn:
		zlog = zlog.Level(zerolog.WarnLevel)
	case LevelError:
		zlog = zlog.Level(zerolog.ErrorLevel)
	default:
		zlog = zlog.Level(zerolog.InfoLevel)
	}

	return &Logger{
		zlog: zlog,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithApp creates a logger for the application
func (l *Logger) WithApp() *Logger {
	return l.withComponent("app")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithAPI creates a logger for the API component
func (l *Logger) WithAPI() *Logger {
	return l.withComponent("api")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithHLS creates a logger for the HLS transcoding component
func (l *Logger) WithHLS() *Logger {
	return l.withComponent("hls")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithCourseScan creates a logger for the course scanning component
func (l *Logger) WithCourseScan() *Logger {
	return l.withComponent("coursescan")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithCardCache creates a logger for the card cache component
func (l *Logger) WithCardCache() *Logger {
	return l.withComponent("cardcache")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithCron creates a logger for the cron jobs component
func (l *Logger) WithCron() *Logger {
	return l.withComponent("cron")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// withComponent creates a new Logger with the specified component (internal use)
func (l *Logger) withComponent(component string) *Logger {
	return &Logger{
		zlog:      l.zlog.With().Str("component", component).Logger(),
		component: component,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Debug returns a debug level event
func (l *Logger) Debug() *zerolog.Event {
	return l.zlog.Debug()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Info returns an info level event
func (l *Logger) Info() *zerolog.Event {
	return l.zlog.Info()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Error returns an error level event
func (l *Logger) Error() *zerolog.Event {
	return l.zlog.Error()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Warn returns a warn level event
func (l *Logger) Warn() *zerolog.Event {
	return l.zlog.Warn()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetZerolog returns the underlying zerolog.Logger for advanced usage
func (l *Logger) GetZerolog() zerolog.Logger {
	return l.zlog
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Component returns the current component name
func (l *Logger) Component() string {
	return l.component
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NilLogger creates a logger that discards all output (useful for tests)
func NilLogger() *Logger {
	zlog := zerolog.New(io.Discard)
	return &Logger{
		zlog: zlog,
	}
}
