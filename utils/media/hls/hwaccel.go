package hls

import (
	"os"

	"github.com/geerew/off-course/utils"
)

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

// DetectHardwareAccel detects and configures hardware acceleration
func DetectHardwareAccel() HwAccelT {
	name := utils.GetEnvOr("OC_HWACCEL", "disabled")

	utils.Infof("HLS: Using hardware acceleration: %s\n", name)

	// superfast/ultrafast create extremely big files, so we prefer to ignore them. Fast
	// is available on all modes so we use that by default (except for vaapi, which does not
	// support the flag)
	preset := utils.GetEnvOr("OC_PRESET", "fast")

	switch name {
	case "disabled", "cpu":
		return HwAccelT{
			Name:        "disabled",
			DecodeFlags: []string{},
			EncodeFlags: []string{
				"-c:v", "libx264",
				"-preset", preset,
				// sc_threshold is a scene detection mechanism used to create a keyframe when
				// the scene changes. It inserts keyframes where we don't want to and breaks
				// force_key_frames. Disable it to prevents whole scenes from being removed due
				// to the -f segment failing to find the corresponding keyframe
				"-sc_threshold", "0",
				// Force 8bits output, keeping it the same as the source
				"-pix_fmt", "yuv420p",
			},
			ScaleFilter: "scale=%d:%d",
		}
	case "vaapi":
		return HwAccelT{
			Name: name,
			DecodeFlags: []string{
				"-hwaccel", "vaapi",
				"-hwaccel_device", utils.GetEnvOr("OC_VAAPI_RENDERER", "/dev/dri/renderD128"),
				"-hwaccel_output_format", "vaapi",
			},
			EncodeFlags: []string{
				// preset or scenecut flags not supported by vaapi
				"-c:v", "h264_vaapi",
			},
			// If the hardware decoding does not work and falls back to soft decoding, instruct ffmpeg
			// to upload the frames back to gpu space (after converting them)
			//
			// Also forces the format to be nv12 since 10bits is not supported via hardware acceleration
			ScaleFilter:    "format=nv12|vaapi,hwupload,scale_vaapi=%d:%d:format=nv12",
			NoResizeFilter: "format=nv12|vaapi,hwupload,scale_vaapi=format=nv12",
		}
	case "qsv", "intel":
		return HwAccelT{
			Name: name,
			DecodeFlags: []string{
				"-hwaccel", "qsv",
				"-qsv_device", utils.GetEnvOr("GOCODER_QSV_RENDERER", "/dev/dri/renderD128"),
				"-hwaccel_output_format", "qsv",
			},
			EncodeFlags: []string{
				"-c:v", "h264_qsv",
				"-preset", preset,
			},
			ScaleFilter:    "format=nv12|qsv,hwupload,scale_qsv=%d:%d:format=nv12",
			NoResizeFilter: "format=nv12|qsv,hwupload,scale_qsv=format=nv12",
		}
	case "nvidia":
		return HwAccelT{
			Name: "nvidia",
			DecodeFlags: []string{
				"-hwaccel", "cuda",
				"-hwaccel_output_format", "cuda",
			},
			EncodeFlags: []string{
				"-c:v", "h264_nvenc",
				"-preset", preset,
				"-no-scenecut", "1",
			},
			ScaleFilter:    "format=nv12|cuda,hwupload,scale_cuda=%d:%d:format=nv12",
			NoResizeFilter: "format=nv12|cuda,hwupload,scale_cuda=format=nv12",
		}
	default:
		utils.Errf("HLS: No hardware accelerator named: %s\n", name)
		os.Exit(2)
		panic("unreachable")
	}
}
