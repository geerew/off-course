package hls

import (
	"errors"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Quality represents a video quality level
type Quality string

const (
	P240     Quality = "240p"
	P360     Quality = "360p"
	P480     Quality = "480p"
	P720     Quality = "720p"
	P1080    Quality = "1080p"
	P1440    Quality = "1440p"
	NoResize Quality = "transcode"
	Original Quality = "original"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Purposefully removing Original from this list (since it requires special treatment anyway)
var Qualities = []Quality{P240, P360, P480, P720, P1080, P1440}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QualityFromString parses a quality string
func QualityFromString(str string) (Quality, error) {
	if str == string(Original) {
		return Original, nil
	}
	if str == string(NoResize) {
		return NoResize, nil
	}

	for _, quality := range Qualities {
		if string(quality) == str {
			return quality, nil
		}
	}
	return Original, errors.New("invalid quality")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AverageBitrate returns the average bitrate for a quality level
func (v Quality) AverageBitrate() uint32 {
	switch v {
	case P240:
		return 400_000
	case P360:
		return 800_000
	case P480:
		return 1_200_000
	case P720:
		return 2_400_000
	case P1080:
		return 4_800_000
	case P1440:
		return 9_600_000
	case Original:
		panic("Original quality must be handled specially")
	}
	panic("Invalid quality value")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MaxBitrate returns the maximum bitrate for a quality level
func (v Quality) MaxBitrate() uint32 {
	switch v {
	case P240:
		return 700_000
	case P360:
		return 1_400_000
	case P480:
		return 2_100_000
	case P720:
		return 4_000_000
	case P1080:
		return 8_000_000
	case P1440:
		return 12_000_000
	case Original:
		panic("Original quality must be handled specially")
	}
	panic("Invalid quality value")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Height returns the height in pixels for a quality level
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
	case Original:
		panic("Original quality must be handled specially")
	}
	panic("Invalid quality value")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetQualityForVideo determines the appropriate quality for a video based on its height and bitrate
func GetQualityForVideo(height uint32, bitrate uint32) Quality {
	for _, quality := range Qualities {
		if quality.Height() >= height || quality.AverageBitrate() >= bitrate {
			return quality
		}
	}

	return P240
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetQualitiesHighestToLowest returns qualities ordered from highest to lowest (excluding original)
func GetQualitiesHighestToLowest(qualities []Quality) []Quality {
	// Filter out Original and return only transcoded qualities
	var transcodedQualities []Quality
	for _, quality := range qualities {
		if quality != Original {
			transcodedQualities = append(transcodedQualities, quality)
		}
	}

	return transcodedQualities
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetQualitiesLowestToHighest returns qualities ordered from lowest to highest (excluding original)
func GetQualitiesLowestToHighest(qualities []Quality) []Quality {
	highestToLowest := GetQualitiesHighestToLowest(qualities)

	// Reverse the slice
	var lowestToHighest []Quality
	for i := len(highestToLowest) - 1; i >= 0; i-- {
		lowestToHighest = append(lowestToHighest, highestToLowest[i])
	}

	return lowestToHighest
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetHighestTranscodedQuality returns the highest available transcoded quality (not original)
func GetHighestTranscodedQuality(qualities []Quality) Quality {
	transcodedQualities := GetQualitiesHighestToLowest(qualities)

	if len(transcodedQualities) > 0 {
		return transcodedQualities[0] // First element is highest
	}

	// Fallback to 720p if no transcoded qualities found
	return P720
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetLowestTranscodedQuality returns the lowest available transcoded quality (not original)
func GetLowestTranscodedQuality(qualities []Quality) Quality {
	transcodedQualities := GetQualitiesLowestToHighest(qualities)

	if len(transcodedQualities) > 0 {
		return transcodedQualities[0] // First element is lowest
	}

	// Fallback to 240p if no transcoded qualities found
	return P240
}
