package coursescan

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) (*CourseScan, context.Context, *[]*logger.Log) {
	t.Helper()

	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.TestWriteFn(&logs, &logsMux),
	})
	require.NoError(t, err, "Failed to initialize logger")

	appFs := appfs.New(afero.NewMemMapFs(), logger)

	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: "./oc_data",
		AppFs:   appFs,
		Testing: true,
	})

	require.NoError(t, err)
	require.NotNil(t, dbManager)

	// Create a mock FFmpeg for testing
	ffmpeg, err := media.NewFFmpeg()
	if err != nil {
		// If FFmpeg is not available, skip the test
		t.Skip("FFmpeg not available; skipping test")
	}

	courseScan := New(&CourseScanConfig{
		Db:     dbManager.DataDb,
		AppFs:  appFs,
		Logger: logger,
		FFmpeg: ffmpeg,
	})

	// Create a user for the context
	user := &models.User{
		Username:     "test-user",
		DisplayName:  "Test User",
		PasswordHash: "test-password",
		Role:         types.UserRoleAdmin,
	}
	require.NoError(t, courseScan.dao.CreateUser(context.Background(), user))

	principal := types.Principal{
		UserID: user.ID,
		Role:   user.Role,
	}
	ctx := context.WithValue(context.Background(), types.PrincipalContextKey, principal)

	return courseScan, ctx, &logs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func intPtr(i int) *int {
	return &i
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		course1 := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course1))

		scan1, err := scanner.Add(ctx, course1.ID)
		require.NoError(t, err)
		require.Equal(t, course1.ID, scan1.CourseID)

		course2 := &models.Course{Title: "Course 2", Path: "/course-2"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course2))

		scan2, err := scanner.Add(ctx, course2.ID)
		require.NoError(t, err)
		require.Equal(t, course2.ID, scan2.CourseID)
	})

	t.Run("duplicate", func(t *testing.T) {
		scanner, ctx, logs := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		first, err := scanner.Add(ctx, course.ID)
		require.NoError(t, err)
		require.Equal(t, course.ID, first.CourseID)

		// Add again
		second, err := scanner.Add(ctx, course.ID)
		require.NoError(t, err)
		require.Equal(t, second.ID, first.ID)
		require.NotEmpty(t, *logs)
		require.Equal(t, "Scan job already exists", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("invalid course", func(t *testing.T) {
		scanner, ctx, _ := setup(t)

		scan, err := scanner.Add(ctx, "1234")
		require.ErrorIs(t, err, utils.ErrCourseNotFound)
		require.Nil(t, scan)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanner_Worker(t *testing.T) {
	t.Run("jobs", func(t *testing.T) {
		scanner, ctx, logs := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course-%d", i)}
			require.NoError(t, scanner.dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		var processingDone = make(chan bool, 1)
		go scanner.Worker(ctx, func(context.Context, *CourseScan, *models.Scan) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}, processingDone)

		// Add the courses
		for i := range 3 {
			scan, err := scanner.Add(ctx, courses[i].ID)
			require.NoError(t, err)
			require.Equal(t, scan.CourseID, courses[i].ID)
		}

		<-processingDone

		count, err := scanner.dao.CountScans(ctx, nil)
		require.NoError(t, err)
		require.Zero(t, count)

		require.NotEmpty(t, *logs)
		require.Equal(t, "Finished processing all scan jobs", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)

		// Add the first 2 courses (again)
		for i := range 2 {
			scan, err := scanner.Add(ctx, courses[i].ID)
			require.NoError(t, err)
			require.Equal(t, scan.CourseID, courses[i].ID)
		}

		<-processingDone

		count, err = scanner.dao.CountScans(ctx, nil)
		require.NoError(t, err)
		require.Zero(t, count)

		require.NotEmpty(t, *logs)
		require.Equal(t, "Finished processing all scan jobs", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})

	t.Run("error processing", func(t *testing.T) {
		scanner, ctx, logs := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, scanner.dao.CreateCourse(ctx, course))

		var processingDone = make(chan bool, 1)
		go scanner.Worker(ctx, func(context.Context, *CourseScan, *models.Scan) error {
			time.Sleep(1 * time.Millisecond)
			return errors.New("processing error")
		}, processingDone)

		scan, err := scanner.Add(ctx, course.ID)
		require.NoError(t, err)
		require.Equal(t, scan.CourseID, course.ID)

		<-processingDone

		require.NotEmpty(t, *logs)
		require.Greater(t, len(*logs), 2)
		require.Equal(t, "Failed to process scan job", (*logs)[len(*logs)-2].Message)
		require.Equal(t, slog.LevelError, (*logs)[len(*logs)-2].Level)
		require.Equal(t, "Finished processing all scan jobs", (*logs)[len(*logs)-1].Message)
		require.Equal(t, slog.LevelDebug, (*logs)[len(*logs)-1].Level)
	})
}
