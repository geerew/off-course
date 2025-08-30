package coursescan

import (
	"database/sql"
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func newAssetGroup(id, module string, prefix int16, title, descPath, descType string) *models.AssetGroup {
	ag := &models.AssetGroup{
		Base:            models.Base{ID: id},
		Module:          module,
		Prefix:          sql.NullInt16{Int16: prefix, Valid: true},
		Title:           title,
		DescriptionPath: descPath,
	}
	if dt := types.NewDescription(descType); dt != nil {
		ag.DescriptionType = *dt
	}
	return ag
}

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

func TestReconcileAssetGroups_NoOp(t *testing.T) {
	existing := []*models.AssetGroup{
		newAssetGroup("g1", "mod1", 1, "Title", "path.md", "md"),
	}
	scanned := []*models.AssetGroup{
		newAssetGroup("", "mod1", 1, "Title", "path.md", "md"),
	}

	ops := reconcileAssetGroups(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(NoAssetGroupOp)
	require.True(t, ok)
	assert.Equal(t, NoOp, op.Type())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssetGroups_Create(t *testing.T) {
	existing := []*models.AssetGroup{}
	scanned := []*models.AssetGroup{
		newAssetGroup("", "mod2", 2, "NewTitle", "new.md", "md"),
	}

	ops := reconcileAssetGroups(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(CreateAssetGroupOp)
	require.True(t, ok)
	assert.Equal(t, CreateOp, op.Type())
	assert.Equal(t, "mod2", op.New.Module)
	assert.Equal(t, int16(2), op.New.Prefix.Int16)
	assert.Equal(t, "NewTitle", op.New.Title)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssetGroups_Update(t *testing.T) {
	existing := []*models.AssetGroup{
		newAssetGroup("g2", "mod3", 3, "OldTitle", "old.md", "md"),
	}
	scanned := []*models.AssetGroup{
		newAssetGroup("", "mod3", 3, "NewTitle", "new.md", "md"),
	}

	ops := reconcileAssetGroups(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(UpdateAssetGroupOp)
	require.True(t, ok)
	assert.Equal(t, UpdateOp, op.Type())
	assert.Equal(t, "g2", op.Existing.ID)
	assert.Equal(t, "NewTitle", op.New.Title)
	assert.Equal(t, "new.md", op.New.DescriptionPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssetGroups_Delete(t *testing.T) {
	existing := []*models.AssetGroup{
		newAssetGroup("g3", "mod4", 4, "Title4", "desc.md", "md"),
	}
	scanned := []*models.AssetGroup{}

	ops := reconcileAssetGroups(scanned, existing)
	require.Len(t, ops, 1)

	op, ok := ops[0].(DeleteAssetGroupOp)
	require.True(t, ok)
	assert.Equal(t, DeleteOp, op.Type())
	assert.Equal(t, "g3", op.Deleted.ID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestReconcileAssetGroups_Mixed(t *testing.T) {
	// existing groups g1(mod1,1), g2(mod2,2), g3(mod3,3)
	existing := []*models.AssetGroup{
		newAssetGroup("g1", "mod1", 1, "T1", "d1.md", "md"),
		newAssetGroup("g2", "mod2", 2, "T2", "d2.md", "md"),
		newAssetGroup("g3", "mod3", 3, "T3", "d3.md", "md"),
	}
	// scanned: keep mod1, update mod2, create mod4
	scanned := []*models.AssetGroup{
		newAssetGroup("", "mod1", 1, "T1", "d1.md", "md"),
		newAssetGroup("", "mod2", 2, "T2-new", "d2-new.md", "md"),
		newAssetGroup("", "mod4", 4, "T4", "d4.md", "md"),
	}

	op := reconcileAssetGroups(scanned, existing)
	require.Len(t, op, 4)

	var create, update, del, noop bool
	for _, e := range op {
		switch v := e.(type) {
		case CreateAssetGroupOp:
			create = true
			assert.Equal(t, "mod4", v.New.Module)
		case UpdateAssetGroupOp:
			update = true
			assert.Equal(t, "g2", v.Existing.ID)
			assert.Equal(t, "T2-new", v.New.Title)
		case DeleteAssetGroupOp:
			del = true
			assert.Equal(t, "g3", v.Deleted.ID)
		case NoAssetGroupOp:
			noop = true
			assert.Equal(t, "g1", v.Existing.ID)
		default:
			t.Errorf("unexpected op %T", v)
		}
	}

	assert.True(t, create, "expected CreateOp")
	assert.True(t, update, "expected UpdateOp")
	assert.True(t, del, "expected DeleteOp")
	assert.True(t, noop, "expected NoAssetGroupOp")
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

	assert.Equal(t, "a1", op.Deleted.ID)
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

	assert.Equal(t, "a1", op.Deleted.ID)
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
			assert.Equal(t, "a2", v.Deleted.ID)
		default:
			t.Errorf("Unexpected op type: %T", op)
		}
	}

	assert.True(t, createFound, "Expected a CreateAttachmentOp")
	assert.True(t, deleteFound, "Expected a DeleteAttachmentOp")
}
