package dao

import (
	"database/sql"
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateVideoMetadata(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		videoMetadata := &models.VideoMetadata{
			AssetID:    asset.ID,
			Duration:   120,
			Width:      1280,
			Height:     720,
			Codec:      "h264",
			Resolution: "720p",
		}
		require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateVideoMetadata(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateVideoMetadata(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		originalVideoMetadata := &models.VideoMetadata{
			AssetID:    asset.ID,
			Duration:   120,
			Width:      1280,
			Height:     720,
			Codec:      "h264",
			Resolution: "720p",
		}
		require.NoError(t, dao.CreateVideoMetadata(ctx, originalVideoMetadata))

		time.Sleep(1 * time.Millisecond)

		newVideoMetadata := &models.VideoMetadata{
			Base:       originalVideoMetadata.Base,
			AssetID:    originalVideoMetadata.AssetID,
			Duration:   150,     // Mutable
			Width:      1920,    // Mutable
			Height:     1080,    // Mutable
			Codec:      "h265",  // Mutable
			Resolution: "1080p", // Mutable
		}
		require.NoError(t, dao.UpdateVideoMetadata(ctx, newVideoMetadata))

		assertResult := &models.VideoMetadata{Base: models.Base{ID: originalVideoMetadata.ID}}
		require.NoError(t, dao.GetVideoMetadata(ctx, assertResult, nil))
		require.Equal(t, newVideoMetadata.ID, assertResult.ID)                             // No change
		require.True(t, newVideoMetadata.CreatedAt.Equal(originalVideoMetadata.CreatedAt)) // No change
		require.Equal(t, newVideoMetadata.AssetID, assertResult.AssetID)                   // No change
		require.Equal(t, newVideoMetadata.Duration, assertResult.Duration)                 // Changed
		require.Equal(t, newVideoMetadata.Width, assertResult.Width)                       // Changed
		require.Equal(t, newVideoMetadata.Height, assertResult.Height)                     // Changed
		require.Equal(t, newVideoMetadata.Codec, assertResult.Codec)                       // Changed
		require.Equal(t, newVideoMetadata.Resolution, assertResult.Resolution)             // Changed
		require.False(t, assertResult.UpdatedAt.Equal(originalVideoMetadata.UpdatedAt))    // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		asset := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Chapter:  "Chapter 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		videoMetadata := &models.VideoMetadata{
			AssetID:    asset.ID,
			Duration:   120,
			Width:      1280,
			Height:     720,
			Codec:      "h264",
			Resolution: "720p",
		}
		require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata))

		// Empty ID
		videoMetadata.ID = ""
		require.ErrorIs(t, dao.UpdateVideoMetadata(ctx, videoMetadata), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateVideoMetadata(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_VideoMetadataDeleteCascade(t *testing.T) {
	dao, ctx := setup(t)

	course := &models.Course{Title: "Course", Path: "/course"}
	require.NoError(t, dao.CreateCourse(ctx, course))

	asset := &models.Asset{
		CourseID: course.ID,
		Title:    "Asset 1",
		Prefix:   sql.NullInt16{Int16: 1, Valid: true},
		Chapter:  "Chapter 1",
		Type:     *types.NewAsset("mp4"),
		Path:     "/course-1/01 asset.mp4",
		FileSize: 1024,
		ModTime:  time.Now().Format(time.RFC3339Nano),
		Hash:     "1234",
	}
	require.NoError(t, dao.CreateAsset(ctx, asset))

	videoMetadata := &models.VideoMetadata{
		AssetID:    asset.ID,
		Duration:   120,
		Width:      1280,
		Height:     720,
		Codec:      "h264",
		Resolution: "720p",
	}
	require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata))

	require.Nil(t, Delete(ctx, dao, asset, nil))

	count, err := dao.Count(ctx, &models.VideoMetadata{}, nil)
	require.NoError(t, err)
	require.Zero(t, count)
}
