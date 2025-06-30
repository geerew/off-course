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
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)
		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateCourse(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("duplicate", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Base: models.Base{ID: "1"}, Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		// Duplicate ID
		course = &models.Course{Base: models.Base{ID: "1"}, Title: "Course 2", Path: "/course-2"}
		require.ErrorContains(t, dao.CreateCourse(ctx, course), "UNIQUE constraint failed: "+models.COURSE_TABLE_ID)

		// Duplicate Path
		course = &models.Course{Base: models.Base{ID: "2"}, Title: "Course 2", Path: "/course-1"}
		require.ErrorContains(t, dao.CreateCourse(ctx, course), "UNIQUE constraint failed: "+models.COURSE_TABLE_PATH)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		principal := ctx.Value(types.PrincipalContextKey).(types.Principal)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		courseResult := &models.Course{}
		require.NoError(t, dao.GetCourse(ctx, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: course.ID}}))
		require.Equal(t, course.ID, courseResult.ID)
		require.Nil(t, courseResult.Progress)

		assetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

		// Create Asset
		asset := &models.Asset{
			CourseID:     course.ID,
			AssetGroupID: assetGroup.ID,
			Title:        "Asset 1",
			Prefix:       sql.NullInt16{Int16: 1, Valid: true},
			Module:       "Module 1",
			Type:         *types.NewAsset("mp4"),
			Path:         "/course-1/01 asset.mp4",
			FileSize:     1024,
			ModTime:      time.Now().Format(time.RFC3339Nano),
			Hash:         "1234",
		}
		require.NoError(t, dao.CreateAsset(ctx, asset))

		// Create asset progress the user in the current context
		assetProgress := &models.AssetProgress{AssetID: asset.ID}
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress))

		// Create another user
		user2 := &models.User{
			Username:     "user2",
			DisplayName:  "User 2",
			PasswordHash: "hash",
			Role:         types.UserRoleUser,
		}
		require.NoError(t, dao.CreateUser(ctx, user2))

		// Create asset progress for user 2
		principal.UserID = user2.ID
		ctx = context.WithValue(ctx, types.PrincipalContextKey, principal)

		assetProgress2 := &models.AssetProgress{AssetID: asset.ID}
		require.NoError(t, dao.CreateOrUpdateAssetProgress(ctx, course.ID, assetProgress2))

		// Confirm there are 2 asset progress records
		count, err := Count(ctx, dao, &models.AssetProgress{}, nil)
		require.NoError(t, err)
		require.Equal(t, 2, count)

		// Get course with progress and assert the progress is for user 2
		courseResult = &models.Course{}
		require.NoError(t, dao.GetCourse(ctx, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: course.ID}}))
		require.Equal(t, course.ID, courseResult.ID)
		require.NotNil(t, courseResult.Progress)
		require.Equal(t, user2.ID, courseResult.Progress.UserID)
	})

	t.Run("initial scan", func(t *testing.T) {
		dao, ctx := setup(t)

		principal := ctx.Value(types.PrincipalContextKey).(types.Principal)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		// Set the user role to user
		principal.Role = types.UserRoleUser
		ctx = context.WithValue(ctx, types.PrincipalContextKey, principal)

		courseResult := &models.Course{Base: models.Base{ID: course.ID}}
		require.Error(t, dao.GetCourse(ctx, courseResult, nil), sql.ErrNoRows)

		// Set the initial scan to true
		course.InitialScan = true
		require.NoError(t, dao.UpdateCourse(ctx, course))

		// Get the course again
		require.NoError(t, dao.GetCourse(ctx, courseResult, nil))
		require.Equal(t, course.ID, courseResult.ID)
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.GetCourse(ctx, nil, nil), utils.ErrNilPtr)
	})

	t.Run("missing principal", func(t *testing.T) {
		dao, _ := setup(t)

		course := &models.Course{Base: models.Base{ID: "1234"}}
		require.ErrorIs(t, dao.GetCourse(context.Background(), course, nil), utils.ErrMissingPrincipal)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListCourses(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course1 := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course1))

		course2 := &models.Course{Title: "Course 2", Path: "/course-2"}
		require.NoError(t, dao.CreateCourse(ctx, course2))

		courses := []*models.Course{}
		require.NoError(t, dao.ListCourses(ctx, &courses, nil))
		require.Len(t, courses, 2)
	})

	t.Run("initial scan", func(t *testing.T) {
		dao, ctx := setup(t)

		principal := ctx.Value(types.PrincipalContextKey).(types.Principal)

		course1 := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course1))

		course2 := &models.Course{Title: "Course 2", Path: "/course-2"}
		require.NoError(t, dao.CreateCourse(ctx, course2))

		// Set the user role to user
		principal.Role = types.UserRoleUser
		ctx = context.WithValue(ctx, types.PrincipalContextKey, principal)

		courses := []*models.Course{}
		require.NoError(t, dao.ListCourses(ctx, &courses, nil))
		require.Empty(t, courses)

		course1.InitialScan = true
		require.NoError(t, dao.UpdateCourse(ctx, course1))

		courses = []*models.Course{}
		require.NoError(t, dao.ListCourses(ctx, &courses, nil))
		require.Len(t, courses, 1)
		require.Equal(t, course1.ID, courses[0].ID)
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.ListCourses(ctx, nil, nil), utils.ErrNilPtr)
	})

	t.Run("missing principal", func(t *testing.T) {
		dao, _ := setup(t)
		require.ErrorIs(t, dao.ListCourses(context.Background(), &[]*models.Course{}, nil), utils.ErrMissingPrincipal)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateCourse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		originalCourse := &models.Course{Title: "Course 1", Path: "/course-1", Available: true, CardPath: "/course-1/card-1"}
		require.NoError(t, dao.CreateCourse(ctx, originalCourse))

		time.Sleep(1 * time.Millisecond)

		newCourse := &models.Course{
			Base:      originalCourse.Base,
			Title:     "Course 2",         // Immutable
			Path:      "/course-2",        // Immutable
			Available: false,              // Mutable
			CardPath:  "/course-2/card-2", // Mutable
		}
		require.NoError(t, dao.UpdateCourse(ctx, newCourse))

		courseResult := &models.Course{}
		require.NoError(t, dao.GetCourse(ctx, courseResult, &database.Options{Where: squirrel.Eq{models.COURSE_TABLE_ID: originalCourse.ID}}))
		require.Equal(t, originalCourse.ID, courseResult.ID)                     // No change
		require.Equal(t, originalCourse.Title, courseResult.Title)               // No change
		require.Equal(t, originalCourse.Path, courseResult.Path)                 // No change
		require.True(t, courseResult.CreatedAt.Equal(originalCourse.CreatedAt))  // No change
		require.False(t, courseResult.Available)                                 // Changed
		require.Equal(t, newCourse.CardPath, courseResult.CardPath)              // Changed
		require.False(t, courseResult.UpdatedAt.Equal(originalCourse.UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		// Empty ID
		course.ID = ""
		require.ErrorIs(t, dao.UpdateCourse(ctx, course), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateCourse(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ClassifyCoursePaths(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		courses := []*models.Course{}
		for i := range 3 {
			c := &models.Course{
				Title: fmt.Sprintf("Course %d", i),
				Path:  fmt.Sprintf("/course-%d", i),
			}
			require.NoError(t, dao.CreateCourse(ctx, c))
			courses = append(courses, c)
		}

		path1 := "/"                       // ancestor
		path2 := "/test"                   // none
		path3 := courses[2].Path           // course
		path4 := courses[2].Path + "/test" // descendant

		result, err := dao.ClassifyCoursePaths(ctx, []string{path1, path2, path3, path4})
		require.Nil(t, err)

		require.Equal(t, types.PathClassificationAncestor, result[path1])
		require.Equal(t, types.PathClassificationNone, result[path2])
		require.Equal(t, types.PathClassificationCourse, result[path3])
		require.Equal(t, types.PathClassificationDescendant, result[path4])
	})

	t.Run("no paths", func(t *testing.T) {
		dao, ctx := setup(t)

		result, err := dao.ClassifyCoursePaths(ctx, []string{})
		require.Nil(t, err)
		require.Empty(t, result)
	})

	t.Run("empty path", func(t *testing.T) {
		dao, ctx := setup(t)

		result, err := dao.ClassifyCoursePaths(ctx, []string{"", "", ""})
		require.Nil(t, err)
		require.Empty(t, result)
	})

	t.Run("db error", func(t *testing.T) {
		dao, ctx := setup(t)

		_, err := dao.db.Exec("DROP TABLE IF EXISTS " + (&models.Course{}).Table())
		require.Nil(t, err)

		result, err := dao.ClassifyCoursePaths(ctx, []string{"/"})
		require.ErrorContains(t, err, "no such table: "+(&models.Course{}).Table())
		require.Empty(t, result)
	})
}
