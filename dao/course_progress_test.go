package dao

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

func Test_CreateCourseProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.Create(ctx, course))

		courseProgress := &models.CourseProgress{CourseID: course.ID}
		require.NoError(t, dao.CreateCourseProgress(ctx, courseProgress))
	})

	t.Run("nil pointer", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateCourseProgress(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("invalid course id", func(t *testing.T) {
		dao, ctx := setup(t)
		courseProgress := &models.CourseProgress{CourseID: "invalid"}
		require.ErrorContains(t, dao.CreateCourseProgress(ctx, courseProgress), "FOREIGN KEY constraint failed")
	})

	t.Run("missing user id", func(t *testing.T) {
		dao, _ := setup(t)
		require.ErrorIs(t, dao.CreateCourseProgress(context.Background(), &models.CourseProgress{}), utils.ErrMissingUserId)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RefreshCourseProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))
		require.Nil(t, course.Progress)

		// Create asset
		asset1 := &models.Asset{
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
		require.NoError(t, dao.CreateAsset(ctx, asset1))

		// Create video metadata
		videoMetadata := &models.VideoMetadata{
			AssetID:    asset1.ID,
			Duration:   10,
			Width:      1920,
			Height:     1080,
			Resolution: "1080p",
			Codec:      "h264",
		}
		require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata))

		assetProgress := &models.AssetProgress{}

		// Set asset1 progress
		{
			assetProgress = &models.AssetProgress{AssetID: asset1.ID, VideoPos: 5}
			require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

			require.NoError(t, dao.GetCourse(ctx, course, nil))
			require.NotNil(t, course.Progress)
			require.True(t, course.Progress.Started)
			require.False(t, course.Progress.StartedAt.IsZero())
			require.Equal(t, 50, course.Progress.Percent)
			require.True(t, course.Progress.CompletedAt.IsZero())
		}

		// Set asset progress (video_pos = 0)
		{
			assetProgress.VideoPos = 0
			require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

			require.NoError(t, dao.GetCourse(ctx, course, nil))
			require.NotNil(t, course.Progress)
			require.False(t, course.Progress.Started)
			require.True(t, course.Progress.StartedAt.IsZero())
			require.Zero(t, course.Progress.Percent)
			require.True(t, course.Progress.CompletedAt.IsZero())
		}

		// Set asset progress (completed = true)
		{
			assetProgress.Completed = true
			require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

			require.NoError(t, dao.GetCourse(ctx, course, nil))
			require.NotNil(t, course.Progress)
			require.True(t, course.Progress.Started)
			require.False(t, course.Progress.StartedAt.IsZero())
			require.Equal(t, 100, course.Progress.Percent)
			require.False(t, course.Progress.CompletedAt.IsZero())
		}

		// Add another asset
		asset2 := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 2",
			Prefix:   sql.NullInt16{Int16: 2, Valid: true},
			Chapter:  "Chapter 2",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/02 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "5678",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset2))

		// Create video metadata for asset2
		videoMetadata2 := &models.VideoMetadata{
			AssetID:    asset2.ID,
			Duration:   10,
			Width:      1920,
			Height:     1080,
			Resolution: "1080p",
			Codec:      "h264",
		}
		require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata2))

		// Set asset2 progress
		assetProgress2 := &models.AssetProgress{AssetID: asset2.ID, VideoPos: 0}
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress2))

		// Check course progress
		{
			require.NoError(t, dao.GetCourse(ctx, course, nil))
			require.NotNil(t, course.Progress)
			require.True(t, course.Progress.Started)
			require.False(t, course.Progress.StartedAt.IsZero())
			require.Equal(t, 50, course.Progress.Percent)
			require.True(t, course.Progress.CompletedAt.IsZero())
		}
	})

	t.Run("invalid course id", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.RefreshCourseProgress(ctx, ""), utils.ErrInvalidId)
	})

	t.Run("missing user id", func(t *testing.T) {
		dao, _ := setup(t)
		require.ErrorIs(t, dao.RefreshCourseProgress(context.Background(), "1234"), utils.ErrMissingUserId)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CourseProgressDeleteCascade(t *testing.T) {
	dao, ctx := setup(t)

	course := &models.Course{Title: "Course", Path: "/course"}
	require.NoError(t, dao.CreateCourse(ctx, course))

	require.NoError(t, dao.RefreshCourseProgress(ctx, course.ID))

	courseProgress := &models.CourseProgress{}
	require.NoError(t, dao.GetCourseProgress(ctx, courseProgress, &database.Options{Where: squirrel.Eq{models.COURSE_PROGRESS_COURSE_ID: course.ID}}))

	require.NoError(t, dao.Delete(ctx, courseProgress, nil))

	count, err := dao.Count(ctx, &models.CourseProgress{}, nil)
	require.NoError(t, err)
	require.Zero(t, count)
}
