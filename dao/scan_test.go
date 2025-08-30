package dao

import (
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

func Test_CreateScan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1", Available: true, CardPath: "/course-1/card-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))
	})

	t.Run("duplicate", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1", Available: true, CardPath: "/course-1/card-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		require.ErrorContains(t, dao.CreateScan(ctx, scan), "UNIQUE constraint failed: "+models.SCAN_TABLE_COURSE_ID)
	})

	t.Run("nil pointer", func(t *testing.T) {
		dao, ctx := setup(t)

		require.ErrorIs(t, dao.CreateScan(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("invalid course ID", func(t *testing.T) {
		dao, ctx := setup(t)

		scan := &models.Scan{CourseID: "invalid"}
		require.ErrorContains(t, dao.CreateScan(ctx, scan), "FOREIGN KEY constraint failed")

		scan = &models.Scan{CourseID: ""}
		require.ErrorContains(t, dao.CreateScan(ctx, scan), "FOREIGN KEY constraint failed")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetScan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_ID: scan.ID})
		record, err := dao.GetScan(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, scan.ID, record.ID)
		require.Equal(t, course.Path, record.CoursePath)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		record, err := dao.GetScan(ctx, nil)
		require.Nil(t, err)
		require.Nil(t, record)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListScans(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		scans := []*models.Scan{}

		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			courses = append(courses, course)
			require.NoError(t, dao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			scans = append(scans, scan)
			require.NoError(t, dao.CreateScan(ctx, scan))
			time.Sleep(1 * time.Millisecond)
		}

		records, err := dao.ListScans(ctx, nil)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, scans[i].ID, record.ID)
			require.Equal(t, courses[i].Path, record.CoursePath)
		}
	})

	t.Run("empty", func(t *testing.T) {
		dao, ctx := setup(t)

		records, err := dao.ListScans(ctx, nil)
		require.Nil(t, err)
		require.Empty(t, records)
	})

	t.Run("order by", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		scans := []*models.Scan{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			courses = append(courses, course)
			require.NoError(t, dao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			scans = append(scans, scan)
			require.NoError(t, dao.CreateScan(ctx, scan))
			time.Sleep(1 * time.Millisecond)
		}

		// Descending order by created_at
		opts := database.NewOptions().WithOrderBy(models.SCAN_TABLE_CREATED_AT + " DESC")

		records, err := dao.ListScans(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, scans[2-i].ID, record.ID)
			require.Equal(t, courses[2-i].Path, record.CoursePath)
		}

		// Ascending order by created_at
		opts = database.NewOptions().WithOrderBy(models.SCAN_TABLE_CREATED_AT + " ASC")

		records, err = dao.ListScans(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, scans[i].ID, record.ID)
			require.Equal(t, courses[i].Path, record.CoursePath)
		}
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_ID: scan.ID})
		records, err := dao.ListScans(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 1)
		require.Equal(t, scan.ID, records[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, ctx := setup(t)

		scans := []*models.Scan{}
		for i := range 17 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			scans = append(scans, scan)
			require.NoError(t, dao.CreateScan(ctx, scan))
			time.Sleep(1 * time.Millisecond)
		}

		// First page with 10 records
		p := database.NewOptions().WithPagination(pagination.New(1, 10))
		records, err := dao.ListScans(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 10)
		require.Equal(t, scans[0].ID, records[0].ID)
		require.Equal(t, scans[9].ID, records[9].ID)

		// Second page with remaining 7 records
		p = database.NewOptions().WithPagination(pagination.New(2, 10))
		records, err = dao.ListScans(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 7)
		require.Equal(t, scans[10].ID, records[0].ID)
		require.Equal(t, scans[16].ID, records[6].ID)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateScan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		originalScan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, originalScan))

		time.Sleep(1 * time.Millisecond)

		updatedScan := &models.Scan{
			Base:     originalScan.Base,
			CourseID: "1234",                          // Immutable
			Status:   types.NewScanStatusProcessing(), // Mutable
		}
		require.NoError(t, dao.UpdateScan(ctx, updatedScan))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_ID: originalScan.ID})
		record, err := dao.GetScan(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, originalScan.ID, record.ID)                     // No change
		require.Equal(t, originalScan.CourseID, record.CourseID)         // No change
		require.True(t, record.CreatedAt.Equal(originalScan.CreatedAt))  // No change
		require.False(t, record.Status.IsWaiting())                      // Changed
		require.False(t, record.UpdatedAt.Equal(originalScan.UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		// Empty ID
		scan.ID = ""
		require.ErrorIs(t, dao.UpdateScan(ctx, scan), utils.ErrId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateScan(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DeleteScans(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course", Path: "/course"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_ID: scan.ID})
		require.Nil(t, dao.DeleteScans(ctx, opts))

		records, err := dao.ListScans(ctx, opts)
		require.NoError(t, err)
		require.Empty(t, records)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course", Path: "/course"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_ID: "non-existent"})
		require.Nil(t, dao.DeleteScans(ctx, opts))

		records, err := dao.ListScans(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, scan.ID, records[0].ID)
	})

	t.Run("missing where", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course", Path: "/course"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		require.ErrorIs(t, dao.DeleteScans(ctx, nil), utils.ErrWhere)

		records, err := dao.ListScans(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, scan.ID, records[0].ID)
	})

	t.Run("cascade", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course", Path: "/course"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, dao.CreateScan(ctx, scan))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: course.ID})
		require.Nil(t, dao.DeleteCourses(ctx, dbOpts))

		records, err := dao.ListScans(ctx, nil)
		require.NoError(t, err)
		require.Empty(t, records)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NextWaitingScan(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		dao, ctx := setup(t)

		scans := []*models.Scan{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("Course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			require.NoError(t, dao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			require.NoError(t, dao.CreateScan(ctx, scan))
			scans = append(scans, scan)

			time.Sleep(1 * time.Millisecond)
		}

		record, err := dao.NextWaitingScan(ctx)
		require.NoError(t, err)
		require.Equal(t, scans[0].ID, record.ID)
	})

	t.Run("next", func(t *testing.T) {
		dao, ctx := setup(t)

		scans := []*models.Scan{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			require.NoError(t, dao.CreateScan(ctx, scan))
			scans = append(scans, scan)

			time.Sleep(1 * time.Millisecond)
		}

		scans[0].Status = types.NewScanStatusProcessing()
		require.NoError(t, dao.UpdateScan(ctx, scans[0]))

		record, err := dao.NextWaitingScan(ctx)
		require.NoError(t, err)
		require.Equal(t, scans[1].ID, record.ID)
	})

	t.Run("empty", func(t *testing.T) {
		dao, ctx := setup(t)

		record, err := dao.NextWaitingScan(ctx)
		require.Nil(t, err)
		require.Nil(t, record)
	})
}
