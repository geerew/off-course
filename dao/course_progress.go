package dao

import (
	"context"
	"math"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// SyncCourseProgress calculates the course progress for a given course ID and upserts a
// course progress record based upon the course id + user id
func (dao *DAO) SyncCourseProgress(ctx context.Context, courseID string) error {
	if courseID == "" {
		return utils.ErrCourseId
	}

	principal, err := principalFromCtx(ctx)
	if err != nil {
		return err
	}

	metrics, err := dao.fetchCourseMetrics(ctx, courseID, principal.UserID)
	if err != nil {
		return err
	}

	if metrics == nil {
		// No metrics means no assets → no progress to track
		return nil
	}

	courseProgress := &models.CourseProgress{
		CourseID: courseID,
		UserID:   principal.UserID,
	}

	courseProgress.RefreshId()
	courseProgress.RefreshCreatedAt()
	courseProgress.RefreshUpdatedAt()

	setProgress(*metrics, courseProgress)

	builderOpts := newBuilderOptions(models.COURSE_PROGRESS_TABLE).
		WithData(
			map[string]interface{}{
				models.BASE_ID:                      courseProgress.ID,
				models.COURSE_PROGRESS_COURSE_ID:    courseProgress.CourseID,
				models.COURSE_PROGRESS_USER_ID:      courseProgress.UserID,
				models.COURSE_PROGRESS_STARTED:      courseProgress.Started,
				models.COURSE_PROGRESS_STARTED_AT:   courseProgress.StartedAt,
				models.COURSE_PROGRESS_PERCENT:      courseProgress.Percent,
				models.COURSE_PROGRESS_COMPLETED_AT: courseProgress.CompletedAt,
				models.BASE_CREATED_AT:              courseProgress.CreatedAt,
				models.BASE_UPDATED_AT:              courseProgress.UpdatedAt,
			},
		).WithSuffix("" +
		"ON CONFLICT( " + models.COURSE_PROGRESS_TABLE_COURSE_ID + "," + models.COURSE_PROGRESS_TABLE_USER_ID + ") DO UPDATE " +
		"SET " +
		models.COURSE_PROGRESS_PERCENT + " = excluded. " + models.COURSE_PROGRESS_PERCENT + ", " +
		models.COURSE_PROGRESS_STARTED + " = excluded. " + models.COURSE_PROGRESS_STARTED + ", " +
		// StartedAt case (only if it is NULL → non-NULL or non-NULL → NULL)
		models.COURSE_PROGRESS_STARTED_AT + " = CASE " +
		"WHEN (coalesce(" + models.COURSE_PROGRESS_STARTED_AT + ",'') = '' AND excluded." + models.COURSE_PROGRESS_STARTED_AT + " IS NOT NULL AND excluded." + models.COURSE_PROGRESS_STARTED_AT + " != '') THEN excluded." + models.COURSE_PROGRESS_STARTED_AT + " " +
		"WHEN (" + models.COURSE_PROGRESS_STARTED_AT + " IS NOT NULL AND " + models.COURSE_PROGRESS_STARTED_AT + " != '' AND (excluded." + models.COURSE_PROGRESS_STARTED_AT + " IS NULL OR excluded." + models.COURSE_PROGRESS_STARTED_AT + " = '')) THEN '' " +
		"ELSE " + models.COURSE_PROGRESS_STARTED_AT + " END, " +
		// CompletedAt case (only if it is NULL → non-NULL or non-NULL → NULL)
		models.COURSE_PROGRESS_COMPLETED_AT + " = CASE " +
		"WHEN (coalesce(" + models.COURSE_PROGRESS_COMPLETED_AT + ",'') = '' AND excluded." + models.COURSE_PROGRESS_COMPLETED_AT + " IS NOT NULL AND excluded." + models.COURSE_PROGRESS_COMPLETED_AT + " != '') THEN excluded." + models.COURSE_PROGRESS_COMPLETED_AT + " " +
		"WHEN (" + models.COURSE_PROGRESS_COMPLETED_AT + " IS NOT NULL AND " + models.COURSE_PROGRESS_COMPLETED_AT + " != '' AND (excluded." + models.COURSE_PROGRESS_COMPLETED_AT + " IS NULL OR excluded." + models.COURSE_PROGRESS_COMPLETED_AT + " = '')) THEN '' " +
		"ELSE " + models.COURSE_PROGRESS_COMPLETED_AT + " END, " +
		// Update the updated_at field
		models.BASE_UPDATED_AT + " = excluded." + models.BASE_UPDATED_AT)

	return createGeneric(ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourseProgress gets a record from the course progress table based upon the where clause in the options. If
// there is no where clause, it will return the first record in the table
func (dao *DAO) GetCourseProgress(ctx context.Context, dbOpts *database.Options) (*models.CourseProgress, error) {
	builderOpts := newBuilderOptions(models.COURSE_PROGRESS_TABLE).
		WithColumns(
			models.COURSE_PROGRESS_TABLE + ".*",
		).
		SetDbOpts(dbOpts).
		WithLimit(1)

	return getGeneric[models.CourseProgress](ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListCourseProgress gets all records from the course progress table based upon the where clause and pagination
// in the options
func (dao *DAO) ListCourseProgress(ctx context.Context, dbOpts *database.Options) ([]*models.CourseProgress, error) {
	builderOpts := newBuilderOptions(models.COURSE_PROGRESS_TABLE).
		WithColumns(
			models.COURSE_PROGRESS_TABLE + ".*",
		).
		SetDbOpts(dbOpts)

	return listGeneric[models.CourseProgress](ctx, dao, *builderOpts)
}

// DeleteCourseProgress deletes records from the course progress table
//
// Errors when a where clause is not provided
func (dao *DAO) DeleteCourseProgress(ctx context.Context, dbOpts *database.Options) error {
	if dbOpts == nil || dbOpts.Where == nil {
		return utils.ErrWhere
	}

	builderOpts := newBuilderOptions(models.COURSE_PROGRESS_TABLE).SetDbOpts(dbOpts)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Private
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseMetrics struct {
	VideoCount             int   `db:"video_count"`
	NonVideoCount          int   `db:"non_video_count"`
	VideosWithMeta         int   `db:"videos_with_metadata"`
	TotalVideoDuration     int64 `db:"total_video_duration"`
	WatchedVideoDuration   int64 `db:"watched_video_duration"`
	CompletedNoMetaCount   int   `db:"completed_videos_no_metadata"`
	TotalNoMetaCount       int   `db:"total_videos_no_metadata"`
	CompletedNonVideoCount int   `db:"completed_non_video_count"`
	StartedAssetCount      int   `db:"started_count"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// fetchCourseMetrics retrieves various metrics for a course based on the course ID and user ID
//
// TODO Change this to work around the lesson
// TODO Change the userID to use ? instead of string interpolation
func (dao *DAO) fetchCourseMetrics(ctx context.Context, courseID, userID string) (*courseMetrics, error) {
	// TODO fix
	// dbOpts := database.NewOptions().
	// 	WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: courseID})

	// builderOpts := newBuilderOptions(models.COURSE_TABLE).
	// 	WithColumns(
	// 		// Count of video
	// 		"COUNT(CASE WHEN "+models.ASSET_TABLE_TYPE+"='video' THEN 1 END) AS video_count",
	// 		// Count of non-video assets
	// 		"COUNT(CASE WHEN "+models.ASSET_TABLE_TYPE+"!='video' THEN 1 END) AS non_video_count",
	// 		// Count of videos with metadata whose duration > 0
	// 		"COUNT(CASE WHEN "+models.ASSET_TABLE_TYPE+"='video' AND "+models.VIDEO_METADATA_TABLE_DURATION+">0 THEN 1 END) AS videos_with_metadata",
	// 		// Total duration of all video assets
	// 		"COALESCE(SUM(CASE WHEN "+models.ASSET_TABLE_TYPE+"='video' THEN "+models.VIDEO_METADATA_TABLE_DURATION+" END),0) AS total_video_duration",
	// 		// Total watched duration of videos with metadata
	// 		"COALESCE(SUM(CASE WHEN "+models.ASSET_TABLE_TYPE+"='video' AND "+models.VIDEO_METADATA_TABLE_DURATION+">0 THEN CASE WHEN "+models.ASSET_PROGRESS_TABLE_COMPLETED+" THEN "+models.VIDEO_METADATA_TABLE_DURATION+" ELSE MIN("+models.ASSET_PROGRESS_TABLE_VIDEO_POS+", "+models.VIDEO_METADATA_TABLE_DURATION+") END END),0) AS watched_video_duration",
	// 		// Count of completed video assets without metadata
	// 		"COUNT(CASE WHEN "+models.ASSET_TABLE_TYPE+"='video' AND ("+models.VIDEO_METADATA_TABLE_DURATION+" IS NULL OR "+models.VIDEO_METADATA_TABLE_DURATION+"=0) AND "+models.ASSET_PROGRESS_TABLE_COMPLETED+" THEN 1 END) AS completed_videos_no_metadata",
	// 		// Count of total video assets without metadata
	// 		"COUNT(CASE WHEN "+models.ASSET_TABLE_TYPE+"='video' AND ("+models.VIDEO_METADATA_TABLE_DURATION+" IS NULL OR "+models.VIDEO_METADATA_TABLE_DURATION+"=0) THEN 1 END) AS total_videos_no_metadata",
	// 		// Count of completed non-video assets
	// 		"COUNT(CASE WHEN "+models.ASSET_TABLE_TYPE+"!='video' AND "+models.ASSET_PROGRESS_TABLE_COMPLETED+" THEN 1 END) AS completed_non_video_count",
	// 		// Count of started assets (video or non-video)
	// 		"COALESCE(SUM(CASE WHEN "+models.ASSET_PROGRESS_TABLE_VIDEO_POS+">0 OR "+models.ASSET_PROGRESS_TABLE_COMPLETED+" THEN 1 ELSE 0 END), 0) AS started_count",
	// 	).
	// 	WithLeftJoin(models.ASSET_TABLE, fmt.Sprintf("%s = %s", models.ASSET_TABLE_COURSE_ID, models.COURSE_TABLE_ID)).
	// 	WithLeftJoin(models.ASSET_PROGRESS_TABLE, fmt.Sprintf("%s = %s AND %s = '%s'", models.ASSET_PROGRESS_TABLE_ASSET_ID, models.ASSET_TABLE_ID, models.ASSET_PROGRESS_TABLE_USER_ID, userID)).
	// 	WithLeftJoin(models.VIDEO_METADATA_TABLE, fmt.Sprintf("%s = %s", models.VIDEO_METADATA_TABLE_ASSET_ID, models.ASSET_TABLE_ID)).
	// 	WithGroupBy(models.COURSE_TABLE_ID).
	// 	SetDbOpts(dbOpts)

	// metrics, err := getGeneric[courseMetrics](ctx, dao, *builderOpts)
	// if err != nil {
	// 	return nil, err
	// }

	// if metrics == nil {
	// 	return nil, utils.ErrCourseId
	// }

	// return metrics, nil
	return nil, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setProgress calculates the progress of a course based on the provided metrics and updates the
// course progress object accordingly
func setProgress(metrics courseMetrics, courseProgress *models.CourseProgress) {
	var percent int
	var started bool
	var startedAt types.DateTime
	var completedAt types.DateTime

	// Set the percent
	totalAssets := metrics.VideoCount + metrics.NonVideoCount
	if totalAssets > 0 {
		// Video with metadata
		vw := 0.0
		if metrics.VideosWithMeta > 0 && metrics.TotalVideoDuration > 0 {
			vw = float64(metrics.WatchedVideoDuration) / float64(metrics.TotalVideoDuration)
		}

		// Video without metadata
		vnm := 0.0
		if metrics.TotalNoMetaCount > 0 {
			vnm = float64(metrics.CompletedNoMetaCount) / float64(metrics.TotalNoMetaCount)
		}

		// Non-video assets
		nv := 0.0
		if metrics.NonVideoCount > 0 {
			nv = float64(metrics.CompletedNonVideoCount) / float64(metrics.NonVideoCount)
		}

		weightedSum := 0.0
		totalWeight := 0.0
		if metrics.VideosWithMeta > 0 {
			w := float64(metrics.VideosWithMeta) / float64(totalAssets)
			weightedSum += vw * w
			totalWeight += w
		}

		if metrics.TotalNoMetaCount > 0 {
			w := float64(metrics.TotalNoMetaCount) / float64(totalAssets)
			weightedSum += vnm * w
			totalWeight += w
		}

		if metrics.NonVideoCount > 0 {
			w := float64(metrics.NonVideoCount) / float64(totalAssets)
			weightedSum += nv * w
			totalWeight += w
		}

		if totalWeight > 0 {
			percent = int(math.Floor((weightedSum / totalWeight) * 100))
		}
	}

	if percent > 100 {
		percent = 100
	}

	now := types.NowDateTime()

	// started and startedAt
	if metrics.StartedAssetCount > 0 || percent > 0 {
		started = true
		startedAt = now
	}

	// completed
	if percent == 100 {
		completedAt = now
	}

	courseProgress.Percent = percent
	courseProgress.Started = started
	courseProgress.StartedAt = startedAt
	courseProgress.CompletedAt = completedAt
}
