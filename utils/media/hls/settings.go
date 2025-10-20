package hls

import (
	"os"
	"path/filepath"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SettingsT holds global HLS transcoding settings
type SettingsT struct {
	CachePath string
	HwAccel   HwAccelT
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Settings is the global settings instance
var Settings SettingsT

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InitSettings initializes the HLS settings with the given data directory
func InitSettings(dataDir string) {
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		absDataDir = dataDir
	}

	Settings = SettingsT{
		CachePath: filepath.Join(absDataDir, "hls"),
		HwAccel:   DetectHardwareAccel(),
	}

	os.MkdirAll(Settings.CachePath, 0o755)
}
