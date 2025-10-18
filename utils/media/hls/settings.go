package hls

import (
	"os"
	"path/filepath"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetEnvOr returns the environment variable value or a default
func GetEnvOr(env string, def string) string {
	out := os.Getenv(env)
	if out == "" {
		return def
	}
	return out
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SettingsT holds global HLS transcoding settings
type SettingsT struct {
	CachePath string
	HwAccel   HwAccelT
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// HwAccelT defines hardware acceleration configuration
type HwAccelT struct {
	Name           string
	DecodeFlags    []string
	EncodeFlags    []string
	NoResizeFilter string
	ScaleFilter    string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Settings is the global settings instance
var Settings SettingsT

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InitSettings initializes the HLS settings with the given data directory
func InitSettings(dataDir string) {
	// Ensure dataDir is absolute
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		// Fallback to original if absolute path fails
		absDataDir = dataDir
	}

	Settings = SettingsT{
		CachePath: filepath.Join(absDataDir, "hls"),
		HwAccel:   DetectHardwareAccel(),
	}

	// Ensure cache directory exists
	os.MkdirAll(Settings.CachePath, 0o755)
}
