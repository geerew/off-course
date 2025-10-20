package hls

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// FileStream represents a file being transcoded
type FileStream struct {
	transcoder *Transcoder
	ready      sync.WaitGroup
	err        error
	Out        string
	Info       *MediaInfo
	videos     utils.CMap[VideoKey, *VideoStream]
	audios     utils.CMap[uint32, *AudioStream]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MediaInfo represents media file information
type MediaInfo struct {
	Path     string
	Duration float64
	Videos   []Video
	Audios   []Audio
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Video represents video metadata
type Video struct {
	Index     uint32
	Title     *string
	Language  *string
	Codec     string
	MimeCodec *string
	Width     uint32
	Height    uint32
	Bitrate   uint32
	IsDefault bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Audio represents audio metadata
type Audio struct {
	Index     uint32
	Title     *string
	Language  *string
	Codec     string
	MimeCodec *string
	Bitrate   uint32
	IsDefault bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// newFileStream creates a new file stream
func (t *Transcoder) newFileStream(ctx context.Context, path string, assetID string) *FileStream {
	ret := &FileStream{
		transcoder: t,
		Out:        filepath.Join(Settings.CachePath, assetID),
		videos:     utils.NewCMap[VideoKey, *VideoStream](),
		audios:     utils.NewCMap[uint32, *AudioStream](),
	}

	ret.ready.Add(1)
	go func() {
		defer ret.ready.Done()

		// Get asset metadata from database
		asset, err := t.dao.GetAsset(ctx, database.NewOptions().
			WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
			WithAssetMetadata())
		if err != nil {
			utils.Errf("HLS: Failed to get asset metadata for %s: %v\n", assetID, err)
			ret.err = err
			return
		}

		// Convert database models to HLS models
		var videos []Video
		var audios []Audio

		// Process video metadata
		if asset.AssetMetadata != nil && asset.AssetMetadata.VideoMetadata != nil {
			videoMeta := asset.AssetMetadata.VideoMetadata
			video := Video{
				Index:     0, // Default index
				Title:     nil,
				Language:  nil,
				Codec:     videoMeta.VideoCodec,
				MimeCodec: nil,
				Width:     uint32(videoMeta.Width),
				Height:    uint32(videoMeta.Height),
				Bitrate:   uint32(videoMeta.OverallBPS),
				IsDefault: true,
			}
			videos = append(videos, video)
		}

		// Process audio metadata
		if asset.AssetMetadata != nil && asset.AssetMetadata.AudioMetadata != nil {
			audioMeta := asset.AssetMetadata.AudioMetadata
			audio := Audio{
				Index:     0, // Default index
				Title:     nil,
				Language:  nil,
				Codec:     audioMeta.Codec,
				MimeCodec: nil,
				Bitrate:   uint32(audioMeta.BitRate),
				IsDefault: true,
			}
			audios = append(audios, audio)
		}

		// Create MediaInfo with real data
		duration := 0.0
		if asset.AssetMetadata != nil && asset.AssetMetadata.VideoMetadata != nil {
			duration = float64(asset.AssetMetadata.VideoMetadata.DurationSec)
		}

		info := &MediaInfo{
			Path:     path,
			Duration: duration,
			Videos:   videos,
			Audios:   audios,
		}
		ret.Info = info
	}()

	return ret
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Kill stops all streams
func (fs *FileStream) Kill() {
	fs.videos.ForEach(func(_ VideoKey, s *VideoStream) {
		s.Kill()
	})
	fs.audios.ForEach(func(_ uint32, s *AudioStream) {
		s.Kill()
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Destroy removes all transcoded files
func (fs *FileStream) Destroy() {
	utils.Infof("HLS: Removing all transcode cache files for %s\n", fs.Info.Path)
	fs.Kill()
	_ = os.RemoveAll(fs.Out)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetMaster generates the master playlist
func (fs *FileStream) GetMaster(assetID string) string {
	master := "#EXTM3U\n"

	// TODO: support multiples audio qualities (and original)
	for _, audio := range fs.Info.Audios {
		master += "#EXT-X-MEDIA:TYPE=AUDIO,"
		master += "GROUP-ID=\"audio\","
		if audio.Language != nil {
			master += fmt.Sprintf("LANGUAGE=\"%s\",", *audio.Language)
		}
		if audio.Title != nil {
			master += fmt.Sprintf("NAME=\"%s\",", *audio.Title)
		} else if audio.Language != nil {
			master += fmt.Sprintf("NAME=\"%s\",", *audio.Language)
		} else {
			master += fmt.Sprintf("NAME=\"Audio %d\",", audio.Index)
		}
		if audio.IsDefault {
			master += "DEFAULT=YES,"
		}
		master += "CHANNELS=\"2\","
		master += fmt.Sprintf("URI=\"audio/%d/index.m3u8\"\n", audio.Index)
	}
	master += "\n"

	// codec is the prefix + the level, the level is not part of the codec we want to compare for the same_codec check bellow
	transcode_prefix := "avc1.6400"
	transcode_codec := transcode_prefix + "28"
	audio_codec := "mp4a.40.2"

	var def_video *Video
	for _, video := range fs.Info.Videos {
		if video.IsDefault {
			def_video = &video
			break
		}
	}
	if def_video == nil && len(fs.Info.Videos) > 0 {
		def_video = &fs.Info.Videos[0]
	}

	if def_video != nil {
		var qualities []Quality
		for _, q := range Qualities {
			if q.Height() < def_video.Height {
				qualities = append(qualities, q)
			}
		}
		transcode_count := len(qualities)

		// NoResize is the same idea as Original but we change the codec.
		// This is only needed when the original's codec is different from what we would transcode it to.
		if def_video.MimeCodec == nil || !strings.HasPrefix(*def_video.MimeCodec, transcode_prefix) {
			qualities = append(qualities, NoResize)
		}
		qualities = append(qualities, Original)

		for _, quality := range qualities {
			for _, video := range fs.Info.Videos {
				master += "#EXT-X-MEDIA:TYPE=VIDEO,"
				master += fmt.Sprintf("GROUP-ID=\"%s\",", quality)
				if video.Language != nil {
					master += fmt.Sprintf("LANGUAGE=\"%s\",", *video.Language)
				}
				if video.Title != nil {
					master += fmt.Sprintf("NAME=\"%s\",", *video.Title)
				} else if video.Language != nil {
					master += fmt.Sprintf("NAME=\"%s\",", *video.Language)
				} else {
					master += fmt.Sprintf("NAME=\"Video %d\",", video.Index)
				}
				if video == *def_video {
					master += "DEFAULT=YES\n"
				} else {
					master += fmt.Sprintf("URI=\"%d/%s/index.m3u8\"\n", video.Index, quality)
				}
			}
		}
		master += "\n"

		aspectRatio := float32(def_video.Width) / float32(def_video.Height)
		for i, quality := range qualities {
			if i >= transcode_count {
				// original & noresize streams
				bitrate := float64(def_video.Bitrate)
				master += "#EXT-X-STREAM-INF:"
				// For original quality, use the video's actual bitrate
				master += fmt.Sprintf("AVERAGE-BANDWIDTH=%d,", int(bitrate*0.8))
				master += fmt.Sprintf("BANDWIDTH=%d,", int(bitrate))
				master += fmt.Sprintf("RESOLUTION=%dx%d,", def_video.Width, def_video.Height)
				if quality != Original {
					master += fmt.Sprintf("CODECS=\"%s\",", strings.Join([]string{transcode_codec, audio_codec}, ","))
				} else if def_video.MimeCodec != nil {
					master += fmt.Sprintf("CODECS=\"%s\",", strings.Join([]string{*def_video.MimeCodec, audio_codec}, ","))
				}
				master += "AUDIO=\"audio\","
				master += "CLOSED-CAPTIONS=NONE\n"
				master += fmt.Sprintf("/api/hls/%s/video/%d/%s/index.m3u8\n", assetID, def_video.Index, quality)
				continue
			}

			master += "#EXT-X-STREAM-INF:"
			master += fmt.Sprintf("AVERAGE-BANDWIDTH=%d,", quality.AverageBitrate())
			master += fmt.Sprintf("BANDWIDTH=%d,", quality.MaxBitrate())
			master += fmt.Sprintf("RESOLUTION=%dx%d,", int(aspectRatio*float32(quality.Height())+0.5), quality.Height())
			master += fmt.Sprintf("CODECS=\"%s\",", strings.Join([]string{transcode_codec, audio_codec}, ","))
			master += "AUDIO=\"audio\","
			master += "CLOSED-CAPTIONS=NONE\n"
			master += fmt.Sprintf("/api/hls/%s/video/%d/%s/index.m3u8\n", assetID, def_video.Index, quality)
		}
	}

	return master
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getVideoStream gets or creates a video stream
func (fs *FileStream) getVideoStream(idx uint32, quality Quality) (*VideoStream, error) {
	stream, _ := fs.videos.GetOrCreate(VideoKey{idx, quality}, func() *VideoStream {
		ret, _ := NewVideoStream(fs, idx, quality)
		return ret
	})
	stream.ready.Wait()
	return stream, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoIndex gets the video index playlist
func (fs *FileStream) GetVideoIndex(idx uint32, quality Quality) (string, error) {
	stream, err := fs.getVideoStream(idx, quality)
	if err != nil {
		return "", err
	}
	return stream.GetIndex()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoSegment gets a video segment
func (fs *FileStream) GetVideoSegment(idx uint32, quality Quality, segment int32) (string, error) {
	stream, err := fs.getVideoStream(idx, quality)
	if err != nil {
		return "", err
	}
	return stream.GetSegment(segment)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getAudioStream gets or creates an audio stream
func (fs *FileStream) getAudioStream(audio uint32) (*AudioStream, error) {
	stream, _ := fs.audios.GetOrCreate(audio, func() *AudioStream {
		ret, _ := NewAudioStream(fs, audio)
		return ret
	})
	stream.ready.Wait()
	return stream, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioIndex gets the audio index playlist
func (fs *FileStream) GetAudioIndex(audio uint32) (string, error) {
	stream, err := fs.getAudioStream(audio)
	if err != nil {
		return "", nil
	}
	return stream.GetIndex()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioSegment gets an audio segment
func (fs *FileStream) GetAudioSegment(audio uint32, segment int32) (string, error) {
	stream, err := fs.getAudioStream(audio)
	if err != nil {
		return "", nil
	}
	return stream.GetSegment(segment)
}
