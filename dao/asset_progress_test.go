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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateOrUpdateAssetProgress(t *testing.T) {
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

		assetProgress := &models.AssetProgress{AssetID: asset.ID}
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

		// Check the course progress was created for this user
		courseProgress := &models.CourseProgress{}
		options := &database.Options{
			Where: squirrel.And{
				squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: course.ID},
				squirrel.Eq{models.COURSE_PROGRESS_TABLE_USER_ID: ctx.Value(types.UserContextKey)},
			},
		}
		require.NoError(t, dao.GetCourseProgress(ctx, courseProgress, options))
		require.Equal(t, courseProgress.CourseID, course.ID)
	})

	t.Run("update", func(t *testing.T) {
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
			Duration:   100,
			Width:      1920,
			Height:     1080,
			Resolution: "1080p",
			Codec:      "h264",
		}
		require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata))

		courseProgressOptions := &database.Options{
			Where: squirrel.And{
				squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: course.ID},
				squirrel.Eq{models.COURSE_PROGRESS_TABLE_USER_ID: ctx.Value(types.UserContextKey)},
			},
		}

		assetProgress := &models.AssetProgress{AssetID: asset.ID}

		// Create asset progress
		{
			require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

			courseProgress := &models.CourseProgress{}
			require.NoError(t, dao.GetCourseProgress(ctx, courseProgress, courseProgressOptions))
			require.Equal(t, course.ID, courseProgress.CourseID)
			require.Equal(t, 0, courseProgress.Percent)

		}

		// Update asset video position
		{
			assetProgress.VideoPos = 20
			require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

			courseProgress := &models.CourseProgress{}
			require.NoError(t, dao.GetCourseProgress(ctx, courseProgress, courseProgressOptions))
			require.Equal(t, course.ID, courseProgress.CourseID)
			require.Equal(t, 20, courseProgress.Percent)
		}

		// Update asset to completed
		{
			assetProgress.Completed = true
			require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

			courseProgress := &models.CourseProgress{}
			require.NoError(t, dao.GetCourseProgress(ctx, courseProgress, courseProgressOptions))
			require.Equal(t, course.ID, courseProgress.CourseID)
			require.Equal(t, 100, courseProgress.Percent)
		}
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateOrUpdateAssetProgress(ctx, "", nil), utils.ErrNilPtr)
	})

	t.Run("missing user id", func(t *testing.T) {
		dao, _ := setup(t)
		require.ErrorIs(t, dao.CreateOrUpdateAssetProgress(context.Background(), "", &models.AssetProgress{}), utils.ErrMissingUserId)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_AssetProgressDeleteCascade(t *testing.T) {
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

	assetProgress := &models.AssetProgress{
		AssetID: asset.ID,
	}
	require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

	require.NoError(t, Delete(ctx, dao, asset, nil))

	err := dao.GetAssetProgress(ctx, assetProgress, &database.Options{Where: squirrel.Eq{models.ASSET_PROGRESS_TABLE_ID: assetProgress.ID}})
	require.ErrorIs(t, err, sql.ErrNoRows)
}
