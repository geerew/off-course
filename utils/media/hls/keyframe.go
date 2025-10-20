package hls

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewKeyframeFromSlice creates a Keyframe from a slice of timestamps
// Keyframes are always complete since they're extracted during course scanning
func NewKeyframeFromSlice(keyframes []float64) *Keyframe {
	return &Keyframe{
		Keyframes: keyframes,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Keyframe holds keyframe timestamps for seeking
type Keyframe struct {
	Keyframes []float64
}

// Get returns the keyframe at the given index
func (kf *Keyframe) Get(idx int32) float64 {
	return kf.Keyframes[idx]
}

// Slice returns a slice of keyframes from start to end
func (kf *Keyframe) Slice(start int32, end int32) []float64 {
	if end <= start {
		return []float64{}
	}
	return kf.Keyframes[start:end]
}

// Length returns the number of keyframes
func (kf *Keyframe) Length() int32 {
	return int32(len(kf.Keyframes))
}
