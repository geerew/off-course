package hls

import (
	"context"
	"os"
	"path/filepath"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/utils"
)

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

// Transcoder manages all HLS transcoding operations for a given asset
type Transcoder struct {
	streams    utils.CMap[string, *StreamWrapper]
	clientChan chan ClientInfo
	tracker    *Tracker
	dao        *dao.DAO
	assetID    string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTranscoder creates a new Transcoder and prepares the cache directory
func NewTranscoder(dao *dao.DAO) (*Transcoder, error) {
	out := Settings.CachePath
	os.MkdirAll(out, 0o755)
	dir, err := os.ReadDir(out)
	if err != nil {
		return nil, err
	}

	// Clean up old cache directories
	for _, d := range dir {
		err = os.RemoveAll(filepath.Join(out, d.Name()))
		if err != nil {
			return nil, err
		}
	}

	ret := &Transcoder{
		streams:    utils.NewCMap[string, *StreamWrapper](),
		clientChan: make(chan ClientInfo, 10),
		dao:        dao,
	}

	ret.tracker = NewTracker(ret)
	return ret, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getStreamWrapper returns an existing StreamWrapper for the asset or creates one
// Blocks until the StreamWrapper is ready or returns an error
func (t *Transcoder) getStreamWrapper(ctx context.Context, path string, assetID string) (*StreamWrapper, error) {
	sw, _ := t.streams.GetOrCreate(assetID, func() *StreamWrapper {
		t.assetID = assetID
		return t.newStreamWrapper(ctx, path, assetID)
	})

	if sw.err != nil {
		t.streams.Remove(assetID)
		return nil, sw.err
	}

	return sw, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetMaster returns the master HLS playlist for an asset
func (t *Transcoder) GetMaster(ctx context.Context, path string, client string, assetID string) (string, error) {
	sw, err := t.getStreamWrapper(ctx, path, assetID)
	if err != nil {
		return "", err
	}

	t.clientChan <- ClientInfo{
		client:  client,
		assetID: assetID,
		path:    path,
		video:   nil,
		audio:   nil,
		vhead:   -1,
		ahead:   -1,
	}

	return sw.GetMaster(assetID), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoIndex returns the video variant index playlist for a specific quality
func (t *Transcoder) GetVideoIndex(
	ctx context.Context,
	path string,
	video uint32,
	quality Quality,
	client string,
	assetID string,
) (string, error) {
	sw, err := t.getStreamWrapper(ctx, path, assetID)
	if err != nil {
		return "", err
	}
	t.clientChan <- ClientInfo{
		client:  client,
		assetID: assetID,
		path:    path,
		video:   &VideoKey{video, quality},
		audio:   nil,
		vhead:   -1,
		ahead:   -1,
	}
	return sw.GetVideoIndex(video, quality)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoSegment returns the path to a requested video segment, transcoding if necessary
func (t *Transcoder) GetVideoSegment(
	ctx context.Context,
	path string,
	video uint32,
	quality Quality,
	segment int32,
	client string,
	assetID string,
) (string, error) {
	sw, err := t.getStreamWrapper(ctx, path, assetID)
	if err != nil {
		return "", err
	}
	t.clientChan <- ClientInfo{
		client:  client,
		assetID: assetID,
		path:    path,
		video:   &VideoKey{video, quality},
		vhead:   segment,
		audio:   nil,
		ahead:   -1,
	}
	return sw.GetVideoSegment(video, quality, segment)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioIndex returns the audio index playlist for the specified audio index
func (t *Transcoder) GetAudioIndex(
	ctx context.Context,
	path string,
	audio uint32,
	client string,
	assetID string,
) (string, error) {
	sw, err := t.getStreamWrapper(ctx, path, assetID)
	if err != nil {
		return "", err
	}
	t.clientChan <- ClientInfo{
		client:  client,
		assetID: assetID,
		path:    path,
		audio:   &audio,
		vhead:   -1,
		ahead:   -1,
	}
	return sw.GetAudioIndex(audio)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioSegment returns the path to a requested audio segment, transcoding if necessary
func (t *Transcoder) GetAudioSegment(
	ctx context.Context,
	path string,
	audio uint32,
	segment int32,
	client string,
	assetID string,
) (string, error) {
	sw, err := t.getStreamWrapper(ctx, path, assetID)
	if err != nil {
		return "", err
	}
	t.clientChan <- ClientInfo{
		client:  client,
		assetID: assetID,
		path:    path,
		audio:   &audio,
		ahead:   segment,
		vhead:   -1,
	}
	return sw.GetAudioSegment(audio, segment)
}
