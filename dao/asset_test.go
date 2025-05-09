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

func Test_CreateAsset(t *testing.T) {
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

		count, err := dao.Count(ctx, &models.AssetProgress{}, nil)
		require.NoError(t, err)
		require.Equal(t, 0, count)

		count, err = dao.Count(ctx, &models.CourseProgress{}, nil)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateAsset(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAsset(t *testing.T) {
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

		assetResult := &models.Asset{}
		require.NoError(t, dao.GetAsset(ctx, assetResult, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE_ID: asset.ID}}))
		require.Equal(t, asset.ID, assetResult.ID)

		require.Nil(t, assetResult.VideoMetadata)
		require.Nil(t, assetResult.Progress)
		require.Empty(t, assetResult.Attachments)

		// Create Asset Progress
		assetProgress := &models.AssetProgress{AssetID: asset.ID}
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

		// Get asset with progress
		assetResult = &models.Asset{}
		require.NoError(t, dao.GetAsset(ctx, assetResult, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE_ID: asset.ID}}))
		require.Equal(t, asset.ID, assetResult.ID)
		require.NotNil(t, assetResult.Progress)
		require.Equal(t, assetProgress.ID, assetResult.Progress.ID)

	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.GetAsset(ctx, nil, nil), utils.ErrNilPtr)
	})

	t.Run("missing user id", func(t *testing.T) {
		dao, _ := setup(t)
		require.ErrorIs(t, dao.GetAsset(context.Background(), &models.Asset{}, nil), utils.ErrMissingUserId)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListAssets(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

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

		asset2 := &models.Asset{
			CourseID: course.ID,
			Title:    "Asset 2",
			Prefix:   sql.NullInt16{Int16: 2, Valid: true},
			Chapter:  "Chapter 2",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/02 asset.mp4",
			FileSize: 2048,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "5678",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset2))

		assets := []*models.Asset{}
		require.NoError(t, dao.ListAssets(ctx, &assets, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE_COURSE_ID: course.ID}}))
		require.Len(t, assets, 2)
		require.Equal(t, asset1.ID, assets[0].ID)
		require.Equal(t, asset2.ID, assets[1].ID)
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.ListAssets(ctx, nil, nil), utils.ErrNilPtr)
	})

	t.Run("missing user id", func(t *testing.T) {
		dao, _ := setup(t)
		require.ErrorIs(t, dao.ListAssets(context.Background(), &[]*models.Asset{}, nil), utils.ErrMissingUserId)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAsset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		originalAsset := &models.Asset{
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
		require.NoError(t, dao.CreateAsset(ctx, originalAsset))

		time.Sleep(1 * time.Millisecond)

		newAsset := &models.Asset{
			Base:     originalAsset.Base,
			Title:    "Asset 2",                            // Mutable
			Prefix:   sql.NullInt16{Int16: 2, Valid: true}, // Mutable
			Chapter:  "Chapter 2",                          // Mutable
			Type:     *types.NewAsset("html"),              // Mutable
			Path:     "/course-1/02 asset.html",            // Mutable
			FileSize: 2048,                                 // Mutable
			ModTime:  time.Now().Format(time.RFC3339Nano),  // Mutable
			Hash:     "5678",                               // Mutable
		}
		require.NoError(t, dao.UpdateAsset(ctx, newAsset))

		assertResult := &models.Asset{}
		require.NoError(t, dao.Get(ctx, assertResult, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE_ID: originalAsset.ID}}))
		require.Equal(t, newAsset.ID, assertResult.ID)                          // No change
		require.True(t, newAsset.CreatedAt.Equal(originalAsset.CreatedAt))      // No change
		require.Equal(t, newAsset.Title, assertResult.Title)                    // Changed
		require.Equal(t, newAsset.Prefix, assertResult.Prefix)                  // Changed
		require.Equal(t, newAsset.Chapter, assertResult.Chapter)                // Changed
		require.Equal(t, newAsset.Type, assertResult.Type)                      // Changed
		require.Equal(t, newAsset.Path, assertResult.Path)                      // Changed
		require.Equal(t, newAsset.FileSize, assertResult.FileSize)              // Changed
		require.Equal(t, newAsset.ModTime, assertResult.ModTime)                // Changed
		require.Equal(t, newAsset.Hash, assertResult.Hash)                      // Changed
		require.False(t, assertResult.UpdatedAt.Equal(originalAsset.UpdatedAt)) // Changed

		count, err := dao.Count(ctx, &models.AssetProgress{}, nil)
		require.NoError(t, err)
		require.Equal(t, 0, count)

		count, err = dao.Count(ctx, &models.CourseProgress{}, nil)
		require.NoError(t, err)
		require.Equal(t, 0, count)
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

		// Empty ID
		asset.ID = ""
		require.ErrorIs(t, dao.UpdateAsset(ctx, asset), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateAsset(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_AssetDeleteCascade(t *testing.T) {
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

	require.Nil(t, dao.Delete(ctx, course, nil))

	count, err := dao.Count(ctx, &models.Asset{}, nil)
	require.NoError(t, err)
	require.Zero(t, count)
}
