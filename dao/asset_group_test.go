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

func helper_createAssetGroups(t *testing.T, ctx context.Context, dao *DAO, numCourses int) ([]*models.Course, []*models.AssetGroup, []*models.Asset, []*models.Attachment) {
	t.Helper()

	allCourses := []*models.Course{}
	allAssetGroups := []*models.AssetGroup{}
	allAssets := []*models.Asset{}
	allAttachments := []*models.Attachment{}

	for i := 0; i < numCourses; i++ {
		course := &models.Course{Title: fmt.Sprintf("Course %d", i+1), Path: fmt.Sprintf("/course %d", i+1)}
		require.NoError(t, dao.CreateCourse(ctx, course))
		allCourses = append(allCourses, course)

		// Create 3 asset groups with 3 assets and 2 attachments each, reversed
		for _, assetGroupIndex := range []int{3, 2, 1} {
			assetGroupPrefix := fmt.Sprintf("%02d", assetGroupIndex)

			assetGroup := &models.AssetGroup{
				CourseID: course.ID,
				Title:    fmt.Sprintf("Asset Group %d", assetGroupIndex),
				Prefix:   sql.NullInt16{Int16: int16(assetGroupIndex), Valid: true},
				Module:   fmt.Sprintf("Module %d", assetGroupIndex),
			}
			require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))
			allAssetGroups = append(allAssetGroups, assetGroup)
			time.Sleep(1 * time.Millisecond)

			// 3 assets, reversed sub-prefix: 3,2,1
			for _, assetIndex := range []int{3, 2, 1} {
				asset := &models.Asset{
					CourseID:     course.ID,
					AssetGroupID: assetGroup.ID,
					Title:        fmt.Sprintf("Asset %d", assetIndex),
					Prefix:       sql.NullInt16{Int16: int16(assetIndex), Valid: true},
					SubPrefix:    sql.NullInt16{Int16: int16(assetIndex), Valid: true},
					Module:       fmt.Sprintf("Module %d", assetIndex),
					Type:         *types.NewAsset("mp4"),
					Path:         fmt.Sprintf("%s/%s asset {%02d}.mp4", course.Path, assetGroupPrefix, assetIndex),
				}
				require.NoError(t, dao.CreateAsset(ctx, asset))
				allAssets = append(allAssets, asset)
				time.Sleep(1 * time.Millisecond)
			}

			// Create 2 attachments, reversed: 2,1
			for _, n := range []int{2, 1} {
				attachment := &models.Attachment{
					AssetGroupID: assetGroup.ID,
					Title:        fmt.Sprintf("%s Attachment %d", assetGroupPrefix, n),
					Path:         fmt.Sprintf("%s/%s attachment %d.pdf", course.Path, assetGroupPrefix, n),
				}
				require.NoError(t, dao.CreateAttachment(ctx, attachment))
				allAttachments = append(allAttachments, attachment)
				time.Sleep(1 * time.Millisecond)
			}
		}

		time.Sleep(1 * time.Millisecond)

	}

	return allCourses, allAssetGroups, allAssets, allAttachments
}

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

	t.Run("nil pointer", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateAssetGroup(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{}
		require.ErrorIs(t, dao.CreateAssetGroup(ctx, assetGroup), utils.ErrCourseId)

		assetGroup.CourseID = course.ID
		require.ErrorIs(t, dao.CreateAssetGroup(ctx, assetGroup), utils.ErrTitle)

		assetGroup.Title = "Asset Group 1"
		require.ErrorIs(t, dao.CreateAssetGroup(ctx, assetGroup), utils.ErrPrefix)

		assetGroup.Prefix = sql.NullInt16{Int16: 1, Valid: true}
		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssetGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		_, allGroups, allAssets, allAttachments := helper_createAssetGroups(t, ctx, dao, 1)

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_GROUP_TABLE_ID: allGroups[0].ID})
		record, err := dao.GetAssetGroup(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, allGroups[0].ID, record.ID)

		require.Len(t, record.Assets, 3)
		require.Equal(t, allAssets[2].ID, record.Assets[0].ID)
		require.Equal(t, allAssets[1].ID, record.Assets[1].ID)
		require.Equal(t, allAssets[0].ID, record.Assets[2].ID)

		require.Len(t, record.Attachments, 2)
		require.Equal(t, allAttachments[1].ID, record.Attachments[0].ID)
		require.Equal(t, allAttachments[0].ID, record.Attachments[1].ID)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		record, err := dao.GetAssetGroup(ctx, nil)
		require.Nil(t, err)
		require.Nil(t, record)
	})

	t.Run("missing principal", func(t *testing.T) {
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

		dbOpts := database.NewOptions().WithProgress()
		record, err := dao.GetAssetGroup(context.Background(), dbOpts)
		require.ErrorIs(t, err, utils.ErrPrincipal)
		require.Nil(t, record)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListAssetGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		helper_createAssetGroups(t, ctx, dao, 3)

		records, err := dao.ListAssetGroups(ctx, nil)
		require.Nil(t, err)
		require.Len(t, records, 9)

		// Simple relation check
		require.Len(t, records[0].Attachments, 2)
		require.Len(t, records[0].Assets, 3)
	})

	t.Run("empty", func(t *testing.T) {
		dao, ctx := setup(t)

		records, err := dao.ListAssetGroups(ctx, nil)
		require.Nil(t, err)
		require.Empty(t, records)
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		courses, assetGroups, _, _ := helper_createAssetGroups(t, ctx, dao, 3)

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_GROUP_TABLE_COURSE_ID: courses[1].ID})
		records, err := dao.ListAssetGroups(ctx, dbOpts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		require.Equal(t, assetGroups[5].ID, records[0].ID)
		require.Equal(t, assetGroups[4].ID, records[1].ID)
		require.Equal(t, assetGroups[3].ID, records[2].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroups := []*models.AssetGroup{}
		for i := range 17 {
			assetGroup := &models.AssetGroup{
				CourseID: course.ID,
				Title:    fmt.Sprintf("Asset Group %d", i),
				Prefix:   sql.NullInt16{Int16: int16(i), Valid: true},
				Module:   fmt.Sprintf("Module %d", i),
			}
			require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))
			assetGroups = append(assetGroups, assetGroup)
		}

		// First page with 10 records
		p := database.NewOptions().WithPagination(pagination.New(1, 10))
		records, err := dao.ListAssetGroups(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 10)
		require.Equal(t, assetGroups[0].ID, records[0].ID)
		require.Equal(t, assetGroups[9].ID, records[9].ID)

		// Second page with remaining 7 records
		p = database.NewOptions().WithPagination(pagination.New(2, 10))
		records, err = dao.ListAssetGroups(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 7)
		require.Equal(t, assetGroups[10].ID, records[0].ID)
		require.Equal(t, assetGroups[16].ID, records[6].ID)
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

		updatedAssetGroup := &models.AssetGroup{
			Base:     originalAssetGroup.Base,
			CourseID: course.ID,
			Title:    "Asset Group 2",
			Prefix:   sql.NullInt16{Int16: 2, Valid: true},
			Module:   "Module 2",
		}
		require.NoError(t, dao.UpdateAssetGroup(ctx, updatedAssetGroup))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_GROUP_TABLE_ID: originalAssetGroup.ID})
		record, err := dao.GetAssetGroup(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, originalAssetGroup.ID, record.ID)                    // No change
		require.Equal(t, originalAssetGroup.CourseID, record.CourseID)        // No change
		require.True(t, record.CreatedAt.Equal(originalAssetGroup.CreatedAt)) // No change
		require.Equal(t, updatedAssetGroup.Title, record.Title)               // Changed
		require.Equal(t, updatedAssetGroup.Prefix, record.Prefix)             // Changed
		require.Equal(t, updatedAssetGroup.Module, record.Module)             // Changed
		require.NotEqual(t, originalAssetGroup.UpdatedAt, record.UpdatedAt)   // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{}

		// Course ID
		require.ErrorIs(t, dao.UpdateAssetGroup(ctx, assetGroup), utils.ErrCourseId)
		assetGroup.CourseID = course.ID

		// Title
		require.ErrorIs(t, dao.UpdateAssetGroup(ctx, assetGroup), utils.ErrTitle)
		assetGroup.Title = "Asset 1"

		// Prefix
		require.ErrorIs(t, dao.UpdateAssetGroup(ctx, assetGroup), utils.ErrPrefix)
		assetGroup.Prefix = sql.NullInt16{Int16: 1, Valid: true}

		// ID
		require.ErrorIs(t, dao.UpdateAssetGroup(ctx, assetGroup), utils.ErrId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateAssetGroup(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAssetGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course", Path: "/course"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_GROUP_TABLE_ID: assetGroup.ID})
		require.Nil(t, dao.DeleteAssetGroups(ctx, opts))

		// TODO add list when supported
		// records, err := dao.ListAssetGroups(ctx, opts)
		// require.NoError(t, err)
		// require.Empty(t, records)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course", Path: "/course"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_GROUP_TABLE_ID: "non-existent"})
		require.Nil(t, dao.DeleteAssetGroups(ctx, opts))

		// records, err := dao.ListAssetGroups(ctx, nil)
		// require.NoError(t, err)
		// require.Len(t, records, 1)
		// require.Equal(t, assetGroup.ID, records[0].ID)
	})

	t.Run("missing where", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course", Path: "/course"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

		require.ErrorIs(t, dao.DeleteAssetGroups(ctx, nil), utils.ErrWhere)

		// records, err := dao.ListAssetGroups(ctx, nil)
		// require.NoError(t, err)
		// require.Len(t, records, 1)
		// require.Equal(t, assetGroup.ID, records[0].ID)
	})

	t.Run("cascade", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course", Path: "/course"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: course.ID})
		require.Nil(t, dao.DeleteCourses(ctx, dbOpts))

		// records, err := dao.ListAssetGroups(ctx, nil)
		// require.NoError(t, err)
		// require.Empty(t, records)
	})
}
