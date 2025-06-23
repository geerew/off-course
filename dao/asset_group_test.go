package dao

import (
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

func Test_CreateAssetGroup(t *testing.T) {
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
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateAssetGroup(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssetGroup(t *testing.T) {
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

		assetResult := &models.AssetGroup{}
		require.NoError(t, dao.GetAssetGroup(ctx, assetResult, &database.Options{Where: squirrel.Eq{models.ASSET_GROUP_TABLE_ID: assetGroup.ID}}))
		require.Equal(t, assetGroup.ID, assetResult.ID)

		require.Empty(t, assetResult.Assets)
		require.Empty(t, assetResult.Attachments)

		// Add attachment
		attachment := &models.Attachment{
			AssetGroupID: assetGroup.ID,
			Title:        "Attachment 1",
			Path:         "/course-1/attachment 1.pdf",
		}
		require.NoError(t, dao.CreateAttachment(ctx, attachment))

		// Add asset
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

		// Check attachment and asset
		require.NoError(t, dao.GetAssetGroup(ctx, assetResult, &database.Options{Where: squirrel.Eq{models.ASSET_GROUP_TABLE_ID: assetGroup.ID}}))
		require.Equal(t, 1, len(assetResult.Attachments))
		require.Equal(t, attachment.ID, assetResult.Attachments[0].ID)
		require.Equal(t, 1, len(assetResult.Assets))
		require.Equal(t, asset.ID, assetResult.Assets[0].ID)
		require.Equal(t, assetGroup.ID, assetResult.Assets[0].AssetGroupID)
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.GetAssetGroup(ctx, nil, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListAssetGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		asset1 := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, asset1))

		asset2 := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group 2",
			Prefix:   sql.NullInt16{Int16: 2, Valid: true},
			Module:   "Module 2",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, asset2))

		assetGroups := []*models.AssetGroup{}
		require.NoError(t, dao.ListAssetGroups(ctx, &assetGroups, &database.Options{Where: squirrel.Eq{models.ASSET_GROUP_TABLE_COURSE_ID: course.ID}}))
		require.Len(t, assetGroups, 2)
		require.Equal(t, asset1.ID, assetGroups[0].ID)
		require.Equal(t, asset2.ID, assetGroups[1].ID)
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.ListAssetGroups(ctx, nil, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAssetGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		originalAssetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, originalAssetGroup))

		time.Sleep(1 * time.Millisecond)

		newAssetGroup := &models.AssetGroup{
			Base:   originalAssetGroup.Base,
			Title:  "Asset Group 2",                      // Mutable
			Prefix: sql.NullInt16{Int16: 2, Valid: true}, // Mutable
			Module: "Module 2",                           // Mutable
		}
		require.NoError(t, dao.UpdateAssetGroup(ctx, newAssetGroup))

		assertResult := &models.AssetGroup{Base: models.Base{ID: originalAssetGroup.ID}}
		require.NoError(t, dao.GetAssetGroup(ctx, assertResult, nil))
		require.Equal(t, newAssetGroup.ID, assertResult.ID)                          // No change
		require.True(t, newAssetGroup.CreatedAt.Equal(originalAssetGroup.CreatedAt)) // No change
		require.Equal(t, newAssetGroup.Title, assertResult.Title)                    // Changed
		require.Equal(t, newAssetGroup.Prefix, assertResult.Prefix)                  // Changed
		require.Equal(t, newAssetGroup.Module, assertResult.Module)                  // Changed
		require.False(t, assertResult.UpdatedAt.Equal(originalAssetGroup.UpdatedAt)) // Changed

		count, err := Count(ctx, dao, &models.AssetProgress{}, nil)
		require.NoError(t, err)
		require.Equal(t, 0, count)

		count, err = Count(ctx, dao, &models.CourseProgress{}, nil)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})

	t.Run("invalid", func(t *testing.T) {
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

		// Empty ID
		assetGroup.ID = ""
		require.ErrorIs(t, dao.UpdateAssetGroup(ctx, assetGroup), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateAssetGroup(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_AssetGroupDeleteCascade(t *testing.T) {
	dao, ctx := setup(t)

	course := &models.Course{Title: "Course", Path: "/course"}
	require.NoError(t, dao.CreateCourse(ctx, course))

	assetGroup := &models.AssetGroup{
		CourseID: course.ID,
		Title:    "Asset Group 1",
		Prefix:   sql.NullInt16{Int16: 1, Valid: true},
		Module:   "Module 1",
	}
	require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

	require.Nil(t, Delete(ctx, dao, course, nil))

	count, err := Count(ctx, dao, &models.AssetGroup{}, nil)
	require.NoError(t, err)
	require.Zero(t, count)
}
