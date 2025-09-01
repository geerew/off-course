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

func Test_UpsertAssetProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		lesson := &models.Lesson{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateLesson(ctx, lesson))

		asset := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Create asset progress
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 50},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_PROGRESS_TABLE_ID: assetProgress.ID})
		record, err := dao.GetAssetProgress(ctx, opts)
		require.NoError(t, err)
		require.Equal(t, asset.ID, record.AssetID)
		require.Equal(t, 50, record.VideoPos)
		require.False(t, record.Completed)
		require.True(t, record.CompletedAt.IsZero())

		// Update asset progress
		assetProgress.VideoPos = 100
		assetProgress.Completed = true
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		record, err = dao.GetAssetProgress(ctx, opts)
		require.NoError(t, err)
		require.Equal(t, asset.ID, record.AssetID)
		require.Equal(t, 100, record.VideoPos)
		require.True(t, record.Completed)
		require.False(t, record.CompletedAt.IsZero())
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.UpsertAssetProgress(ctx, "", nil), utils.ErrNilPtr)
	})

	t.Run("missing principal", func(t *testing.T) {
		dao, _ := setup(t)
		require.ErrorIs(t, dao.UpsertAssetProgress(context.Background(), "", &models.AssetProgress{}), utils.ErrPrincipal)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetAssetProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		lesson := &models.Lesson{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateLesson(ctx, lesson))

		asset := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Create asset progress
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 50},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_PROGRESS_TABLE_ID: assetProgress.ID})
		record, err := dao.GetAssetProgress(ctx, opts)
		require.NoError(t, err)
		require.Equal(t, assetProgress.ID, record.ID)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		record, err := dao.GetAssetProgress(ctx, nil)
		require.Nil(t, err)
		require.Nil(t, record)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListAssetProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		assetProgresses := []*models.AssetProgress{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))

			lesson := &models.Lesson{
				CourseID: course.ID,
				Title:    "Asset Group 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Module:   "Module 1",
			}
			require.NoError(t, dao.CreateLesson(ctx, lesson))

			asset := &models.Asset{
				CourseID: course.ID,
				LessonID: lesson.ID,
				Title:    "Asset 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Module:   "Module 1",
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course-%d/01 asset.mp4", i),
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "1234",
			}
			require.NoError(t, dao.CreateAsset(ctx, asset))

			// Set the asset progress to 5
			assetProgress := &models.AssetProgress{
				AssetID:           asset.ID,
				AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
			}
			assetProgresses = append(assetProgresses, assetProgress)
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))
			time.Sleep(1 * time.Millisecond)
		}

		records, err := dao.ListAssetProgress(ctx, nil)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, assetProgresses[i].ID, record.ID)
		}
	})

	t.Run("empty", func(t *testing.T) {
		dao, ctx := setup(t)

		records, err := dao.ListAssetProgress(ctx, nil)
		require.Nil(t, err)
		require.Empty(t, records)
	})

	t.Run("order by", func(t *testing.T) {
		dao, ctx := setup(t)

		assetProgresses := []*models.AssetProgress{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))

			lesson := &models.Lesson{
				CourseID: course.ID,
				Title:    "Asset Group 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Module:   "Module 1",
			}
			require.NoError(t, dao.CreateLesson(ctx, lesson))

			asset := &models.Asset{
				CourseID: course.ID,
				LessonID: lesson.ID,
				Title:    "Asset 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Module:   "Module 1",
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course-%d/01 asset.mp4", i),
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "1234",
			}
			require.NoError(t, dao.CreateAsset(ctx, asset))

			// Set the asset progress to 5
			assetProgress := &models.AssetProgress{
				AssetID:           asset.ID,
				AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
			}
			assetProgresses = append(assetProgresses, assetProgress)
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))
			time.Sleep(1 * time.Millisecond)
		}

		// Descending order by created_at
		opts := database.NewOptions().WithOrderBy(models.ASSET_PROGRESS_TABLE_CREATED_AT + " DESC")

		records, err := dao.ListAssetProgress(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, assetProgresses[2-i].ID, record.ID)
		}

		// Ascending order by created_at
		opts = database.NewOptions().WithOrderBy(models.ASSET_PROGRESS_TABLE_CREATED_AT + " ASC")

		records, err = dao.ListAssetProgress(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, assetProgresses[i].ID, record.ID)
		}
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		lesson := &models.Lesson{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateLesson(ctx, lesson))

		asset := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Create asset progress
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 50},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_PROGRESS_TABLE_ID: assetProgress.ID})
		records, err := dao.ListAssetProgress(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 1)
		require.Equal(t, assetProgress.ID, records[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, ctx := setup(t)

		assetProgresses := []*models.AssetProgress{}
		for i := range 17 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))

			lesson := &models.Lesson{
				CourseID: course.ID,
				Title:    "Asset Group 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Module:   "Module 1",
			}
			require.NoError(t, dao.CreateLesson(ctx, lesson))

			asset := &models.Asset{
				CourseID: course.ID,
				LessonID: lesson.ID,
				Title:    "Asset 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Module:   "Module 1",
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course-%d/01 asset.mp4", i),
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "1234",
			}
			require.NoError(t, dao.CreateAsset(ctx, asset))

			// Set the asset progress to 5
			assetProgress := &models.AssetProgress{
				AssetID:           asset.ID,
				AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
			}
			assetProgresses = append(assetProgresses, assetProgress)
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))
			time.Sleep(1 * time.Millisecond)
		}

		// First page with 10 records
		p := database.NewOptions().WithPagination(pagination.New(1, 10))
		records, err := dao.ListAssetProgress(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 10)
		require.Equal(t, assetProgresses[0].ID, records[0].ID)
		require.Equal(t, assetProgresses[9].ID, records[9].ID)

		// Second page with remaining 7 records
		p = database.NewOptions().WithPagination(pagination.New(2, 10))
		records, err = dao.ListAssetProgress(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 7)
		require.Equal(t, assetProgresses[10].ID, records[0].ID)
		require.Equal(t, assetProgresses[16].ID, records[6].ID)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteAssetProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		lesson := &models.Lesson{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateLesson(ctx, lesson))

		asset := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Create asset progress
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 50},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_PROGRESS_TABLE_ID: assetProgress.ID})
		require.Nil(t, dao.DeleteAssetProgress(ctx, opts))

		records, err := dao.ListAssetProgress(ctx, opts)
		require.NoError(t, err)
		require.Empty(t, records)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		lesson := &models.Lesson{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateLesson(ctx, lesson))

		asset := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Create asset progress
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 50},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_PROGRESS_TABLE_ID: "non-existent"})
		require.Nil(t, dao.DeleteAssetProgress(ctx, opts))

		records, err := dao.ListAssetProgress(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, assetProgress.ID, records[0].ID)
	})

	t.Run("missing where", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		lesson := &models.Lesson{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateLesson(ctx, lesson))

		asset := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Create asset progress
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 50},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		require.ErrorIs(t, dao.DeleteAssetProgress(ctx, nil), utils.ErrWhere)

		records, err := dao.ListAssetProgress(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, assetProgress.ID, records[0].ID)
	})

	t.Run("cascade", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		lesson := &models.Lesson{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateLesson(ctx, lesson))

		asset := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Asset 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
			Type:     *types.NewAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Create asset progress
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 50},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: course.ID})
		require.Nil(t, dao.DeleteCourses(ctx, dbOpts))

		records, err := dao.ListAssetProgress(ctx, nil)
		require.NoError(t, err)
		require.Empty(t, records)
	})
}

func Test_DeleteAssetProgressForCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		assetProgresses := []*models.AssetProgress{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))
			courses = append(courses, course)

			lesson := &models.Lesson{
				CourseID: course.ID,
				Title:    "Asset Group 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Module:   "Module 1",
			}
			require.NoError(t, dao.CreateLesson(ctx, lesson))

			asset := &models.Asset{
				CourseID: course.ID,
				LessonID: lesson.ID,
				Title:    "Asset 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Module:   "Module 1",
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("/course-%d/01 asset.mp4", i),
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "1234",
			}
			require.NoError(t, dao.CreateAsset(ctx, asset))

			// Set the asset progress to 5
			assetProgress := &models.AssetProgress{
				AssetID:           asset.ID,
				AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
			}
			assetProgresses = append(assetProgresses, assetProgress)
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))
			time.Sleep(1 * time.Millisecond)
		}

		principal, err := principalFromCtx(ctx)
		require.NoError(t, err)

		err = dao.DeleteAssetProgressForCourse(ctx, courses[1].ID, principal.UserID)
		require.NoError(t, err)

		records, err := dao.ListAssetProgress(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 2)
		require.Equal(t, assetProgresses[0].ID, records[0].ID)
		require.Equal(t, assetProgresses[2].ID, records[1].ID)
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		err := dao.DeleteAssetProgressForCourse(ctx, "", "")
		require.ErrorIs(t, err, utils.ErrCourseId)

		err = dao.DeleteAssetProgressForCourse(ctx, "course_id", "")
		require.ErrorIs(t, err, utils.ErrUserId)
	})
}
