package hls

import (
	"os/exec"
	"sync"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Flags represents stream type flags
type Flags int32

const (
	AudioF   Flags = 1 << 0
	VideoF   Flags = 1 << 1
	Transmux Flags = 1 << 3
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// StreamHandle defines the interface that video and audio streams must implement
type StreamHandle interface {
	getTranscodeArgs(segments string) []string
	getOutPath(encoderID int) string
	getFlags() Flags
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Segment represents a single HLS segment
type Segment struct {
	// channel open if the segment is not ready, closed if ready
	// You can wait for it to be ready (non-blocking if already ready) by doing:
	//  <-segments[i].channel
	channel chan struct{}
	encoder int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Head represents an encoding process
type Head struct {
	segment int32
	end     int32
	command *exec.Cmd
}

// DeletedHead is a marker for a head that has been killed
var DeletedHead = Head{
	segment: -1,
	end:     -1,
	command: nil,
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// VideoKey uniquely identifies a video stream by index and quality
type VideoKey struct {
	idx     uint32
	quality Quality
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ClientInfo tracks what a client is watching
type ClientInfo struct {
	client  string
	assetID string
	path    string
	video   *VideoKey
	audio   *uint32
	vhead   int32
	ahead   int32
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Keyframe holds keyframe timestamps for seeking
type Keyframe struct {
	Keyframes []float64
	IsDone    bool
	info      *KeyframeInfo
}

// KeyframeInfo holds synchronization primitives for keyframes
type KeyframeInfo struct {
	ready     sync.WaitGroup
	mutex     sync.RWMutex
	listeners []func(keyframes []float64)
}

// Get returns the keyframe at the given index
func (kf *Keyframe) Get(idx int32) float64 {
	kf.info.mutex.RLock()
	defer kf.info.mutex.RUnlock()
	return kf.Keyframes[idx]
}

// Slice returns a slice of keyframes from start to end
func (kf *Keyframe) Slice(start int32, end int32) []float64 {
	if end <= start {
		return []float64{}
	}
	kf.info.mutex.RLock()
	defer kf.info.mutex.RUnlock()

	ref := kf.Keyframes[start:end]
	if kf.IsDone {
		return ref
	}
	// make a copy since we will continue to mutate the array
	ret := make([]float64, end-start)
	copy(ret, ref)
	return ret
}

// Length returns the number of keyframes and whether extraction is complete
func (kf *Keyframe) Length() (int32, bool) {
	kf.info.mutex.RLock()
	defer kf.info.mutex.RUnlock()
	return int32(len(kf.Keyframes)), kf.IsDone
}

// add appends new keyframes and notifies listeners
func (kf *Keyframe) add(values []float64) {
	kf.info.mutex.Lock()
	defer kf.info.mutex.Unlock()
	kf.Keyframes = append(kf.Keyframes, values...)
	for _, listener := range kf.info.listeners {
		listener(kf.Keyframes)
	}
}

// AddListener registers a callback for keyframe updates
func (kf *Keyframe) AddListener(callback func(keyframes []float64)) {
	kf.info.mutex.Lock()
	defer kf.info.mutex.Unlock()
	kf.info.listeners = append(kf.info.listeners, callback)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewKeyframeFromSlice creates a Keyframe from a slice of timestamps
func NewKeyframeFromSlice(keyframes []float64, isDone bool) *Keyframe {
	kf := &Keyframe{
		Keyframes: keyframes,
		IsDone:    isDone,
		info:      &KeyframeInfo{},
	}
	kf.info.ready.Add(1)
	kf.info.ready.Done() // Mark as ready immediately since we have the data
	return kf
}
