package hls

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Quality represents a video quality level
type Quality string

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	// Video quality levels (starting with original only)
	Original Quality = "original"

	// Future quality levels (not implemented yet)
	P240  Quality = "240p"
	P360  Quality = "360p"
	P480  Quality = "480p"
	P720  Quality = "720p"
	P1080 Quality = "1080p"
	P1440 Quality = "1440p"
	P4k   Quality = "4k"
	P8k   Quality = "8k"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Qualities is the list of available qualities (starting with original only)
var Qualities = []Quality{Original}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsOriginal checks if this is the original quality
func (q Quality) IsOriginal() bool {
	return q == Original
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Height returns the height in pixels for this quality
func (q Quality) Height() uint32 {
	switch q {
	case P240:
		return 240
	case P360:
		return 360
	case P480:
		return 480
	case P720:
		return 720
	case P1080:
		return 1080
	case P1440:
		return 1440
	case P4k:
		return 2160
	case P8k:
		return 4320
	case Original:
		return 0 // Original doesn't have a fixed height
	default:
		return 0
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AverageBitrate returns the average bitrate for this quality in bps
func (q Quality) AverageBitrate() uint32 {
	switch q {
	case P240:
		return 400_000 // 400 kbps
	case P360:
		return 800_000 // 800 kbps
	case P480:
		return 1_200_000 // 1.2 Mbps
	case P720:
		return 2_500_000 // 2.5 Mbps
	case P1080:
		return 5_000_000 // 5 Mbps
	case P1440:
		return 8_000_000 // 8 Mbps
	case P4k:
		return 15_000_000 // 15 Mbps
	case P8k:
		return 30_000_000 // 30 Mbps
	case Original:
		return 0 // Original uses source bitrate
	default:
		return 0
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MaxBitrate returns the maximum bitrate for this quality in bps
func (q Quality) MaxBitrate() uint32 {
	switch q {
	case P240:
		return 600_000 // 600 kbps
	case P360:
		return 1_200_000 // 1.2 Mbps
	case P480:
		return 1_800_000 // 1.8 Mbps
	case P720:
		return 3_750_000 // 3.75 Mbps
	case P1080:
		return 7_500_000 // 7.5 Mbps
	case P1440:
		return 12_000_000 // 12 Mbps
	case P4k:
		return 22_500_000 // 22.5 Mbps
	case P8k:
		return 45_000_000 // 45 Mbps
	case Original:
		return 0 // Original uses source bitrate
	default:
		return 0
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String returns the string representation of the quality
func (q Quality) String() string {
	return string(q)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsValid checks if this is a valid quality level
func (q Quality) IsValid() bool {
	for _, valid := range Qualities {
		if q == valid {
			return true
		}
	}
	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetQualityFromHeight returns the appropriate quality for a given height
func GetQualityFromHeight(height uint32) Quality {
	switch {
	case height >= 4320:
		return P8k
	case height >= 2160:
		return P4k
	case height >= 1440:
		return P1440
	case height >= 1080:
		return P1080
	case height >= 720:
		return P720
	case height >= 480:
		return P480
	case height >= 360:
		return P360
	case height >= 240:
		return P240
	default:
		return Original
	}
}
