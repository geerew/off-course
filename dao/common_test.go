package dao

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(tb testing.TB) (*DAO, context.Context) {
	tb.Helper()

	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.TestWriteFn(&logs, &logsMux),
	})
	require.NoError(tb, err, "Failed to initialize logger")

	// DB
	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: "./oc_data",
		AppFs:   appfs.New(afero.NewMemMapFs(), logger),
		Testing: true,
	})

	require.NoError(tb, err)
	require.NotNil(tb, dbManager)

	dao := &DAO{db: dbManager.DataDb}

	// User
	user := &models.User{
		Username:     "test-user",
		DisplayName:  "Test User",
		PasswordHash: "test-password",
		Role:         types.UserRoleAdmin,
	}
	require.NoError(tb, dao.CreateUser(context.Background(), user))

	ctx := context.WithValue(context.Background(), types.UserContextKey, user.ID)

	return dao, ctx
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Count(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{}
		count, err := Count(ctx, dao, course, nil)
		require.NoError(t, err)
		require.Zero(t, count)
	})

	t.Run("entries", func(t *testing.T) {
		dao, ctx := setup(t)

		for i := range 5 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.CreateCourse(ctx, course))
		}

		course := &models.Course{}
		count, err := Count(ctx, dao, course, nil)
		require.NoError(t, err)
		require.Equal(t, count, 5)
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, Create(ctx, dao, course))
			courses = append(courses, course)
		}

		course := &models.Course{}

		// ----------------------------
		// EQUALS ID
		// ----------------------------
		count, err := Count(ctx, dao, course, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: courses[1].ID}})
		require.NoError(t, err)
		require.Equal(t, 1, count)

		// ----------------------------
		// NOT EQUALS ID
		// ----------------------------
		count, err = Count(ctx, dao, course, &database.Options{Where: squirrel.NotEq{models.COURSE_TABLE_ID: courses[1].ID}})
		require.NoError(t, err)
		require.Equal(t, 2, count)

		// ----------------------------
		// ERROR
		// ----------------------------
		count, err = Count(ctx, dao, course, &database.Options{Where: squirrel.Eq{"": ""}})
		require.ErrorContains(t, err, "syntax error")
		require.Zero(t, count)
	})

	t.Run("invalid model", func(t *testing.T) {
		dao, ctx := setup(t)

		count, err := Count(ctx, dao, nil, nil)
		require.ErrorIs(t, err, utils.ErrNilPtr)
		require.Zero(t, count)
	})

	t.Run("db error", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{}

		_, err := dao.db.Exec("DROP TABLE IF EXISTS " + course.Table())
		require.NoError(t, err)

		_, err = Count(ctx, dao, course, nil)
		require.ErrorContains(t, err, "no such table: "+course.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{}
		scan := &models.Scan{}
		videoAsset := &models.Asset{}
		videoAssetProgress := &models.AssetProgress{}
		videoMetadata := &models.VideoMetadata{}
		attachment := &models.Attachment{}
		tag := &models.Tag{}

		// Create user
		{
			user := &models.User{Username: "user1", DisplayName: "User 1", PasswordHash: "1234", Role: types.UserRoleAdmin}
			require.NoError(t, dao.CreateUser(ctx, user))

			// Get user
			userResult := &models.User{}
			require.NoError(t, Get(ctx, dao, userResult, &database.Options{Where: squirrel.Eq{models.USER_TABLE_ID: user.ID}}))
			require.Equal(t, user.ID, userResult.ID)
			require.True(t, userResult.CreatedAt.Equal(user.CreatedAt))
			require.True(t, userResult.UpdatedAt.Equal(user.UpdatedAt))
			require.Equal(t, user.Username, userResult.Username)
			require.Equal(t, user.PasswordHash, userResult.PasswordHash)
			require.Equal(t, user.Role, userResult.Role)
		}

		// Create course
		{
			course = &models.Course{Title: "Course 1", Path: "/course 1", Available: true, CardPath: "/course 1/card 1.jpg"}
			require.NoError(t, dao.CreateCourse(ctx, course))

			// Get course
			courseResult := &models.Course{}
			require.NoError(t, dao.GetCourse(ctx, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: course.ID}}))
			require.Equal(t, course.ID, courseResult.ID)
			require.True(t, courseResult.CreatedAt.Equal(course.CreatedAt))
			require.True(t, courseResult.UpdatedAt.Equal(course.UpdatedAt))
			require.Equal(t, course.Title, courseResult.Title)
			require.Equal(t, course.Path, courseResult.Path)
			require.Equal(t, course.CardPath, courseResult.CardPath)
			require.True(t, courseResult.Available)
			require.Empty(t, courseResult.ScanStatus)
			require.Nil(t, courseResult.Progress)
		}

		// Create scan
		{
			scan = &models.Scan{CourseID: course.ID}
			require.NoError(t, dao.CreateScan(ctx, scan))

			scanResult := &models.Scan{}
			require.NoError(t, Get(ctx, dao, scanResult, nil))
			require.Equal(t, scan.ID, scanResult.ID)
			require.True(t, scanResult.CreatedAt.Equal(scan.CreatedAt))
			require.True(t, scanResult.UpdatedAt.Equal(scan.UpdatedAt))
			require.Equal(t, scan.CourseID, scanResult.CourseID)
			require.True(t, scanResult.Status.IsWaiting())
			require.Equal(t, course.Path, scanResult.CoursePath)
		}

		// Get course (again) and check scan status
		{
			courseResult := &models.Course{}
			require.NoError(t, dao.GetCourse(ctx, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: course.ID}}))
			require.Equal(t, course.ID, courseResult.ID)
			require.True(t, courseResult.ScanStatus.IsWaiting())
		}

		// Create video asset
		{
			videoAsset = &models.Asset{
				CourseID: course.ID,
				Title:    "Asset 1",
				Prefix:   sql.NullInt16{Int16: 1, Valid: true},
				Chapter:  "Chapter 1",
				Type:     *types.NewAsset("mp4"),
				Path:     "/course 1/Chapter 1/01 videoAsset.mp4",
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "1234",
			}
			require.NoError(t, dao.CreateAsset(ctx, videoAsset))

			videoAssetResult := &models.Asset{}
			require.NoError(t, Get(ctx, dao, videoAssetResult, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE_ID: videoAsset.ID}}))
			require.Equal(t, videoAsset.ID, videoAssetResult.ID)
			require.True(t, videoAssetResult.CreatedAt.Equal(videoAsset.CreatedAt))
			require.True(t, videoAssetResult.UpdatedAt.Equal(videoAsset.UpdatedAt))
			require.Equal(t, videoAsset.CourseID, videoAssetResult.CourseID)
			require.Equal(t, videoAsset.Title, videoAssetResult.Title)
			require.Equal(t, videoAsset.Prefix, videoAssetResult.Prefix)
			require.Equal(t, videoAsset.Chapter, videoAssetResult.Chapter)
			require.Equal(t, videoAsset.Type, videoAssetResult.Type)
			require.Equal(t, videoAsset.Path, videoAssetResult.Path)
			require.Equal(t, videoAsset.FileSize, videoAssetResult.FileSize)
			require.Equal(t, videoAsset.ModTime, videoAssetResult.ModTime)
			require.Equal(t, videoAsset.Hash, videoAssetResult.Hash)
			require.Len(t, videoAssetResult.Attachments, 0)
			require.Nil(t, videoAssetResult.VideoMetadata)
			require.Nil(t, videoAssetResult.Progress)
		}

		// Create attachment
		{
			attachment = &models.Attachment{
				AssetID: videoAsset.ID,
				Title:   "Attachment 1",
				Path:    "/course 1/01 Attachment 1.txt",
			}
			require.NoError(t, dao.CreateAttachment(ctx, attachment))

			// Get attachment
			attachmentResult := &models.Attachment{}
			require.NoError(t, Get(ctx, dao, attachmentResult, nil))
			require.Equal(t, attachment.ID, attachmentResult.ID)
			require.True(t, attachmentResult.CreatedAt.Equal(attachment.CreatedAt))
			require.True(t, attachmentResult.UpdatedAt.Equal(attachment.UpdatedAt))
			require.Equal(t, attachment.AssetID, attachmentResult.AssetID)
			require.Equal(t, attachment.Title, attachmentResult.Title)
			require.Equal(t, attachment.Path, attachmentResult.Path)
		}

		// Get video asset and check attachments
		{
			videoAssetResult := &models.Asset{}
			require.NoError(t, Get(ctx, dao, videoAssetResult, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE_ID: videoAsset.ID}}))
			require.Equal(t, videoAsset.ID, videoAssetResult.ID)
			require.Len(t, videoAssetResult.Attachments, 1)
			require.Equal(t, attachment.Title, videoAssetResult.Attachments[0].Title)
		}

		// Create video metadata
		{
			videoMetadata = &models.VideoMetadata{
				AssetID:    videoAsset.ID,
				Duration:   120,
				Width:      1280,
				Height:     720,
				Codec:      "h264",
				Resolution: "720p",
			}
			require.NoError(t, dao.CreateVideoMetadata(ctx, videoMetadata))

			// Get video metadata
			videoMetadataResult := &models.VideoMetadata{}
			require.NoError(t, Get(ctx, dao, videoMetadataResult, nil))
			require.Equal(t, videoMetadata.ID, videoMetadataResult.ID)
			require.True(t, videoMetadataResult.CreatedAt.Equal(videoMetadata.CreatedAt))
			require.True(t, videoMetadataResult.UpdatedAt.Equal(videoMetadata.UpdatedAt))
			require.Equal(t, videoMetadata.AssetID, videoMetadataResult.AssetID)
			require.Equal(t, videoMetadata.Duration, videoMetadataResult.Duration)
			require.Equal(t, videoMetadata.Width, videoMetadataResult.Width)
			require.Equal(t, videoMetadata.Height, videoMetadataResult.Height)
			require.Equal(t, videoMetadata.Codec, videoMetadataResult.Codec)
			require.Equal(t, videoMetadata.Resolution, videoMetadataResult.Resolution)
		}

		// Get video asset and check video metadata
		{
			videoAssetResult := &models.Asset{}
			require.NoError(t, Get(ctx, dao, videoAssetResult, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE_ID: videoAsset.ID}}))
			require.Equal(t, videoAsset.ID, videoAssetResult.ID)
			require.NotNil(t, videoAssetResult.VideoMetadata)
			require.Equal(t, videoAssetResult.VideoMetadata.ID, videoMetadata.ID)
		}

		// Create video asset progress. It will be created for the user within the context
		{
			videoAssetProgress = &models.AssetProgress{
				AssetID:   videoAsset.ID,
				VideoPos:  60,
				Completed: false,
			}
			require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, videoAssetProgress))

			// Get video asset progress
			videoAssetProgressResult := &models.AssetProgress{}
			require.NoError(t, Get(ctx, dao, videoAssetProgressResult, &database.Options{Where: squirrel.Eq{models.ASSET_PROGRESS_TABLE_ID: videoAssetProgress.ID}}))
			require.Equal(t, videoAssetProgress.ID, videoAssetProgressResult.ID)
			require.True(t, videoAssetProgressResult.CreatedAt.Equal(videoAssetProgress.CreatedAt))
			require.True(t, videoAssetProgressResult.UpdatedAt.Equal(videoAssetProgress.UpdatedAt))
			require.Equal(t, videoAssetProgress.AssetID, videoAssetProgressResult.AssetID)
			require.Equal(t, videoAssetProgress.VideoPos, videoAssetProgressResult.VideoPos)
			require.Equal(t, videoAssetProgress.Completed, videoAssetProgressResult.Completed)
		}

		// Get course (again) and check progress for the user
		{
			courseResult := &models.Course{}
			require.NoError(t, dao.GetCourse(ctx, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: course.ID}}))
			require.Equal(t, course.ID, courseResult.ID)
			require.NotNil(t, courseResult.Progress)
			require.True(t, courseResult.Progress.Started)
			require.Equal(t, 50, courseResult.Progress.Percent)
		}

		// Create pdf asset
		{
			pdfAsset := &models.Asset{
				CourseID: course.ID,
				Title:    "Asset 2",
				Prefix:   sql.NullInt16{Int16: 2, Valid: true},
				Chapter:  "Chapter 1",
				Type:     *types.NewAsset("pdf"),
				Path:     "/course 1/Chapter 1/02 pdfAsset.pdf",
				FileSize: 1024,
				ModTime:  time.Now().Format(time.RFC3339Nano),
				Hash:     "4321",
			}
			require.NoError(t, dao.CreateAsset(ctx, pdfAsset))

			// Get pdf asset
			pdfAssetResult := &models.Asset{}
			require.NoError(t, Get(ctx, dao, pdfAssetResult, &database.Options{Where: squirrel.Eq{models.ASSET_TABLE_ID: pdfAsset.ID}}))
			require.Equal(t, pdfAsset.ID, pdfAssetResult.ID)
			require.True(t, pdfAssetResult.CreatedAt.Equal(pdfAsset.CreatedAt))
			require.True(t, pdfAssetResult.UpdatedAt.Equal(pdfAsset.UpdatedAt))
			require.Equal(t, pdfAsset.CourseID, pdfAssetResult.CourseID)
			require.Equal(t, pdfAsset.Title, pdfAssetResult.Title)
			require.Equal(t, pdfAsset.Prefix, pdfAssetResult.Prefix)
			require.Equal(t, pdfAsset.Chapter, pdfAssetResult.Chapter)
			require.Equal(t, pdfAsset.Type, pdfAssetResult.Type)
			require.Equal(t, pdfAsset.Path, pdfAssetResult.Path)
			require.Equal(t, pdfAsset.FileSize, pdfAssetResult.FileSize)
			require.Equal(t, pdfAsset.ModTime, pdfAssetResult.ModTime)
			require.Equal(t, pdfAsset.Hash, pdfAssetResult.Hash)
			require.Len(t, pdfAssetResult.Attachments, 0)
			require.Nil(t, pdfAssetResult.VideoMetadata)
			require.Nil(t, pdfAssetResult.Progress)
		}

		// Create tag
		{
			tag = &models.Tag{Tag: "Tag 1"}
			require.NoError(t, dao.CreateTag(ctx, tag))

			// Get tag
			tagResult := &models.Tag{}
			require.NoError(t, Get(ctx, dao, tagResult, nil))
			require.Equal(t, tag.ID, tagResult.ID)
			require.True(t, tagResult.CreatedAt.Equal(tag.CreatedAt))
			require.True(t, tagResult.UpdatedAt.Equal(tag.UpdatedAt))
			require.Equal(t, tag.Tag, tagResult.Tag)
			require.Zero(t, tagResult.CourseCount)
		}

		// Create course tag
		{
			courseTag := &models.CourseTag{TagID: tag.ID, CourseID: course.ID}
			require.NoError(t, dao.CreateCourseTag(ctx, courseTag))

			// Get course tag
			courseTagResult := &models.CourseTag{}
			require.NoError(t, Get(ctx, dao, courseTagResult, nil))
			require.Equal(t, courseTag.ID, courseTagResult.ID)
			require.True(t, courseTagResult.CreatedAt.Equal(courseTag.CreatedAt))
			require.True(t, courseTagResult.UpdatedAt.Equal(courseTag.UpdatedAt))
			require.Equal(t, courseTag.TagID, courseTagResult.TagID)
			require.Equal(t, courseTag.CourseID, courseTagResult.CourseID)
			require.Equal(t, course.Title, courseTagResult.Course)
			require.Equal(t, tag.Tag, courseTagResult.Tag)
		}

		// Get tag (again) and check course count
		{
			tagResult := &models.Tag{}
			require.NoError(t, Get(ctx, dao, tagResult, nil))
			require.Equal(t, tag.ID, tagResult.ID)
			require.Equal(t, 1, tagResult.CourseCount)
		}
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{}
		err := Get(ctx, dao, course, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_PATH: "1234"}})
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, Create(ctx, dao, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		courseResult := &models.Course{}
		require.NoError(t, Get(ctx, dao, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_PATH: courses[1].Path}}))
		require.Equal(t, courses[1].ID, courseResult.ID)
	})

	t.Run("orderby", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, Create(ctx, dao, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		result := &models.Course{}
		options := &database.Options{OrderBy: []string{models.COURSE_TABLE_TITLE + " DESC"}}
		require.NoError(t, Get(ctx, dao, result, options))
		require.Equal(t, courses[2].ID, result.ID)
	})

	t.Run("invalid model", func(t *testing.T) {
		dao, ctx := setup(t)
		err := Get(ctx, dao, nil, nil)
		require.ErrorIs(t, err, utils.ErrNilPtr)
	})

	t.Run("invalid where", func(t *testing.T) {
		dao, ctx := setup(t)
		err := Get(ctx, dao, &models.Course{}, &database.Options{Where: squirrel.Eq{"`": "`"}})
		require.ErrorContains(t, err, "unrecognized token")
	})

	t.Run("db error", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{}

		_, err := dao.db.Exec("DROP TABLE IF EXISTS " + course.Table())
		require.NoError(t, err)

		err = Get(ctx, dao, course, nil)
		require.ErrorContains(t, err, "no such table: "+course.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_List(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		err := List(ctx, dao, &courses, nil)
		require.NoError(t, err)
		require.Empty(t, courses)
	})

	t.Run("entries", func(t *testing.T) {
		dao, ctx := setup(t)

		for i := range 5 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, Create(ctx, dao, course))
		}

		courses := []*models.Course{}
		err := List(ctx, dao, &courses, nil)
		require.NoError(t, err)
		require.Len(t, courses, 5)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 17 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, Create(ctx, dao, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		coursesResult := []*models.Course{}

		// Page 1 (10 items)
		p := pagination.New(1, 10)
		require.NoError(t, List(ctx, dao, &coursesResult, &database.Options{Pagination: p}))
		require.Len(t, coursesResult, 10)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, courses[0].ID, coursesResult[0].ID)
		require.Equal(t, courses[9].ID, coursesResult[9].ID)

		// Page 2 (7 items)
		p = pagination.New(2, 10)
		require.NoError(t, List(ctx, dao, &coursesResult, &database.Options{Pagination: p}))
		require.Len(t, coursesResult, 7)
		require.Equal(t, 17, p.TotalItems())
		require.Equal(t, courses[10].ID, coursesResult[0].ID)
		require.Equal(t, courses[16].ID, coursesResult[6].ID)
	})

	t.Run("orderby", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, Create(ctx, dao, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		// CREATED_AT DESC
		coursesResult := []*models.Course{}
		options := &database.Options{OrderBy: []string{models.COURSE_TABLE_TITLE + " DESC"}}
		require.NoError(t, List(ctx, dao, &coursesResult, options))
		require.Len(t, coursesResult, 3)
		require.Equal(t, courses[2].ID, coursesResult[0].ID)
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, Create(ctx, dao, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		// Equals ID or ID
		coursesResult := []*models.Course{}
		options := &database.Options{
			Where: squirrel.Or{
				squirrel.Eq{models.COURSE_TABLE_ID: courses[1].ID},
				squirrel.Eq{models.COURSE_TABLE_ID: courses[2].ID},
			},
			OrderBy: []string{models.COURSE_TABLE_CREATED_AT + " ASC"},
		}
		require.NoError(t, List(ctx, dao, &coursesResult, options))
		require.Len(t, coursesResult, 2)
		require.Equal(t, courses[1].ID, coursesResult[0].ID)
		require.Equal(t, courses[2].ID, coursesResult[1].ID)
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		// Nil
		require.ErrorIs(t, List(ctx, dao, nil, nil), utils.ErrNilPtr)

		// Not a pointer
		require.ErrorIs(t, List(ctx, dao, []*models.Course{}, nil), utils.ErrNotPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListPluck(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		dao, ctx := setup(t)

		ids, err := ListPluck[string](ctx, dao, &models.Course{}, nil, models.BASE_ID)
		require.NoError(t, err)
		require.Empty(t, ids)
	})

	t.Run("entries", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 5 {
			course := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, Create(ctx, dao, course))
			courses = append(courses, course)
			time.Sleep(1 * time.Millisecond)
		}

		options := &database.Options{OrderBy: []string{models.COURSE_TABLE_CREATED_AT + " ASC"}}

		// Course IDs
		ids, err := ListPluck[string](ctx, dao, &models.Course{}, options, models.BASE_ID)
		require.NoError(t, err)
		require.Len(t, ids, 5)
		for i := range 5 {
			require.Equal(t, courses[i].ID, ids[i])
		}

		// Course paths
		paths, err := ListPluck[string](ctx, dao, &models.Course{}, options, models.COURSE_PATH)
		require.NoError(t, err)
		require.Len(t, paths, 5)
		for i := range 5 {
			require.Equal(t, courses[i].Path, paths[i])
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, Create(ctx, dao, course))

		require.NoError(t, Delete(ctx, dao, course, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_PATH: course.Path}}))
	})

	t.Run("nil model", func(t *testing.T) {
		dao, ctx := setup(t)
		err := Delete(ctx, dao, nil, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_PATH: "1234"}})
		require.ErrorIs(t, err, utils.ErrNilPtr)
	})

	t.Run("nil where", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, Create(ctx, dao, course))
		require.NoError(t, Delete(ctx, dao, course, nil))

		// Check if it was deleted
		courseResult := &models.Course{Base: models.Base{ID: course.ID}}
		require.ErrorIs(t, Get(ctx, dao, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: course.ID}}), sql.ErrNoRows)
	})

	t.Run("db error", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{}

		_, err := dao.db.Exec("DROP TABLE IF EXISTS " + course.Table())
		require.NoError(t, err)

		err = Delete(ctx, dao, course, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_PATH: "1234"}})
		require.ErrorContains(t, err, "no such table: "+course.Table())
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RawQuery(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, Create(ctx, dao, course))

		courses := []*models.Course{}
		err := RawQuery(ctx, dao, &courses, "SELECT * FROM "+course.Table()+" WHERE "+models.COURSE_TABLE_PATH+" = ?", course.Path)
		require.NoError(t, err)
		require.Len(t, courses, 1)
		require.Equal(t, course.ID, courses[0].ID)
	})

	t.Run("nil model", func(t *testing.T) {
		dao, ctx := setup(t)
		err := RawQuery(ctx, dao, nil, "SELECT * FROM courses")
		require.ErrorIs(t, err, utils.ErrNilPtr)
	})

	t.Run("invalid query", func(t *testing.T) {
		dao, ctx := setup(t)
		courses := []*models.Course{}
		err := RawQuery(ctx, dao, &courses, "SELECT * FROM abcd1234")
		require.ErrorContains(t, err, "no such table: abcd1234")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RawExec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, Create(ctx, dao, course))

		_, err := RawExec(ctx, dao, "DELETE FROM "+course.Table()+" WHERE "+models.COURSE_TABLE_PATH+" = ?", course.Path)
		require.NoError(t, err)

		courses := []*models.Course{}
		err = RawQuery(ctx, dao, &courses, "SELECT * FROM "+course.Table()+" WHERE "+models.COURSE_TABLE_PATH+" = ?", course.Path)
		require.NoError(t, err)
		require.Empty(t, courses)
	})

	t.Run("invalid query", func(t *testing.T) {
		dao, ctx := setup(t)
		_, err := RawExec(ctx, dao, "DELETE FROM abcd1234 WHERE id = ?", "123")
		require.ErrorContains(t, err, "no such table: abcd1234")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Benchmark_Get(b *testing.B) {
	dao, ctx := setup(b)

	for i := 0; i < 1000; i++ {
		course := &models.Course{}
		course.ID = fmt.Sprintf("%d", i)
		course.Title = fmt.Sprintf("Course %d", i)
		course.Path = fmt.Sprintf("/course-%d", i)
		require.NoError(b, dao.CreateCourse(ctx, course))
		require.NoError(b, dao.RefreshCourseProgress(ctx, course.ID))

		courseProgress := &models.CourseProgress{}
		require.NoError(b, Get(ctx, dao, courseProgress, &database.Options{Where: squirrel.Eq{models.COURSE_PROGRESS_TABLE_COURSE_ID: course.ID}}))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		courseResult := &models.Course{Base: models.Base{ID: fmt.Sprintf("%d", (i % 1000))}}
		require.NoError(b, Get(ctx, dao, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: courseResult.ID}}))
	}
}
