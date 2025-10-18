package hls

import (
	"context"
	"os"
	"path/filepath"

	"github.com/geerew/off-course/dao"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Transcoder manages all transcoding operations
type Transcoder struct {
	// All file streams currently running, index is asset_id
	streams    CMap[string, *FileStream]
	clientChan chan ClientInfo
	tracker    *Tracker
	dao        *dao.DAO
	assetID    string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTranscoder creates a new transcoder
func NewTranscoder(dao *dao.DAO) (*Transcoder, error) {
	out := Settings.CachePath
	os.MkdirAll(out, 0o755)
	dir, err := os.ReadDir(out)
	if err != nil {
		return nil, err
	}
	for _, d := range dir {
		err = os.RemoveAll(filepath.Join(out, d.Name()))
		if err != nil {
			return nil, err
		}
	}

	ret := &Transcoder{
		streams:    NewCMap[string, *FileStream](),
		clientChan: make(chan ClientInfo, 10),
		dao:        dao,
	}
	ret.tracker = NewTracker(ret)
	return ret, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getFileStream gets or creates a file stream
func (t *Transcoder) getFileStream(ctx context.Context, path string, assetID string) (*FileStream, error) {
	ret, _ := t.streams.GetOrCreate(assetID, func() *FileStream {
		t.assetID = assetID
		return t.newFileStream(ctx, path, assetID)
	})
	ret.ready.Wait()
	if ret.err != nil {
		t.streams.Remove(assetID)
		return nil, ret.err
	}
	return ret, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetMaster gets the master playlist for an asset
func (t *Transcoder) GetMaster(ctx context.Context, path string, client string, assetID string) (string, error) {
	stream, err := t.getFileStream(ctx, path, assetID)
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
	return stream.GetMaster(assetID), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoIndex gets the video index playlist
func (t *Transcoder) GetVideoIndex(
	ctx context.Context,
	path string,
	video uint32,
	quality Quality,
	client string,
	assetID string,
) (string, error) {
	stream, err := t.getFileStream(ctx, path, assetID)
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
	return stream.GetVideoIndex(video, quality)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioIndex gets the audio index playlist
func (t *Transcoder) GetAudioIndex(
	ctx context.Context,
	path string,
	audio uint32,
	client string,
	assetID string,
) (string, error) {
	stream, err := t.getFileStream(ctx, path, assetID)
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
	return stream.GetAudioIndex(audio)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoSegment gets a video segment
func (t *Transcoder) GetVideoSegment(
	ctx context.Context,
	path string,
	video uint32,
	quality Quality,
	segment int32,
	client string,
	assetID string,
) (string, error) {
	stream, err := t.getFileStream(ctx, path, assetID)
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
	return stream.GetVideoSegment(video, quality, segment)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioSegment gets an audio segment
func (t *Transcoder) GetAudioSegment(
	ctx context.Context,
	path string,
	audio uint32,
	segment int32,
	client string,
	assetID string,
) (string, error) {
	stream, err := t.getFileStream(ctx, path, assetID)
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
	return stream.GetAudioSegment(audio, segment)
}
