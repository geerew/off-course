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

	if courseProgress.UserID == "" {
		userId, ok := ctx.Value(types.UserContextKey).(string)
		if !ok || userId == "" {
			return utils.ErrMissingUserId
		}
		courseProgress.UserID = userId
	}

	return dao.Create(ctx, courseProgress)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourseProgress retrieves an course progress
//
// When options is nil or options.Where is nil, the function will use the ID to filter course progress
func (dao *DAO) GetCourseProgress(ctx context.Context, courseProgress *models.CourseProgress, options *database.Options) error {
	if courseProgress == nil {
		return utils.ErrNilPtr
	}

	if options == nil || options.Where == nil {
		if courseProgress.Id() == "" {
			return utils.ErrInvalidId
		}

		options = &database.Options{
			Where: squirrel.Eq{courseProgress.Table() + "." + models.BASE_ID: courseProgress.Id()},
		}
	}

	if options.Where == nil {
	}

	return dao.Get(ctx, courseProgress, options)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshCourseProgress calculates and updates the progress for a course
//
// This function calculates the overall course progress by:
// 1. Counting video and non-video assets separately
// 2. For video assets with metadata, calculating progress based on watched duration
// 3. For non-video assets and videos without metadata, using completion flags
// 4. Weighting the progress based on the proportion of each asset type
// 5. Updating the course progress record with the calculated percentage and status
//
// The course is considered:
// - Started: when any asset has progress or is completed
// - Completed: when the overall progress reaches 100%
//
// Parameters:
//   - ctx: The context for the database operation
//   - courseID: The ID of the course to refresh progress for
//
// Returns:
//   - error: Any error encountered during the refresh operation
//
// RefreshCourseProgress calculates and updates the progress for a course.
//
// This function calculates the overall course progress by:
// 1. Counting video and non-video assets separately
// 2. For video assets with metadata, calculating progress based on watched duration
// 3. For non-video assets and videos without metadata, using completion flags
// 4. Weighting the progress based on the proportion of each asset type
// 5. Updating the course progress record with the calculated percentage and status
//
// The course is considered:
// - Started: when any asset has progress or is completed
// - Completed: when the overall progress reaches 100%
//
// Parameters:
//   - ctx: The context for the database operation
//   - courseID: The ID of the course to refresh progress for
//
// Returns:
//   - error: Any error encountered during the refresh operation
//
// TODO optimize this function to reduce the number of queries
func (dao *DAO) RefreshCourseProgress(ctx context.Context, courseID string) error {
	if courseID == "" {
		return utils.ErrInvalidId
	}

	// Extract user ID from context
	userId, ok := ctx.Value(types.UserContextKey).(string)
	if !ok || userId == "" {
		return utils.ErrMissingUserId
	}

	// Count video and non-video assets separately
	assetCountQuery, assetCountArgs, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(
			"COUNT(CASE WHEN "+models.ASSET_TYPE+" = 'video' THEN 1 END) AS video_count",
			"COUNT(CASE WHEN "+models.ASSET_TYPE+" != 'video' THEN 1 END) AS non_video_count").
		From(models.ASSET_TABLE).
		Where(squirrel.Eq{models.ASSET_TABLE_COURSE_ID: courseID}).
		ToSql()

	var videoCount sql.NullInt32
	var nonVideoCount sql.NullInt32

	q := database.QuerierFromContext(ctx, dao.db)
	err := q.QueryRow(assetCountQuery, assetCountArgs...).Scan(&videoCount, &nonVideoCount)
	if err != nil {
		return err
	}

	// Calculate progress metrics - include user_id in the join conditions
	progressQuery, progressArgs, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(
			// Count videos with metadata
			"COUNT(CASE WHEN "+models.ASSET_TYPE+" = 'video' AND "+models.VIDEO_METADATA_TABLE_DURATION+" IS NOT NULL AND "+models.VIDEO_METADATA_TABLE_DURATION+" > 0 THEN 1 END) AS videos_with_metadata",
			// Sum of all video durations
			"COALESCE(SUM(CASE WHEN "+models.ASSET_TYPE+" = 'video' THEN "+models.VIDEO_METADATA_TABLE_DURATION+" ELSE 0 END), 0) AS total_video_duration",
			// Calculate watched video duration (for videos with metadata)
			"COALESCE(SUM(CASE WHEN "+models.ASSET_TYPE+" = 'video' AND "+models.VIDEO_METADATA_TABLE_DURATION+" IS NOT NULL AND "+models.VIDEO_METADATA_TABLE_DURATION+" > 0 THEN "+
				"CASE WHEN "+models.ASSET_PROGRESS_TABLE_COMPLETED+" THEN "+models.VIDEO_METADATA_TABLE_DURATION+" "+
				"ELSE MIN("+models.ASSET_PROGRESS_TABLE_VIDEO_POS+", "+models.VIDEO_METADATA_TABLE_DURATION+") END "+
				"ELSE 0 END), 0) AS watched_video_duration",
			// Count completed videos without metadata
			"COUNT(CASE WHEN "+models.ASSET_TYPE+" = 'video' AND ("+models.VIDEO_METADATA_TABLE_DURATION+" IS NULL OR "+models.VIDEO_METADATA_TABLE_DURATION+" = 0) AND "+models.ASSET_PROGRESS_TABLE_COMPLETED+" THEN 1 END) AS completed_videos_no_metadata",
			// Count total videos without metadata
			"COUNT(CASE WHEN "+models.ASSET_TYPE+" = 'video' AND ("+models.VIDEO_METADATA_TABLE_DURATION+" IS NULL OR "+models.VIDEO_METADATA_TABLE_DURATION+" = 0) THEN 1 END) AS total_videos_no_metadata",
			// Count completed non-video assets
			"COUNT(CASE WHEN "+models.ASSET_TYPE+" != 'video' AND "+models.ASSET_PROGRESS_TABLE_COMPLETED+" THEN 1 END) AS completed_non_video_count",
			// Count started assets
			"SUM(CASE WHEN "+models.ASSET_PROGRESS_TABLE_VIDEO_POS+" > 0 OR "+models.ASSET_PROGRESS_TABLE_COMPLETED+" THEN 1 ELSE 0 END) AS started_count").
		From(models.ASSET_TABLE).
		// Include user_id in the join condition
		LeftJoin(models.ASSET_PROGRESS_TABLE + " ON " + models.ASSET_TABLE_ID + " = " + models.ASSET_PROGRESS_TABLE_ASSET_ID +
			" AND " + models.ASSET_PROGRESS_TABLE_USER_ID + " = ?").
		LeftJoin(models.VIDEO_METADATA_TABLE + " ON " + models.ASSET_TABLE_ID + " = " + models.VIDEO_METADATA_TABLE_ASSET_ID).
		Where(squirrel.Eq{models.ASSET_TABLE_COURSE_ID: courseID}).
		ToSql()

	// Add userID as the first parameter
	progressArgs = append([]interface{}{userId}, progressArgs...)

	var videosWithMetadata sql.NullInt32
	var totalVideoDuration sql.NullInt64
	var watchedVideoDuration sql.NullInt64
	var completedVideosNoMetadata sql.NullInt32
	var totalVideosNoMetadata sql.NullInt32
	var completedNonVideoCount sql.NullInt32
	var startedAssetCount sql.NullInt32

	err = q.QueryRow(progressQuery, progressArgs...).Scan(
		&videosWithMetadata,
		&totalVideoDuration,
		&watchedVideoDuration,
		&completedVideosNoMetadata,
		&totalVideosNoMetadata,
		&completedNonVideoCount,
		&startedAssetCount,
	)
	if err != nil {
		return err
	}

	// Get the course progress for this specific user
	courseProgress := &models.CourseProgress{}
	err = dao.GetCourseProgress(ctx, courseProgress, &database.Options{
		Where: squirrel.And{
			squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: courseID},
			squirrel.Eq{models.COURSE_PROGRESS_TABLE_USER_ID: userId},
		},
	})

	// If no progress record exists for this user, create one
	if err == sql.ErrNoRows {
		courseProgress = &models.CourseProgress{
			CourseID: courseID,
			UserID:   userId,
		}

		if err = dao.Create(ctx, courseProgress); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	now := types.NowDateTime()

	// Calculate percentage based on both video and non-video assets
	var percent int
	totalAssets := videoCount.Int32 + nonVideoCount.Int32

	if totalAssets > 0 {
		// For videos with metadata, calculate based on duration
		var videosWithMetadataPercent float64 = 0
		if videosWithMetadata.Int32 > 0 && totalVideoDuration.Int64 > 0 {
			videosWithMetadataPercent = float64(watchedVideoDuration.Int64) / float64(totalVideoDuration.Int64)
		}

		// For videos without metadata, calculate based on completion flag
		var videosWithoutMetadataPercent float64 = 0
		if totalVideosNoMetadata.Int32 > 0 {
			videosWithoutMetadataPercent = float64(completedVideosNoMetadata.Int32) / float64(totalVideosNoMetadata.Int32)
		}

		// For non-videos, calculate based on completion flag
		var nonVideoProgressPercent float64 = 0
		if nonVideoCount.Int32 > 0 {
			nonVideoProgressPercent = float64(completedNonVideoCount.Int32) / float64(nonVideoCount.Int32)
		}

		// Calculate the weighted percentage
		var totalWeight float64 = 0
		var weightedSum float64 = 0

		if videosWithMetadata.Int32 > 0 {
			weight := float64(videosWithMetadata.Int32) / float64(totalAssets)
			weightedSum += videosWithMetadataPercent * weight
			totalWeight += weight
		}

		if totalVideosNoMetadata.Int32 > 0 {
			weight := float64(totalVideosNoMetadata.Int32) / float64(totalAssets)
			weightedSum += videosWithoutMetadataPercent * weight
			totalWeight += weight
		}

		if nonVideoCount.Int32 > 0 {
			weight := float64(nonVideoCount.Int32) / float64(totalAssets)
			weightedSum += nonVideoProgressPercent * weight
			totalWeight += weight
		}

		// Calculate final percentage, handling possible division by zero
		if totalWeight > 0 {
			percent = int(math.Floor((weightedSum / totalWeight) * 100))
		}
	}

	// Cap at 100%
	if percent > 100 {
		percent = 100
	}
	courseProgress.Percent = percent

	// Update started status
	if startedAssetCount.Int32 > 0 || courseProgress.Percent > 0 {
		courseProgress.Started = true
		// Only set started_at if it's not already set
		if courseProgress.StartedAt.IsZero() {
			courseProgress.StartedAt = now
		}
	} else {
		courseProgress.Started = false
		courseProgress.StartedAt = types.DateTime{}
	}

	// Update completed status - must be 100% to be considered completed
	if courseProgress.Percent == 100 {
		// Only set completed_at if it's not already set
		if courseProgress.CompletedAt.IsZero() {
			courseProgress.CompletedAt = now
		}
	} else {
		courseProgress.CompletedAt = types.DateTime{}
	}

	// Update the course progress
	_, err = dao.Update(ctx, courseProgress)
	return err
}
