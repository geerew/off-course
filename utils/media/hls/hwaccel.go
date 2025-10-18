package hls

import (
	"fmt"
	"os/exec"
	"strings"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// HwAccelType represents the type of hardware acceleration
type HwAccelType string

const (
	HwAccelNone  HwAccelType = "none"
	HwAccelVAAPI HwAccelType = "vaapi"
	HwAccelQSV   HwAccelType = "qsv"
	HwAccelNVENC HwAccelType = "nvenc"
	HwAccelVTB   HwAccelType = "videotoolbox"
	HwAccelAuto  HwAccelType = "auto"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// HwAccelConfig holds hardware acceleration configuration
type HwAccelConfig struct {
	Type           HwAccelType
	Device         string // GPU device path (for VAAPI)
	Available      bool
	DecodeFlags    []string
	EncodeFlags    []string
	ScaleFilter    string // Hardware-accelerated scaling filter
	NoResizeFilter string // Filter for when no scaling is needed
	Preset         string // Encoding preset (fast, medium, slow, etc.)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DetectHardwareAcceleration detects available hardware acceleration
func DetectHardwareAcceleration(ffmpegPath string) *HwAccelConfig {
	// Try to detect available hardware acceleration
	config := &HwAccelConfig{
		Type:      HwAccelNone,
		Available: false,
		Preset:    "fast", // Default preset
	}

	// Check if ffmpeg supports hardware acceleration
	cmd := exec.Command(ffmpegPath, "-hide_banner", "-hwaccels")
	output, err := cmd.Output()
	if err != nil {
		return config
	}

	hwaccels := strings.ToLower(string(output))

	// Detect based on available hardware acceleration
	if strings.Contains(hwaccels, "vaapi") {
		config.Type = HwAccelVAAPI
		config.Available = true
		config.Device = "/dev/dri/renderD128"
		config.DecodeFlags = []string{
			"-hwaccel", "vaapi",
			"-hwaccel_device", config.Device,
			"-hwaccel_output_format", "vaapi",
		}
		config.EncodeFlags = []string{
			"-c:v", "h264_vaapi",
		}
		// VAAPI scaling filter with format conversion
		config.ScaleFilter = "format=nv12|vaapi,hwupload,scale_vaapi=%d:%d:format=nv12"
		config.NoResizeFilter = "format=nv12|vaapi,hwupload,scale_vaapi=format=nv12"
	} else if strings.Contains(hwaccels, "qsv") {
		config.Type = HwAccelQSV
		config.Available = true
		config.Device = "/dev/dri/renderD128"
		config.DecodeFlags = []string{
			"-hwaccel", "qsv",
			"-qsv_device", config.Device,
			"-hwaccel_output_format", "qsv",
		}
		config.EncodeFlags = []string{
			"-c:v", "h264_qsv",
			"-preset", config.Preset,
		}
		// QSV scaling filter
		config.ScaleFilter = "format=nv12|qsv,hwupload,scale_qsv=%d:%d:format=nv12"
		config.NoResizeFilter = "format=nv12|qsv,hwupload,scale_qsv=format=nv12"
	} else if strings.Contains(hwaccels, "nvenc") {
		config.Type = HwAccelNVENC
		config.Available = true
		config.DecodeFlags = []string{
			"-hwaccel", "cuda",
			"-hwaccel_output_format", "cuda",
		}
		config.EncodeFlags = []string{
			"-c:v", "h264_nvenc",
			"-preset", config.Preset,
			"-no-scenecut", "1", // Equivalent of -sc_threshold for NVENC
		}
		// CUDA scaling filter
		config.ScaleFilter = "format=nv12|cuda,hwupload,scale_cuda=%d:%d:format=nv12"
		config.NoResizeFilter = "format=nv12|cuda,hwupload,scale_cuda=format=nv12"
	} else if strings.Contains(hwaccels, "videotoolbox") {
		config.Type = HwAccelVTB
		config.Available = true
		config.DecodeFlags = []string{
			"-hwaccel", "videotoolbox",
		}
		config.EncodeFlags = []string{
			"-c:v", "h264_videotoolbox",
			"-preset", config.Preset,
		}
		// VideoToolbox doesn't need special scaling filters
		config.ScaleFilter = "scale=%d:%d"
		config.NoResizeFilter = ""
	}

	return config
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetHwAccelConfig returns hardware acceleration configuration
func GetHwAccelConfig(ffmpegPath string, preferredType HwAccelType) *HwAccelConfig {
	detected := DetectHardwareAcceleration(ffmpegPath)

	// If no preference or auto, use detected
	if preferredType == HwAccelAuto || preferredType == HwAccelNone {
		return detected
	}

	// Check if preferred type is available
	if preferredType == detected.Type {
		return detected
	}

	// Fall back to detected or none
	return detected
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetTranscodeArgs returns FFmpeg arguments for transcoding with hardware acceleration
func (config *HwAccelConfig) GetTranscodeArgs() []string {
	if !config.Available {
		// Software fallback - matches Kyoo's approach
		return []string{
			"-c:v", "libx264",
			"-preset", config.Preset,
			// Disable scene detection to prevent unwanted keyframes
			"-sc_threshold", "0",
			// Force 8-bit output for compatibility
			"-pix_fmt", "yuv420p",
		}
	}

	// Combine decode and encode flags
	args := make([]string, 0, len(config.DecodeFlags)+len(config.EncodeFlags))
	args = append(args, config.DecodeFlags...)
	args = append(args, config.EncodeFlags...)

	return args
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetSegmentArgs returns FFmpeg arguments for segment generation
func (config *HwAccelConfig) GetSegmentArgs() []string {
	args := []string{
		"-f", "segment",
		"-segment_time", "4",
		"-segment_list_flags", "+live",
		"-segment_format", "mpegts",
		"-segment_list_type", "m3u8",
	}

	// Add hardware acceleration if available
	if config.Available {
		args = append(config.DecodeFlags, args...)
		args = append(args, config.EncodeFlags...)
	} else {
		// Software fallback
		args = append(args, "-c:v", "libx264", "-preset", "fast", "-crf", "23")
	}

	return args
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsHardwareAccelerated returns true if hardware acceleration is available
func (config *HwAccelConfig) IsHardwareAccelerated() bool {
	return config.Available && config.Type != HwAccelNone
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScaleFilter returns the hardware-accelerated scaling filter for the given dimensions
func (config *HwAccelConfig) GetScaleFilter(width, height int) string {
	if !config.Available || config.ScaleFilter == "" {
		// Software fallback
		return fmt.Sprintf("scale=%d:%d", width, height)
	}
	return fmt.Sprintf(config.ScaleFilter, width, height)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetNoResizeFilter returns the filter for when no scaling is needed
func (config *HwAccelConfig) GetNoResizeFilter() string {
	if !config.Available || config.NoResizeFilter == "" {
		return ""
	}
	return config.NoResizeFilter
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String returns the string representation of the hardware acceleration type
func (config *HwAccelConfig) String() string {
	if !config.Available {
		return "none"
	}
	return string(config.Type)
}
