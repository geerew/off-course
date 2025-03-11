package dao

import (
	"context"
	"database/sql"
	"math"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourseProgress creates a course progress
func (dao *DAO) CreateCourseProgress(ctx context.Context, courseProgress *models.CourseProgress) error {
	if courseProgress == nil {
		return utils.ErrNilPtr
	}

	return dao.Create(ctx, courseProgress)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourseProgress update a course progress
func (dao *DAO) UpdateCourseProgress(ctx context.Context, courseProgress *models.CourseProgress) error {
	if courseProgress == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, courseProgress)
	return err
}

// Refresh refreshes the current course progress for the given ID
//
// It calculates the number of assets, number of completed assets and number of started video assets,
// then calculates the percent complete and whether the course has been started
//
// Based upon this calculation,
//   - If the course has been started and `started_at` is null, `started_at` will be set to NOW
//   - If the course is not started, `started_at` is set to null
//   - If the course is complete and `completed_at` is null, `completed_at` is set to NOW
//   - If the course is not complete, `completed_at` is set to null
func (dao *DAO) RefreshCourseProgress(ctx context.Context, courseID string) error {
	if courseID == "" {
		return utils.ErrInvalidId
	}

	// Count the number of assets, number of completed assets and number of video assets started for
	// this course
	query, args, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(
			"COUNT(DISTINCT "+models.ASSET_TABLE_ID+") AS total_count",
			"SUM(CASE WHEN "+models.ASSET_PROGRESS_TABLE_COMPLETED+" THEN 1 ELSE 0 END) AS completed_count",
			"SUM(CASE WHEN "+models.ASSET_PROGRESS_TABLE_VIDEO_POS+" > 0 THEN 1 ELSE 0 END) AS started_count").
		From(models.ASSET_TABLE).
		LeftJoin(models.ASSET_PROGRESS_TABLE + " ON " + models.ASSET_TABLE_ID + " = " + models.ASSET_PROGRESS_TABLE_ASSET_ID).
		Where(squirrel.And{squirrel.Eq{models.ASSET_TABLE_COURSE_ID: courseID}}).
		ToSql()

	var totalAssetCount sql.NullInt32
	var completedAssetCount sql.NullInt32
	var startedAssetCount sql.NullInt32

	q := database.QuerierFromContext(ctx, dao.db)
	err := q.QueryRow(query, args...).Scan(&totalAssetCount, &completedAssetCount, &startedAssetCount)
	if err != nil {
		return err
	}

	// Get the course progress
	courseProgress := &models.CourseProgress{}
	err = dao.Get(ctx, courseProgress, &database.Options{Where: squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: courseID}})
	if err != nil {
		return err
	}

	now := types.NowDateTime()

	courseProgress.Percent = int(math.Abs((float64(completedAssetCount.Int32) * float64(100)) / float64(totalAssetCount.Int32)))

	if startedAssetCount.Int32 > 0 || courseProgress.Percent > 0 && courseProgress.Percent <= 100 {
		courseProgress.Started = true
		courseProgress.StartedAt = now
	} else {
		courseProgress.Started = false
		courseProgress.StartedAt = types.DateTime{}
	}

	if courseProgress.Percent == 100 {
		courseProgress.CompletedAt = now
	} else {
		courseProgress.CompletedAt = types.DateTime{}
	}

	// Update the course progress
	_, err = dao.Update(ctx, courseProgress)
	return err
}
