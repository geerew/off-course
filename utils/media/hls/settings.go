package hls

import (
	"path/filepath"

	"github.com/geerew/off-course/utils/appfs"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SettingsT holds global HLS transcoding settings
type SettingsT struct {
	CachePath string
	HwAccel   HwAccelT
	AppFs     *appfs.AppFs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Settings is the global settings instance
var Settings SettingsT

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InitSettings initializes the HLS settings with the given data directory and appfs
func InitSettings(dataDir string, appFs *appfs.AppFs) {
	// Use relative paths for in-memory filesystems
	var cachePath string
	if _, ok := appFs.Fs.(*afero.MemMapFs); ok {
		// In-memory filesystem
		cachePath = filepath.Join(dataDir, "hls")
	} else {
		// Real filesystem
		absDataDir, err := filepath.Abs(dataDir)
		if err != nil {
			absDataDir = dataDir
		}
		cachePath = filepath.Join(absDataDir, "hls")
	}

	Settings = SettingsT{
		CachePath: cachePath,
		HwAccel:   DetectHardwareAccel(),
		AppFs:     appFs,
	}

	appFs.Fs.MkdirAll(Settings.CachePath, 0o755)
}
