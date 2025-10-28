package hls

import (
	"time"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tracker manages stream cleanup
type Tracker struct {
	lastUsage  map[string]time.Time
	transcoder *Transcoder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTracker creates a new tracker
func NewTracker(t *Transcoder) *Tracker {
	ret := &Tracker{
		lastUsage:  make(map[string]time.Time),
		transcoder: t,
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
		case assetID, ok := <-t.transcoder.assetChan:
			if !ok {
				return
			}

			// Update when an asset was last viewed
			t.lastUsage[assetID] = time.Now()

		case <-cleanupTimer:
			// Cleanup assets that haven't been viewed in 2 hours
			cleanupTimer = time.After(cleanupInterval)

			for assetID, lastUsed := range t.lastUsage {
				if time.Since(lastUsed) > 2*time.Hour {
					t.transcoder.config.Logger.Debug().
						Str("asset_id", assetID).
						Dur("last_used_ago", time.Since(lastUsed)).
						Msg("Cleaning up old stream")
					t.DestroyStreamIfOld(assetID)
					delete(t.lastUsage, assetID)
				}
			}
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
