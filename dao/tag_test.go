package dao

import (
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)
		tag := &models.Tag{Tag: "Tag 1"}
		require.NoError(t, dao.CreateTag(ctx, tag))
	})

	t.Run("nil", func(t *testing.T) {
		dao, ctx := setup(t)
		require.ErrorIs(t, dao.CreateTag(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("duplicate", func(t *testing.T) {
		dao, ctx := setup(t)

		tag := &models.Tag{Base: models.Base{ID: "1"}, Tag: "Tag 1"}
		require.NoError(t, dao.CreateTag(ctx, tag))

		// Duplicate ID
		tag = &models.Tag{Base: models.Base{ID: "1"}, Tag: "Tag 2"}
		require.ErrorContains(t, dao.CreateTag(ctx, tag), "UNIQUE constraint failed: "+models.TAG_TABLE_ID)

		// Duplicate tag
		tag = &models.Tag{Base: models.Base{ID: "2"}, Tag: "Tag 1"}
		require.ErrorContains(t, dao.CreateTag(ctx, tag), "UNIQUE constraint failed: "+models.TAG_TABLE_TAG)

		// Duplicate tag (case-insensitive)
		tag = &models.Tag{Base: models.Base{ID: "3"}, Tag: "tag 1"}
		require.ErrorContains(t, dao.CreateTag(ctx, tag), "UNIQUE constraint failed: "+models.TAG_TABLE_TAG)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_UpdateTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setup(t)

		originalTag := &models.Tag{Tag: "Tag 1"}
		require.Nil(t, dao.CreateTag(ctx, originalTag))

		time.Sleep(1 * time.Millisecond)

		newTag := &models.Tag{
			Base: originalTag.Base,
			Tag:  "Tag 2", // Mutable
		}
		require.NoError(t, dao.UpdateTag(ctx, newTag))

		tagResult := &models.Tag{}
		require.NoError(t, dao.Get(ctx, tagResult, &database.Options{Where: squirrel.Eq{models.TAG_TABLE_ID: originalTag.ID}}))
		require.Equal(t, originalTag.ID, tagResult.ID)                     // No change
		require.True(t, tagResult.CreatedAt.Equal(originalTag.CreatedAt))  // No change
		require.False(t, tagResult.UpdatedAt.Equal(originalTag.UpdatedAt)) // Changed
		require.Equal(t, newTag.Tag, tagResult.Tag)                        // Changed
	})

	t.Run("invalid", func(t *testing.T) {
		dao, ctx := setup(t)

		tag := &models.Tag{Tag: "Tag 1"}
		require.NoError(t, dao.CreateTag(ctx, tag))

		// Empty ID
		tag.ID = ""
		require.ErrorIs(t, dao.UpdateTag(ctx, tag), utils.ErrInvalidId)

		// Nil Model
		require.ErrorIs(t, dao.UpdateTag(ctx, nil), utils.ErrNilPtr)
	})
}
