package dao

import (
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
	t.Run("success (video)", func(t *testing.T) {
		dao, ctx := setup(t)

		// Course + lesson
		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))
		lesson := &models.Lesson{
			CourseID: course.ID,
			Title:    "Lesson 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateLesson(ctx, lesson))

		// Video asset
		video := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Video 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Type:     types.MustAsset("mp4"),
			Path:     "/course-1/01-video.mp4",
			FileSize: 111,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "v1",
			Weight:   1,
		}
		require.NoError(t, dao.CreateAsset(ctx, video))

		// Video metadata (100s duration)
		require.NoError(t, dao.CreateAssetMetadata(ctx, &models.AssetMetadata{
			AssetID: video.ID,
			VideoMetadata: &models.VideoMetadata{
				DurationSec: 100,
				Container:   "mp4",
				MIMEType:    "video/mp4",
				VideoCodec:  "h264",
				Width:       1280,
				Height:      720,
				FPSNum:      30,
				FPSDen:      1,
			},
		}))

		// No progress
		opts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: course.ID})
		cp, err := dao.GetCourseProgress(ctx, opts)
		require.NoError(t, err)
		require.Nil(t, cp)

		// Set the asset progress to 50% complete
		ap := &models.AssetProgress{AssetID: video.ID, Position: 50}
		require.NoError(t, dao.UpsertAssetProgress(ctx, ap))

		cp, err = dao.GetCourseProgress(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, cp)
		require.Equal(t, course.ID, cp.CourseID)
		require.True(t, cp.Started)
		require.False(t, cp.StartedAt.IsZero())
		require.Equal(t, 50, cp.Percent)
		require.True(t, cp.CompletedAt.IsZero())

		// Update position to 100 (100%) and set completed
		ap.Completed = true
		ap.Position = 100
		require.NoError(t, dao.UpsertAssetProgress(ctx, ap))

		cp2, err := dao.GetCourseProgress(ctx, opts)

		require.NoError(t, err)
		require.NotNil(t, cp2)
		require.Equal(t, 100, cp2.Percent)
		require.False(t, cp2.CompletedAt.IsZero())
		require.False(t, cp2.StartedAt.IsZero())
	})

	t.Run("success (multiple assets)", func(t *testing.T) {
		dao, ctx := setup(t)

		// Course + lesson
		course := &models.Course{Title: "Course M", Path: "/course-m"}
		require.NoError(t, dao.CreateCourse(ctx, course))
		lesson := &models.Lesson{
			CourseID: course.ID,
			Title:    "Lesson M",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
		}
		require.NoError(t, dao.CreateLesson(ctx, lesson))

		// Video A
		vA := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Video A",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Type:     types.MustAsset("mp4"),
			Path:     "/course-m/01-a.mp4",
			FileSize: 100,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "va",
			Weight:   1,
		}
		require.NoError(t, dao.CreateAsset(ctx, vA))
		require.NoError(t, dao.CreateAssetMetadata(ctx, &models.AssetMetadata{
			AssetID: vA.ID,
			VideoMetadata: &models.VideoMetadata{
				DurationSec: 100,
				Container:   "mp4",
				MIMEType:    "video/mp4",
				VideoCodec:  "h264",
				Width:       1280,
				Height:      720,
				FPSNum:      30,
				FPSDen:      1,
			},
		}))

		// Video B
		vB := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Video B",
			Prefix:   sql.NullInt16{Int16: 2, Valid: true},
			Type:     types.MustAsset("mp4"),
			Path:     "/course-m/02-b.mp4",
			FileSize: 200,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "vb",
			Weight:   1,
		}
		require.NoError(t, dao.CreateAsset(ctx, vB))
		require.NoError(t, dao.CreateAssetMetadata(ctx, &models.AssetMetadata{
			AssetID: vB.ID,
			VideoMetadata: &models.VideoMetadata{
				DurationSec: 200,
				Container:   "mp4",
				MIMEType:    "video/mp4",
				VideoCodec:  "h264",
				Width:       1920,
				Height:      1080,
				FPSNum:      30,
				FPSDen:      1,
			},
		}))

		// Document C
		doc := &models.Asset{
			CourseID: course.ID,
			LessonID: lesson.ID,
			Title:    "Doc C",
			Prefix:   sql.NullInt16{Int16: 3, Valid: true},
			Type:     types.MustAsset("md"),
			Path:     "/course-m/03-c.md",
			FileSize: 10,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "dc",
			Weight:   1,
		}
		require.NoError(t, dao.CreateAsset(ctx, doc))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: course.ID})

		// A @ 50s (0.5), B @ 0 (0.0), Doc not completed (0.0):
		// percent = round(100 * (0.5 + 0 + 0) / 3) = 17
		require.NoError(t, dao.UpsertAssetProgress(ctx, &models.AssetProgress{
			AssetID:  vA.ID,
			Position: 50,
		}))
		cp, err := dao.GetCourseProgress(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, cp)
		require.Equal(t, 17, cp.Percent)
		require.True(t, cp.Started)
		require.False(t, cp.StartedAt.IsZero())

		// Now B @ 100/200 (0.5): avg = (0.5 + 0.5 + 0) / 3 = 0.333.. -> 33
		require.NoError(t, dao.UpsertAssetProgress(ctx, &models.AssetProgress{
			AssetID:  vB.ID,
			Position: 100,
		}))
		cp, err = dao.GetCourseProgress(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, cp)
		require.Equal(t, 33, cp.Percent)

		// Mark doc completed (1.0): avg = (0.5 + 0.5 + 1.0) / 3 = 0.666.. -> 67
		require.NoError(t, dao.UpsertAssetProgress(ctx, &models.AssetProgress{
			AssetID:   doc.ID,
			Completed: true,
		}))
		cp, err = dao.GetCourseProgress(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, cp)
		require.Equal(t, 67, cp.Percent)
		require.True(t, cp.CompletedAt.IsZero()) // not 100 yet

		// Complete both videos: (1 + 1 + 1) / 3 = 1.0 -> 100, completed_at set
		require.NoError(t, dao.UpsertAssetProgress(ctx, &models.AssetProgress{
			AssetID:   vA.ID,
			Completed: true,
			Position:  100,
		}))
		require.NoError(t, dao.UpsertAssetProgress(ctx, &models.AssetProgress{
			AssetID:   vB.ID,
			Completed: true,
			Position:  200,
		}))
		cp, err = dao.GetCourseProgress(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, cp)
		require.Equal(t, 100, cp.Percent)
		require.False(t, cp.CompletedAt.IsZero())
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
				Type:     types.MustAsset("mp4"),
				Path:     fmt.Sprintf("/course-%d/01 asset.mp4", i),
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "1234",
			}
			require.NoError(t, dao.CreateAsset(ctx, asset))

			// Set the asset progress to 5
			assetProgress := &models.AssetProgress{
				AssetID:  asset.ID,
				Position: 5,
			}
			require.NoError(t, dao.UpsertAssetProgress(ctx, assetProgress))
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
				Type:     types.MustAsset("mp4"),
				Path:     fmt.Sprintf("/course-%d/01 asset.mp4", i),
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "1234",
			}
			require.NoError(t, dao.CreateAsset(ctx, asset))

			// Set the asset progress to 5
			assetProgress := &models.AssetProgress{
				AssetID:  asset.ID,
				Position: 5,
			}
			require.NoError(t, dao.UpsertAssetProgress(ctx, assetProgress))
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
			Type:     types.MustAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:  asset.ID,
			Position: 5,
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, assetProgress))

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
				Type:     types.MustAsset("mp4"),
				Path:     fmt.Sprintf("/course-%d/01 asset.mp4", i),
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "1234",
			}
			require.NoError(t, dao.CreateAsset(ctx, asset))

			// Set the asset progress to 5
			assetProgress := &models.AssetProgress{
				AssetID:  asset.ID,
				Position: 5,
			}
			require.NoError(t, dao.UpsertAssetProgress(ctx, assetProgress))
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
			Type:     types.MustAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:  asset.ID,
			Position: 5,
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, assetProgress))

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
			Type:     types.MustAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:  asset.ID,
			Position: 5,
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, assetProgress))

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
			Type:     types.MustAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:  asset.ID,
			Position: 5,
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, assetProgress))

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
			Type:     types.MustAsset("mp4"),
			Path:     "/course-1/01 asset.mp4",
			FileSize: 1024,
			ModTime:  time.Now().Format(time.RFC3339Nano),
			Hash:     "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Set the asset progress to 5
		assetProgress := &models.AssetProgress{
			AssetID:  asset.ID,
			Position: 5,
		}
		require.NoError(t, dao.UpsertAssetProgress(ctx, assetProgress))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: course.ID})
		require.Nil(t, dao.DeleteCourses(ctx, dbOpts))

		records, err := dao.ListCourseProgress(ctx, nil)
		require.NoError(t, err)
		require.Empty(t, records)
	})
}
