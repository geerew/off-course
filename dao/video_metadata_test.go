package dao

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func helper_createVideoMetadata(t *testing.T, ctx context.Context, dao *DAO, count int) ([]*models.Asset, []*models.VideoMetadata) {
	t.Helper()

	course := &models.Course{Title: "Course 1", Path: "/course-1"}
	require.NoError(t, dao.CreateCourse(ctx, course))

	assetGroup := &models.AssetGroup{
		CourseID: course.ID,
		Title:    "Asset Group 1",
		Prefix:   sql.NullInt16{Int16: 1, Valid: true},
		Module:   "Module 1",
	}
	require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

	assets := []*models.Asset{}
	videoMetadata := []*models.VideoMetadata{}
	for i := range count {
		asset := &models.Asset{
			CourseID:     course.ID,
			AssetGroupID: assetGroup.ID,
			Title:        fmt.Sprintf("Asset %d", i+1),
			Prefix:       sql.NullInt16{Int16: int16(i + 1), Valid: true},
			Module:       "Module 1",
			Type:         *types.NewAsset("mp4"),
			Path:         fmt.Sprintf("/course-1/0%d asset.mp4", i+1),
			FileSize:     1024,
			ModTime:      time.Now().Format(time.RFC3339Nano),
			Hash:         "1234",
		}
		assets = append(assets, asset)
		require.NoError(t, dao.CreateAsset(ctx, asset))

		metadata := &models.VideoMetadata{
			AssetID: asset.ID,
			VideoMetadataInfo: models.VideoMetadataInfo{
				Duration:   i + 1,
				Width:      1280,
				Height:     720,
				Codec:      "h264",
				Resolution: "720p",
			},
		}
		videoMetadata = append(videoMetadata, metadata)
		require.NoError(t, dao.CreateVideoMetadata(ctx, metadata))

		time.Sleep(1 * time.Millisecond)
	}

	return assets, videoMetadata
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateVideoMetadata(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

		asset := &models.Asset{
			CourseID:     course.ID,
			AssetGroupID: assetGroup.ID,
			Title:        "Asset 1",
			Prefix:       sql.NullInt16{Int16: 1, Valid: true},
			Module:       "Module 1",
			Type:         *types.NewAsset("mp4"),
			Path:         "/course-1/01 asset.mp4",
			FileSize:     1024,
			ModTime:      time.Now().Format(time.RFC3339Nano),
			Hash:         "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		videoMetadata := &models.VideoMetadata{
			AssetID: asset.ID,
			VideoMetadataInfo: models.VideoMetadataInfo{
				Duration:   120,
				Width:      1280,
				Height:     720,
				Codec:      "h264",
				Resolution: "720p",
			},
		}
		require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateVideoMetadata(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetVideoMetadata(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		_, videoMetadata := helper_createVideoMetadata(t, ctx, dao, 1)

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.VIDEO_METADATA_TABLE_ID: videoMetadata[0].ID})
		record, err := dao.GetVideoMetadata(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, videoMetadata[0].ID, record.ID)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		record, err := dao.GetVideoMetadata(ctx, nil)
		require.Nil(t, err)
		require.Nil(t, record)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListVideoMetadata(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		assets, videoMetadata := helper_createVideoMetadata(t, ctx, dao, 3)

		records, err := dao.ListVideoMetadata(ctx, nil)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, result := range records {
			require.Equal(t, videoMetadata[i].ID, result.ID)
			require.Equal(t, assets[i].ID, result.AssetID)
		}
	})

	t.Run("empty", func(t *testing.T) {
		dao, ctx := setup(t)

		records, err := dao.ListVideoMetadata(ctx, nil)
		require.Nil(t, err)
		require.Empty(t, records)
	})

	t.Run("order by", func(t *testing.T) {
		dao, ctx := setup(t)

		assets, videoMetadata := helper_createVideoMetadata(t, ctx, dao, 3)

		// Descending order by created_at
		opts := database.NewOptions().WithOrderBy(models.VIDEO_METADATA_TABLE_CREATED_AT + " DESC")

		records, err := dao.ListVideoMetadata(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, videoMetadata[2-i].ID, record.ID)
			require.Equal(t, assets[2-i].ID, record.AssetID)

		}

		// Ascending order by created_at
		opts = database.NewOptions().WithOrderBy(models.VIDEO_METADATA_TABLE_CREATED_AT + " ASC")

		records, err = dao.ListVideoMetadata(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, videoMetadata[i].ID, record.ID)
			require.Equal(t, assets[i].ID, record.AssetID)
		}
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		_, videoMetadata := helper_createVideoMetadata(t, ctx, dao, 3)

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.VIDEO_METADATA_TABLE_ID: videoMetadata[1].ID})
		records, err := dao.ListVideoMetadata(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 1)
		require.Equal(t, videoMetadata[1].ID, records[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, ctx := setup(t)

		_, videoMetadata := helper_createVideoMetadata(t, ctx, dao, 17)

		// First page with 10 records
		p := database.NewOptions().WithPagination(pagination.New(1, 10))
		records, err := dao.ListVideoMetadata(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 10)
		require.Equal(t, videoMetadata[0].ID, records[0].ID)
		require.Equal(t, videoMetadata[9].ID, records[9].ID)

		// Second page with remaining 7 records
		p = database.NewOptions().WithPagination(pagination.New(2, 10))
		records, err = dao.ListVideoMetadata(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 7)
		require.Equal(t, videoMetadata[10].ID, records[0].ID)
		require.Equal(t, videoMetadata[16].ID, records[6].ID)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateVideoMetadata(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		_, originalVideoMetadata := helper_createVideoMetadata(t, ctx, dao, 1)

		time.Sleep(1 * time.Millisecond)

		newVideoMetadata := &models.VideoMetadata{
			Base:    originalVideoMetadata[0].Base,
			AssetID: originalVideoMetadata[0].AssetID,
			VideoMetadataInfo: models.VideoMetadataInfo{
				Duration:   150,     // Mutable
				Width:      1920,    // Mutable
				Height:     1080,    // Mutable
				Codec:      "h265",  // Mutable
				Resolution: "1080p", // Mutable
			},
		}

		require.NoError(t, dao.UpdateVideoMetadata(ctx, newVideoMetadata))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.VIDEO_METADATA_TABLE_ID: originalVideoMetadata[0].ID})
		record, err := dao.GetVideoMetadata(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, originalVideoMetadata[0].ID, record.ID)                     // No change
		require.Equal(t, originalVideoMetadata[0].AssetID, record.AssetID)           // No change
		require.True(t, record.CreatedAt.Equal(originalVideoMetadata[0].CreatedAt))  // No change
		require.Equal(t, newVideoMetadata.Duration, record.Duration)                 // Changed
		require.Equal(t, newVideoMetadata.Width, record.Width)                       // Changed
		require.Equal(t, newVideoMetadata.Height, record.Height)                     // Changed
		require.Equal(t, newVideoMetadata.Codec, record.Codec)                       // Changed
		require.Equal(t, newVideoMetadata.Resolution, record.Resolution)             // Changed
		require.False(t, record.UpdatedAt.Equal(originalVideoMetadata[0].UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		_, videoMetadata := helper_createVideoMetadata(t, ctx, dao, 1)

		// Empty ID
		videoMetadata[0].ID = ""
		require.ErrorIs(t, dao.UpdateVideoMetadata(ctx, videoMetadata[0]), utils.ErrId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateVideoMetadata(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteVideoMetadata(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		_, videoMetadata := helper_createVideoMetadata(t, ctx, dao, 1)

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.VIDEO_METADATA_TABLE_ID: videoMetadata[0].ID})
		require.Nil(t, dao.DeleteVideoMetadata(ctx, opts))

		records, err := dao.ListVideoMetadata(ctx, opts)
		require.NoError(t, err)
		require.Empty(t, records)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		_, videoMetadata := helper_createVideoMetadata(t, ctx, dao, 1)

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.VIDEO_METADATA_TABLE_ID: "non-existent"})
		require.Nil(t, dao.DeleteVideoMetadata(ctx, opts))

		records, err := dao.ListVideoMetadata(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, videoMetadata[0].ID, records[0].ID)
	})

	t.Run("missing where", func(t *testing.T) {
		dao, ctx := setup(t)

		_, videoMetadata := helper_createVideoMetadata(t, ctx, dao, 1)

		require.ErrorIs(t, dao.DeleteVideoMetadata(ctx, nil), utils.ErrWhere)

		records, err := dao.ListVideoMetadata(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, videoMetadata[0].ID, records[0].ID)
	})

	t.Run("cascade", func(t *testing.T) {
		dao, ctx := setup(t)

		assets, _ := helper_createVideoMetadata(t, ctx, dao, 1)

		// TODO change to deleteAsset when done
		opts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assets[0].ID})
		require.Nil(t, dao.DeleteAssets(ctx, opts))

		records, err := dao.ListVideoMetadata(ctx, nil)
		require.NoError(t, err)
		require.Empty(t, records)
	})
}
