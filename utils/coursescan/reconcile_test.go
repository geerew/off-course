package coursescan

import (
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func newAsset(id, path string, size int64, mod, hash string) *models.Asset {
	return &models.Asset{
		Base:     models.Base{ID: id},
		Path:     path,
		FileSize: size,
		ModTime:  mod,
		Hash:     hash,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func newAttachment(id, path string) *models.Attachment {
	return &models.Attachment{
		Base: models.Base{ID: id},
		Path: path,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssets_NoChange(t *testing.T) {
	existing := []*models.Asset{
		newAsset("a1", "/same.mp4", 1024, "t1", "hash123"),
	}
	scanned := []*models.Asset{
		newAsset("", "/same.mp4", 1024, "t1", "has123"),
	}

	ops := reconcileAssets(scanned, existing)
	require.Empty(t, ops)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssets_Create(t *testing.T) {
	existing := []*models.Asset{}
	scanned := []*models.Asset{
		newAsset("", "/foo/video.mp4", 1024, "t1", "hash123"),
	}

	ops := reconcileAssets(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(CreateAssetOp)
	require.True(t, ok, "expected CreateAssetOp")
	assert.Equal(t, CreateOp, op.Type())

	assert.Equal(t, "/foo/video.mp4", op.New.Path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssets_Update(t *testing.T) {
	t.Run("rename", func(t *testing.T) {
		existing := []*models.Asset{
			newAsset("a1", "/old.mp4", 1024, "t1", "hash123"),
		}
		scanned := []*models.Asset{
			newAsset("", "/new.mp4", 1024, "t2", "hash123"),
		}

		ops := reconcileAssets(scanned, existing)
		require.Len(t, ops, 1)

		op, ok := ops[0].(UpdateAssetOp)
		require.True(t, ok)
		assert.Equal(t, UpdateOp, op.Type())

		assert.Equal(t, "a1", op.Existing.ID)
		assert.Equal(t, "/new.mp4", op.New.Path)
	})

	t.Run("description", func(t *testing.T) {
		existing := []*models.Asset{
			newAsset("a1", "/lesson.mp4", 1024, "mod1", "hash123"),
		}
		scanned := []*models.Asset{
			newAsset("", "/lesson.mp4", 1024, "mod1", "hash123"),
		}

		// Set the description
		scanned[0].DescriptionPath = "/path/to/01 description.md"

		ops := reconcileAssets(scanned, existing)
		require.Len(t, ops, 1)

		op, ok := ops[0].(UpdateAssetOp)
		require.True(t, ok, "expected UpdateAssetOp")
		assert.Equal(t, UpdateOp, op.Type())

		// It should carry the existing ID and pick up the new description path
		assert.Equal(t, "a1", op.Existing.ID)
		assert.Equal(t, "/path/to/01 description.md", op.New.DescriptionPath)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssets_Overwrite(t *testing.T) {
	t.Run("forward", func(t *testing.T) {
		existing := []*models.Asset{
			newAsset("a1", "/01 file 1.mp4", 1024, "t1", "hash-abc"),
			newAsset("a2", "/02 file 2.mp4", 1024, "t1", "hash-def"),
		}
		scanned := []*models.Asset{
			newAsset("", "/02 file 2.mp4", 1024, "t2", "hash-abc"),
		}

		ops := reconcileAssets(scanned, existing)
		require.Len(t, ops, 1)

		op, ok := ops[0].(OverwriteAssetOp)
		require.True(t, ok, "expected OverwriteAssetOp")
		assert.Equal(t, OverwriteOp, op.Type())

		assert.Equal(t, "a2", op.Deleted.ID)
		assert.Equal(t, "/02 file 2.mp4", op.Renamed.Path)
	})

	t.Run("reverse", func(t *testing.T) {
		existing := []*models.Asset{
			newAsset("a1", "/01 file 1.mp4", 1024, "t1", "hash-abc"),
			newAsset("a2", "/02 file 2.mp4", 1024, "t1", "hash-def"),
		}
		scanned := []*models.Asset{
			newAsset("", "/01 file 1.mp4", 1024, "t2", "hash-def"),
		}

		ops := reconcileAssets(scanned, existing)
		require.Len(t, ops, 1)

		op, ok := ops[0].(OverwriteAssetOp)
		require.True(t, ok, "expected OverwriteAssetOp")
		assert.Equal(t, OverwriteOp, op.Type())

		assert.Equal(t, "a1", op.Deleted.ID)
		assert.Equal(t, "/01 file 1.mp4", op.Renamed.Path)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssets_Swap(t *testing.T) {
	existing := []*models.Asset{
		newAsset("a1", "/a.mp4", 1024, "t1", "hash1"),
		newAsset("b1", "/b.mp4", 1024, "t1", "hash2"),
	}
	scanned := []*models.Asset{
		newAsset("", "/a.mp4", 1024, "t2", "hash2"),
		newAsset("", "/b.mp4", 1024, "t2", "hash1"),
	}

	ops := reconcileAssets(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(SwapAssetOp)
	require.True(t, ok)
	assert.Equal(t, SwapOp, op.Type())

	// Validate old asset IDs
	oldIDs := []string{op.ExistingA.ID, op.ExistingB.ID}
	assert.ElementsMatch(t, []string{"a1", "b1"}, oldIDs)

	// Validate swapped correctly regardless of order
	swapMap := map[string]string{
		op.NewA.Path: op.NewA.Hash,
		op.NewB.Path: op.NewB.Hash,
	}

	assert.Equal(t, "hash2", swapMap["/a.mp4"])
	assert.Equal(t, "hash1", swapMap["/b.mp4"])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssets_Replace(t *testing.T) {
	existing := []*models.Asset{
		newAsset("a1", "/vid.mp4", 1024, "t1", "oldhash"),
	}
	scanned := []*models.Asset{
		newAsset("", "/vid.mp4", 2048, "t2", "newhash"),
	}

	ops := reconcileAssets(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(ReplaceAssetOp)
	require.True(t, ok)
	assert.Equal(t, ReplaceOp, op.Type())

	assert.Equal(t, "a1", op.Existing.ID)
	assert.Equal(t, "newhash", op.New.Hash)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssets_Delete(t *testing.T) {
	existing := []*models.Asset{
		newAsset("a1", "/gone.mp4", 1024, "t1", "hash123"),
	}
	scanned := []*models.Asset{}

	ops := reconcileAssets(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(DeleteAssetOp)
	require.True(t, ok)
	assert.Equal(t, DeleteOp, op.Type())

	assert.Equal(t, "a1", op.Asset.ID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAttachments_NoChange(t *testing.T) {
	existing := []*models.Attachment{
		newAttachment("a1", "/file1.pdf"),
		newAttachment("a2", "/file2.pdf"),
	}
	scanned := []*models.Attachment{
		newAttachment("", "/file1.pdf"),
		newAttachment("", "/file2.pdf"),
	}

	ops := reconcileAttachments(scanned, existing)
	require.Empty(t, ops)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAttachments_Create(t *testing.T) {
	existing := []*models.Attachment{}
	scanned := []*models.Attachment{
		newAttachment("", "/newfile.pdf"),
	}

	ops := reconcileAttachments(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(CreateAttachmentOp)
	require.True(t, ok)
	assert.Equal(t, CreateOp, op.Type())

	assert.Equal(t, "/newfile.pdf", op.New.Path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAttachments_Delete(t *testing.T) {
	existing := []*models.Attachment{
		newAttachment("a1", "/oldfile.pdf"),
	}
	scanned := []*models.Attachment{}

	ops := reconcileAttachments(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(DeleteAttachmentOp)
	require.True(t, ok)
	assert.Equal(t, DeleteOp, op.Type())

	assert.Equal(t, "a1", op.Attachment.ID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAttachments_MixedOps(t *testing.T) {
	existing := []*models.Attachment{
		newAttachment("a1", "/keep.pdf"),
		newAttachment("a2", "/remove.pdf"),
	}
	scanned := []*models.Attachment{
		newAttachment("", "/keep.pdf"),
		newAttachment("", "/add.pdf"),
	}

	ops := reconcileAttachments(scanned, existing)
	require.Len(t, ops, 2)

	var createFound, deleteFound bool
	for _, op := range ops {
		switch v := op.(type) {
		case CreateAttachmentOp:
			createFound = true
			assert.Equal(t, "/add.pdf", v.New.Path)
		case DeleteAttachmentOp:
			deleteFound = true
			assert.Equal(t, "a2", v.Attachment.ID)
		default:
			t.Errorf("Unexpected op type: %T", op)
		}
	}

	assert.True(t, createFound, "Expected a CreateAttachmentOp")
	assert.True(t, deleteFound, "Expected a DeleteAttachmentOp")
}
