package cron

import (
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseAvailability_Run(t *testing.T) {
	t.Run("update", func(t *testing.T) {
		db, appFs, ctx, logger := setup(t)

		dao := dao.New(db)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course-%d", i), Available: false}
			require.NoError(t, dao.CreateCourse(ctx, course))
			courses = append(courses, course)
		}

		ca := &courseAvailability{
			db:        db,
			dao:       dao,
			appFs:     appFs,
			logger:    logger,
			batchSize: 2,
		}

		err := ca.run()
		require.NoError(t, err)

		for _, course := range courses {
			require.Nil(t, appFs.Fs.MkdirAll(course.Path, 0755))
		}

		err = ca.run()
		require.NoError(t, err)

		for _, course := range courses {
			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: course.ID})
			course, err := dao.GetCourse(ctx, dbOpts)
			require.NoError(t, err)
			require.True(t, course.Available)
		}
	})

	t.Run("stat error", func(t *testing.T) {
		db, _, ctx, logger := setup(t)

		dao := dao.New(db)

		course := &models.Course{Title: "course 1", Path: "/course-1", Available: false}
		require.NoError(t, dao.CreateCourse(ctx, course))

		fsWithError := &mocks.MockFsWithError{
			Fs:          afero.NewMemMapFs(),
			ErrToReturn: fmt.Errorf("stat error"),
		}

		ca := &courseAvailability{
			db:        db,
			dao:       dao,
			appFs:     appfs.New(fsWithError),
			logger:    logger,
			batchSize: 1,
		}

		err := ca.run()
		require.Equal(t, fmt.Errorf("stat error"), err)

		// Note: Log assertions removed as we no longer have access to log entries in the new logger system
	})

	t.Run("db error", func(t *testing.T) {
		db, appFs, _, logger := setup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + models.COURSE_TABLE)
		require.NoError(t, err)

		ca := &courseAvailability{
			db:        db,
			dao:       dao.New(db),
			appFs:     appFs,
			logger:    logger,
			batchSize: 1,
		}

		err = ca.run()
		require.ErrorContains(t, err, "no such table: "+models.COURSE_TABLE)

		// Note: Log assertions removed as we no longer have access to log entries in the new logger system
	})
}
