package daos

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func CourseProgressSetup(t *testing.T) (*appFs.AppFs, *CourseProgressDao, database.Database) {
	appFs, db := setup(t)
	cpDao := NewCourseProgressDao(db)
	return appFs, cpDao, db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseProgress_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, dao, db := CourseProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		cp, err := dao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.False(t, cp.Started)
		assert.True(t, cp.StartedAt.IsZero())
		assert.Zero(t, cp.Percent)
		assert.True(t, cp.CompletedAt.IsZero())
		assert.False(t, cp.CreatedAt.IsZero())
		assert.False(t, cp.UpdatedAt.IsZero())
	})

	t.Run("duplicate course id", func(t *testing.T) {
		_, dao, db := CourseProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		cp, err := dao.Get(workingData[0].ID)
		require.Nil(t, err)

		err = dao.Create(cp)
		require.ErrorContains(t, err, fmt.Sprintf("UNIQUE constraint failed: %s.course_id", dao.table))
	})

	t.Run("constraint errors", func(t *testing.T) {
		_, dao, db := CourseProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)

		// Delete the courses_progress row using squirrel
		query, args, _ := squirrel.StatementBuilder.Delete(dao.table).Where(squirrel.Eq{"course_id": workingData[0].ID}).ToSql()
		_, err := db.Exec(query, args...)
		require.Nil(t, err)

		// Course ID
		cp := &models.CourseProgress{}
		require.ErrorContains(t, dao.Create(cp), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.table))
		cp.CourseID = ""
		require.ErrorContains(t, dao.Create(cp), fmt.Sprintf("NOT NULL constraint failed: %s.course_id", dao.table))
		cp.CourseID = "1234"

		// Invalid Course ID
		require.ErrorContains(t, dao.Create(cp), "FOREIGN KEY constraint failed")
		cp.CourseID = workingData[0].ID

		// Success
		require.Nil(t, dao.Create(cp))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseProgress_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		_, dao, db := CourseProgressSetup(t)

		workingData := NewTestData(t, db, 1, true, 0, 0)

		cp, err := dao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.Equal(t, workingData[0].ID, cp.CourseID)
	})

	t.Run("not found", func(t *testing.T) {
		_, dao, _ := CourseProgressSetup(t)

		cp, err := dao.Get("1234")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, cp)
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, _ := CourseProgressSetup(t)

		cp, err := dao.Get("")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, cp)
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := CourseProgressSetup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		_, err = dao.Get("1234")
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseProgress_Update(t *testing.T) {
	t.Run("status", func(t *testing.T) {
		_, dao, db := CourseProgressSetup(t)

		apDao := NewAssetProgressDao(db)

		workingData := NewTestData(t, db, 1, false, 2, 0)

		// There are 2 assets. Create an asset progress for the first asset
		aps := []*models.AssetProgress{}
		for _, a := range workingData[0].Assets {
			if a.ID != workingData[0].Assets[1].ID {
				ap := newTestAssetsProgress(t, db, a.ID, workingData[0].ID)
				aps = append(aps, ap)
			}
		}

		// Ensure the percent is 0, started is false, and the started_at and completed_at are not set
		origCp, err := dao.Get(workingData[0].ID)
		require.Nil(t, err)
		assert.False(t, origCp.Started)
		assert.True(t, origCp.StartedAt.IsZero())
		require.Zero(t, origCp.Percent)
		assert.True(t, origCp.CompletedAt.IsZero())

		// ----------------------------
		// Set the first asset to completed
		// ----------------------------
		time.Sleep(1 * time.Millisecond)
		aps[0].Completed = true
		require.Nil(t, apDao.Update(aps[0]))

		// Check the percent is 50, started is true, started_at is set and completed_at is not set
		updatedCp1, err := dao.Get(origCp.CourseID)
		require.Nil(t, err)
		assert.True(t, updatedCp1.Started)
		assert.False(t, updatedCp1.StartedAt.IsZero())
		require.Equal(t, 50, updatedCp1.Percent)
		assert.True(t, updatedCp1.CompletedAt.IsZero())

		// ----------------------------
		// Set the second asset to completed
		// ----------------------------
		ap2 := newTestAssetsProgress(t, db, workingData[0].Assets[1].ID, workingData[0].ID)
		ap2.Completed = true
		require.Nil(t, apDao.Update(ap2))

		// Check the percent is 100, started is true, and started_at and completed_at are set
		updatedCp2, err := dao.Get(origCp.CourseID)
		require.Nil(t, err)
		assert.True(t, updatedCp2.Started)
		assert.False(t, updatedCp2.StartedAt.IsZero())
		assert.Equal(t, updatedCp2.StartedAt.String(), updatedCp1.StartedAt.String())
		require.Equal(t, 100, updatedCp2.Percent)
		assert.False(t, updatedCp2.CompletedAt.IsZero())

		// ----------------------------
		// Set the second asset as uncompleted
		// ----------------------------
		ap2.Completed = false
		require.Nil(t, apDao.Update(ap2))

		// Check the percent is 50, started is true, started_at is set and completed_at is not set
		updatedCp3, err := dao.Get(origCp.CourseID)
		require.Nil(t, err)
		assert.True(t, updatedCp3.Started)
		assert.False(t, updatedCp3.StartedAt.IsZero())
		assert.Equal(t, updatedCp3.StartedAt.String(), updatedCp2.StartedAt.String())
		require.Equal(t, 50, updatedCp3.Percent)
		assert.True(t, updatedCp3.CompletedAt.IsZero())

		// ----------------------------
		// Set the first asset as uncompleted
		// ----------------------------
		time.Sleep(1 * time.Millisecond)
		aps[0].Completed = false
		require.Nil(t, apDao.Update(aps[0]))

		// Check the percent is 0, started is false and started_at and completed_at are not set
		updatedCp4, err := dao.Get(origCp.CourseID)
		require.Nil(t, err)
		assert.False(t, updatedCp4.Started)
		assert.True(t, updatedCp4.StartedAt.IsZero())
		require.Zero(t, updatedCp4.Percent)
		assert.True(t, updatedCp4.CompletedAt.IsZero())
	})

	t.Run("empty id", func(t *testing.T) {
		_, dao, db := CourseProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		origCp, err := dao.Get(workingData[0].ID)
		require.Nil(t, err)

		origCp.CourseID = ""

		err = dao.Refresh(origCp.CourseID)
		assert.EqualError(t, err, "id cannot be empty")
	})

	t.Run("db error", func(t *testing.T) {
		_, dao, db := CourseProgressSetup(t)

		workingData := NewTestData(t, db, 1, false, 0, 0)
		origCp, err := dao.Get(workingData[0].ID)
		require.Nil(t, err)

		_, err = db.Exec("DROP TABLE IF EXISTS " + dao.table)
		require.Nil(t, err)

		err = dao.Refresh(origCp.CourseID)
		require.ErrorContains(t, err, "no such table: "+dao.table)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCourseProgress_DeleteCascade(t *testing.T) {
	_, dao, db := CourseProgressSetup(t)

	workingData := NewTestData(t, db, 1, true, 0, 0)

	// Delete the course
	courseDao := NewCourseDao(db)
	err := courseDao.Delete(workingData[0].ID)
	require.Nil(t, err)

	// Check the course progress was deleted
	cp, err := dao.Get(workingData[0].ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, cp)
}
