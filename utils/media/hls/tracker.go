package hls

import (
	"time"

	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tracker manages stream cleanup (simplified - no client tracking)
type Tracker struct {
	// key: asset_id
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// start begins the tracker's main loop
func (t *Tracker) start() {
	cleanup_interval := 30 * time.Minute // Check every 30 minutes
	cleanup_timer := time.After(cleanup_interval)

	for {
		select {
		case info, ok := <-t.transcoder.clientChan:
			if !ok {
				return
			}
			// Just update the last usage time, no client tracking
			t.lastUsage[info.assetID] = time.Now()

		case <-cleanup_timer:
			cleanup_timer = time.After(cleanup_interval)
			// Simple time-based cleanup - remove streams older than 2 hours
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

// KillStreamIfDead kills a stream (simplified - no client checking)
func (t *Tracker) KillStreamIfDead(assetID string, path string) bool {
	utils.Infof("HLS: Killing stream for %s\n", path)

	stream, ok := t.transcoder.streams.Get(assetID)
	if !ok {
		return false
	}
	stream.Kill()
	go func() {
		time.Sleep(4 * time.Hour)
		t.deletedStream <- assetID
	}()
	return true
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// KillAudioIfDead kills an audio stream (simplified - no client checking)
func (t *Tracker) KillAudioIfDead(assetID string, path string, audio uint32) bool {
	utils.Infof("HLS: Killing audio stream %d for %s\n", audio, path)

	stream, ok := t.transcoder.streams.Get(assetID)
	if !ok {
		return false
	}
	astream, aok := stream.audios.Get(audio)
	if !aok {
		return false
	}
	astream.Kill()
	return true
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// KillVideoIfDead kills a video stream (simplified - no client checking)
func (t *Tracker) KillVideoIfDead(assetID string, path string, video VideoKey) bool {
	utils.Infof("HLS: Killing video stream %d quality %s for %s\n", video.idx, video.quality, path)

	stream, ok := t.transcoder.streams.Get(assetID)
	if !ok {
		return false
	}
	vstream, vok := stream.videos.Get(video)
	if !vok {
		return false
	}
	vstream.Kill()
	return true
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// KillOrphanedHeads kills orphaned encoding heads (simplified)
func (t *Tracker) KillOrphanedHeads(assetID string, video *VideoKey, audio *uint32) {
	stream, ok := t.transcoder.streams.Get(assetID)
	if !ok {
		return
	}

	if video != nil {
		vstream, vok := stream.videos.Get(*video)
		if vok {
			t.killOrphanedeheads(&vstream.Stream, true)
		}
	}
	if audio != nil {
		astream, aok := stream.audios.Get(*audio)
		if aok {
			t.killOrphanedeheads(&astream.Stream, false)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// killOrphanedeheads kills orphaned encoding heads for a stream (simplified)
func (t *Tracker) killOrphanedeheads(stream *Stream, is_video bool) {
	stream.lock.Lock()
	defer stream.lock.Unlock()

	for encoder_id, head := range stream.heads {
		if head == DeletedHead {
			continue
		}
		// Simplified: just kill heads that are very far ahead (no client checking)
		if head.segment > 100 { // Kill heads that are more than 100 segments ahead
			utils.Infof("HLS: Killing orphaned head %s %d (segment %d)\n", stream.file.Info.Path, encoder_id, head.segment)
			stream.KillHead(encoder_id)
		}
	}
}
