package dao

import (
	"database/sql"
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateAttachment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

		attachment := &models.Attachment{
			AssetGroupID: assetGroup.ID,
			Title:        "Attachment 1",
			Path:         "/course-1/01 attachment.txt",
		}
		require.NoError(t, dao.CreateAttachment(ctx, attachment))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateAttachment(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateAttachment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}

		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

		originalAttachment := &models.Attachment{
			AssetGroupID: assetGroup.ID,
			Title:        "Attachment 1",
			Path:         "/course-1/01 Attachment 1.txt",
		}
		require.NoError(t, dao.CreateAttachment(ctx, originalAttachment))

		time.Sleep(1 * time.Millisecond)

		newAttachment := &models.Attachment{
			Base:         originalAttachment.Base,
			AssetGroupID: assetGroup.ID,                   // Immutable
			Title:        "Attachment 2",                  // Mutable
			Path:         "/course-1/01 Attachment 2.txt", // Mutable
		}
		require.NoError(t, dao.UpdateAttachment(ctx, newAttachment))

		attachmentResult := &models.Attachment{Base: models.Base{ID: originalAttachment.ID}}
		require.NoError(t, dao.GetAttachment(ctx, attachmentResult, nil))
		require.Equal(t, newAttachment.ID, attachmentResult.ID)                          // No change
		require.Equal(t, newAttachment.AssetGroupID, attachmentResult.AssetGroupID)      // No change
		require.True(t, newAttachment.CreatedAt.Equal(originalAttachment.CreatedAt))     // No change
		require.Equal(t, newAttachment.Title, attachmentResult.Title)                    // Changed
		require.Equal(t, newAttachment.Path, attachmentResult.Path)                      // Changed
		require.False(t, attachmentResult.UpdatedAt.Equal(originalAttachment.UpdatedAt)) // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		course := &models.Course{Title: "Course 1", Path: "/course-1"}
		require.NoError(t, dao.CreateCourse(ctx, course))

		assetGroup := &models.AssetGroup{
			CourseID: course.ID,
			Title:    "Asset Group 1",
			Prefix:   sql.NullInt16{Int16: 1, Valid: true},
			Module:   "Module 1",
		}
		require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

		attachment := &models.Attachment{
			AssetGroupID: assetGroup.ID,
			Title:        "Attachment 1",
			Path:         "/course-1/01 attachment.txt",
		}
		require.NoError(t, dao.CreateAttachment(ctx, attachment))

		// Empty ID
		attachment.ID = ""
		require.ErrorIs(t, dao.UpdateAttachment(ctx, attachment), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateAttachment(ctx, nil), utils.ErrNilPtr)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_AttachmentDeleteCascade(t *testing.T) {
	dao, ctx := setup(t)

	course := &models.Course{Title: "Course", Path: "/course"}
	require.NoError(t, dao.CreateCourse(ctx, course))

	assetGroup := &models.AssetGroup{
		CourseID: course.ID,
		Title:    "Asset Group 1",
		Prefix:   sql.NullInt16{Int16: 1, Valid: true},
		Module:   "Module 1",
	}
	require.NoError(t, dao.CreateAssetGroup(ctx, assetGroup))

	attachment := &models.Attachment{
		AssetGroupID: assetGroup.ID,
		Title:        "Attachment 1",
		Path:         "/course-1/01 attachment.txt",
	}
	require.NoError(t, dao.CreateAttachment(ctx, attachment))

	require.Nil(t, Delete(ctx, dao, assetGroup, nil))

	count, err := Count(ctx, dao, &models.Attachment{}, nil)
	require.NoError(t, err)
	require.Zero(t, count)
}
