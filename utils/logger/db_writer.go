package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DbWriter implements io.Writer for writing logs to the database
type DbWriter struct {
	createLogFn func(context.Context, *models.Log) error
	ctx         context.Context
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewDbWriter creates a new DbWriter
func NewDbWriter(createLogFn func(context.Context, *models.Log) error) *DbWriter {
	return &DbWriter{
		createLogFn: createLogFn,
		ctx:         context.Background(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Write implements io.Writer interface
// It parses zerolog JSON output and writes to the database
func (w *DbWriter) Write(p []byte) (n int, err error) {
	// Parse the zerolog JSON output
	var logEntry struct {
		Level     string                 `json:"level"`
		Time      time.Time              `json:"time"`
		Message   string                 `json:"message"`
		Component string                 `json:"component,omitempty"`
		Data      map[string]interface{} `json:"-"`
	}

	// Parse JSON, ignoring unknown fields
	if err := json.Unmarshal(p, &logEntry); err != nil {
		// If parsing fails, create a basic log entry
		logEntry = struct {
			Level     string                 `json:"level"`
			Time      time.Time              `json:"time"`
			Message   string                 `json:"message"`
			Component string                 `json:"component,omitempty"`
			Data      map[string]interface{} `json:"-"`
		}{
			Level:   "info",
			Time:    time.Now(),
			Message: string(p),
		}
	}

	// Extract additional fields from the JSON
	var rawData map[string]interface{}
	if err := json.Unmarshal(p, &rawData); err == nil {
		// Remove standard fields to get custom data
		delete(rawData, "level")
		delete(rawData, "time")
		delete(rawData, "message")
		delete(rawData, "component")
		logEntry.Data = rawData
	}

	// Convert zerolog level to our level
	var level int
	switch logEntry.Level {
	case "debug":
		level = int(LevelDebug)
	case "info":
		level = int(LevelInfo)
	case "warn":
		level = int(LevelInfo) // Map warn to info in our 3-level system
	case "error":
		level = int(LevelError)
	default:
		level = int(LevelInfo)
	}

	// Create log model
	log := &models.Log{
		Level:   level,
		Message: logEntry.Message,
		Data:    types.JsonMap(logEntry.Data),
	}

	// Add component to data if present
	if logEntry.Component != "" {
		if log.Data == nil {
			log.Data = make(types.JsonMap)
		}
		log.Data["component"] = logEntry.Component
	}

	// Write to database
	if w.createLogFn != nil {
		if err := w.createLogFn(w.ctx, log); err != nil {
			// If database write fails, we can't return an error from Write()
			// as it would break the logger. Log to stderr instead.
			fmt.Fprintf(io.Discard, "Failed to write log to database: %v\n", err)
		}
	}

	return len(p), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DbWriterFunc is a function type for creating log entries in the database
type DbWriterFunc func(context.Context, *models.Log) error

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateDbWriter creates a DbWriter using the provided DAO
func CreateDbWriter(dao interface {
	CreateLog(context.Context, *models.Log) error
}) *DbWriter {
	return NewDbWriter(dao.CreateLog)
}
