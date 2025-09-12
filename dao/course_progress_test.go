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

func Test_GetCourseProgress(t *testing.T) {
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

		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: course.ID})
		record, err := dao.GetCourseProgress(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, course.ID, record.CourseID)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		record, err := dao.GetCourseProgress(ctx, nil)
		require.Nil(t, err)
		require.Nil(t, record)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListCourseProgress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}

		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			courses = append(courses, course)
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
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))
			time.Sleep(1 * time.Millisecond)
		}

		records, err := dao.ListCourseProgress(ctx, nil)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, courses[i].ID, record.CourseID)
		}
	})

	t.Run("empty", func(t *testing.T) {
		dao, ctx := setup(t)

		records, err := dao.ListCourseProgress(ctx, nil)
		require.Nil(t, err)
		require.Empty(t, records)
	})

	t.Run("order by", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			courses = append(courses, course)
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
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))
			time.Sleep(1 * time.Millisecond)
		}

		// Descending order by created_at
		opts := database.NewOptions().WithOrderBy(models.COURSE_PROGRESS_TABLE_CREATED_AT + " DESC")

		records, err := dao.ListCourseProgress(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, courses[2-i].ID, record.CourseID)
		}

		// Ascending order by created_at
		opts = database.NewOptions().WithOrderBy(models.COURSE_PROGRESS_TABLE_CREATED_AT + " ASC")

		records, err = dao.ListCourseProgress(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, courses[i].ID, record.CourseID)
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

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: course.ID})
		records, err := dao.ListCourseProgress(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 1)
		require.Equal(t, course.ID, records[0].CourseID)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 17 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			courses = append(courses, course)
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
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))
			time.Sleep(1 * time.Millisecond)
		}

		// First page with 10 records
		p := database.NewOptions().WithPagination(pagination.New(1, 10))
		records, err := dao.ListCourseProgress(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 10)
		require.Equal(t, courses[0].ID, records[0].CourseID)
		require.Equal(t, courses[9].ID, records[9].CourseID)

		// Second page with remaining 7 records
		p = database.NewOptions().WithPagination(pagination.New(2, 10))
		records, err = dao.ListCourseProgress(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 7)
		require.Equal(t, courses[10].ID, records[0].CourseID)
		require.Equal(t, courses[16].ID, records[6].CourseID)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TODO Add test to check when the course progress is x but then an asset is deleted
func Test_SyncCourseProgress(t *testing.T) {
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

		asset1 := &models.Asset{
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
		require.NoError(t, dao.CreateAsset(ctx, asset1))

		// Create video metadata
		// TODO fix
		// videoMetadata := &models.VideoMetadata{
		// 	AssetID: asset1.ID,
		// 	VideoMetadataInfo: models.VideoMetadataInfo{
		// 		Duration:   10,
		// 		Width:      1920,
		// 		Height:     1080,
		// 		Resolution: "1080p",
		// 		Codec:      "h264",
		// 	},
		// }
		// require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata))

		principal, err := principalFromCtx(ctx)
		require.NoError(t, err)

		dbOpts := database.NewOptions().
			WithWhere(squirrel.And{
				squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: course.ID},
				squirrel.Eq{models.COURSE_PROGRESS_TABLE_USER_ID: principal.UserID},
			})

		assetProgress := &models.AssetProgress{}

		// Set the asset progress to 5, which is 50% of the video duration
		{
			assetProgress = &models.AssetProgress{
				AssetID:           asset1.ID,
				AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
			}
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

			record, err := dao.GetCourseProgress(ctx, dbOpts)
			require.NoError(t, err)
			require.True(t, record.Started)
			require.False(t, record.StartedAt.IsZero())
			require.Equal(t, 50, record.Percent)
			require.True(t, record.CompletedAt.IsZero())
		}

		// Clear the asset progress
		{
			assetProgress.VideoPos = 0
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

			record, err := dao.GetCourseProgress(ctx, dbOpts)
			require.NoError(t, err)
			require.False(t, record.Started)
			require.True(t, record.StartedAt.IsZero())
			require.Zero(t, record.Percent)
			require.True(t, record.CompletedAt.IsZero())
		}

		// Set asset as completed
		{
			assetProgress.Completed = true
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

			record, err := dao.GetCourseProgress(ctx, dbOpts)
			require.NoError(t, err)
			require.True(t, record.Started)
			require.False(t, record.StartedAt.IsZero())
			require.Equal(t, 100, record.Percent)
			require.False(t, record.CompletedAt.IsZero())
		}

		// Add another asset
		{
			asset2 := &models.Asset{
				CourseID: course.ID,
				LessonID: lesson.ID,
				Title:    "Asset 2",
				Prefix:   sql.NullInt16{Int16: 2, Valid: true},
				Module:   "Module 2",
				Type:     *types.NewAsset("mp4"),
				Path:     "/course-1/02 asset.mp4",
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "5678",
			}
			require.NoError(t, dao.CreateAsset(ctx, asset2))

			// TODO fix
			// videoMetadata2 := &models.VideoMetadata{
			// 	AssetID: asset2.ID,
			// 	VideoMetadataInfo: models.VideoMetadataInfo{
			// 		Duration:   10,
			// 		Width:      1920,
			// 		Height:     1080,
			// 		Resolution: "1080p",
			// 		Codec:      "h264",
			// 	},
			// }
			// require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata2))

			// Set asset2 progress
			assetProgress2 := &models.AssetProgress{
				AssetID:           asset2.ID,
				AssetProgressInfo: models.AssetProgressInfo{VideoPos: 0},
			}
			require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress2))

			record, err := dao.GetCourseProgress(ctx, dbOpts)
			require.NoError(t, err)
			require.True(t, record.Started)
			require.False(t, record.StartedAt.IsZero())
			require.Equal(t, 50, record.Percent)
			require.True(t, record.CompletedAt.IsZero())
		}
	})

	t.Run("invalid course id", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.SyncCourseProgress(ctx, ""), utils.ErrCourseId)

		require.ErrorIs(t, dao.SyncCourseProgress(ctx, "invalid"), utils.ErrCourseId)
	})

	t.Run("missing principal", func(t *testing.T) {
		dao, _ := setup(t)
		require.ErrorIs(t, dao.SyncCourseProgress(context.Background(), "1234"), utils.ErrPrincipal)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteCourseProgress(t *testing.T) {
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

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: course.ID})
		require.Nil(t, dao.DeleteCourseProgress(ctx, opts))

		records, err := dao.ListCourseProgress(ctx, opts)
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

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_PROGRESS_TABLE_ID: "non-existent"})
		require.Nil(t, dao.DeleteCourseProgress(ctx, opts))

		records, err := dao.ListCourseProgress(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, course.ID, records[0].CourseID)
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

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		require.ErrorIs(t, dao.DeleteCourseProgress(ctx, nil), utils.ErrWhere)

		records, err := dao.ListCourseProgress(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, course.ID, records[0].CourseID)
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

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:           asset.ID,
			AssetProgressInfo: models.AssetProgressInfo{VideoPos: 5},
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, course.ID, assetProgress))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: course.ID})
		require.Nil(t, dao.DeleteCourses(ctx, dbOpts))

		records, err := dao.ListCourseProgress(ctx, nil)
		require.NoError(t, err)
		require.Empty(t, records)
	})
}
