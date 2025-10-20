package hls

import (
	"time"

	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tracker manages stream cleanup
type Tracker struct {
	lastUsage     map[string]time.Time
	transcoder    *Transcoder
	deletedStream chan string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTracker creates a new tracker
func NewTracker(t *Transcoder) *Tracker {
	ret := &Tracker{
		lastUsage:     make(map[string]time.Time),
		deletedStream: make(chan string),
		transcoder:    t,
	}
	go ret.start()
	return ret
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// start begins the tracker's main loop
func (t *Tracker) start() {
	cleanupInterval := 30 * time.Minute
	cleanupTimer := time.After(cleanupInterval)

	for {
		select {
		case info, ok := <-t.transcoder.clientChan:
			if !ok {
				return
			}

			t.lastUsage[info.assetID] = time.Now()

		case <-cleanupTimer:
			cleanupTimer = time.After(cleanupInterval)

			for assetID, lastUsed := range t.lastUsage {
				if time.Since(lastUsed) > 2*time.Hour {
					utils.Infof("HLS: Cleaning up old stream for asset %s (last used %v ago)\n", assetID, time.Since(lastUsed))
					t.DestroyStreamIfOld(assetID)
					delete(t.lastUsage, assetID)
				}
			}
		case assetID := <-t.deletedStream:
			t.DestroyStreamIfOld(assetID)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DestroyStreamIfOld destroys a stream if it's old enough
func (t *Tracker) DestroyStreamIfOld(assetID string) {
	if time.Since(t.lastUsage[assetID]) < 4*time.Hour {
		return
	}
	stream, ok := t.transcoder.streams.GetAndRemove(assetID)
	if !ok {
		return
	}
	stream.Destroy()
}
