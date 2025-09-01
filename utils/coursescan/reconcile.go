package coursescan

import (
	"fmt"

	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type OpType string

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	NoOp        OpType = "noop"
	CreateOp    OpType = "create"
	UpdateOp    OpType = "update"
	SwapOp      OpType = "swap"
	OverwriteOp OpType = "overwrite"
	ReplaceOp   OpType = "replace"
	DeleteOp    OpType = "delete"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Op represents a generic operation that can be performed to reconcile scanned
// assets or attachments with the existing state in the database. Each implementation
// represents a specific kind of operation such as create, rename, delete, etc
type Op interface {
	Type() OpType
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAssetOp represents the case where a new asset is found on disk that does not
// exist in the database
type CreateAssetOp struct {
	// The new asset to create
	New *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for CreateAssetOp
func (o CreateAssetOp) Type() OpType { return CreateOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAssetOp represents the case when an new asset has the same hash as an existing asset but
// different metadata and so needs to be updated in the database. This is typically occurs following
// a rename
type UpdateAssetOp struct {
	// The existing asset in the DB to update. We will give its ID to the new asset so it
	// can update the existing record and preserve state
	Existing *models.Asset
	// The new asset with changed metadata
	New *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for UpdateAssetOp
func (o UpdateAssetOp) Type() OpType { return UpdateOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// OverwriteAssetOp represents the case where an asset has been renamed, taking the place of an
// another asset
type OverwriteAssetOp struct {
	// The asset that no longer exists (but is still in the DB). We will give its lesson ID
	// to the renamed asset
	Deleted *models.Asset
	// The existing asset that is now taking the place of the deleted asset. We will give its ID
	// to the renamed asset so it can update the existing record and preserve state
	Existing *models.Asset
	// The renamed asset that now takes the place of the deleted asset
	Renamed *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for OverwriteAssetOp
func (o OverwriteAssetOp) Type() OpType { return OverwriteOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ReplaceAssetOp represents the case where an asset is replaced with new contents (new hash)
type ReplaceAssetOp struct {
	// The existing asset to delete
	Existing *models.Asset
	// The new asset to create
	New *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for ReplaceAssetOp
func (o ReplaceAssetOp) Type() OpType { return ReplaceOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SwapAssetOp represents the case when 2 assets are swapped on disk.
type SwapAssetOp struct {
	// The existing asset to delete. We will give its lesson ID to the new asset B
	ExistingA *models.Asset
	// The new asset to create
	NewA *models.Asset
	// The existing asset to delete. We will give its lesson ID to the new asset A
	ExistingB *models.Asset
	// The new asset to create
	NewB *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for SwapAssetOp
func (o SwapAssetOp) Type() OpType { return SwapOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAssetOp represents the case where an asset has been deleted on disk
type DeleteAssetOp struct {
	// The deleted asset that no longer exists on disk but is still in the DB
	Deleted *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for DeleteAssetOp
func (o DeleteAssetOp) Type() OpType { return DeleteOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// reconcileAssets compares the current scanned assets on disk with the existing assets in the
// database and returns a list of operations (Op) that describe how to transition the database
// state to match the disk state
//
// Cases handled:
//   - No-op: Same path, same mod time and file size → do nothing
//   - Create: Completely new path and content → create new asset
//   - Update: Same content, different metadata → update existing asset
//   - Replace: Same metadata, new content → delete and re-create
//   - Overwrite: A known asset was renamed to the path of a now-deleted asset (same hash)
//   - Swap: Two assets exchanged names → delete both and re-create in reverse
//   - Delete: Asset in DB no longer exists on disk → delete it
func reconcileAssets(scanned []*models.Asset, existing []*models.Asset) []Op {
	var ops []Op

	seen := map[string]bool{}
	pathMap := map[string]*models.Asset{}
	hashMap := map[string]*models.Asset{}
	scannedPathMap := map[string]*models.Asset{}

	for _, e := range existing {
		pathMap[e.Path] = e
		hashMap[e.Hash] = e
	}

	for _, s := range scanned {
		scannedPathMap[s.Path] = s
	}

	for _, s := range scanned {
		existingByPath := pathMap[s.Path]
		existingByHash := hashMap[s.Hash]

		// No-op: same path, same mod time and file size
		if existingByPath != nil && existingByPath.FileSize == s.FileSize && existingByPath.ModTime == s.ModTime {
			// fmt.Printf("[No-Op] Match on path+mod+size: %s\n", s.Path)
			s.ID = existingByPath.ID
			seen[existingByPath.ID] = true
			continue
		}

		// Overwrite: content moved onto another file's path
		if existingByHash != nil && existingByPath != nil && existingByHash.ID != existingByPath.ID {
			if other, ok := scannedPathMap[existingByHash.Path]; ok && other.Hash == existingByPath.Hash {
				continue
			}

			// Only overwrite if the original path is not on disk or has a different hash
			if onDisk, ok := scannedPathMap[existingByPath.Path]; !ok || onDisk.Hash != existingByPath.Hash {
				// fmt.Printf("[Overwrite] %s (new hash=%s) → %s\n", existingByHash.Path, s.Hash, s.Path)
				ops = append(ops, OverwriteAssetOp{
					Deleted:  existingByPath,
					Existing: existingByHash,
					Renamed:  s,
				})

				seen[existingByHash.ID] = true
				seen[existingByPath.ID] = true
				continue
			}
		}

		// Update (rename): same hash, new path
		if existingByHash != nil && pathMap[s.Path] == nil {
			// fmt.Printf("[Update (Rename)] %s → %s\n", existingByHash.Path, s.Path)
			ops = append(ops, UpdateAssetOp{
				Existing: existingByHash,
				New:      s,
			})

			seen[existingByHash.ID] = true
			continue
		}

		// Replace: same path, new hash
		if existingByPath != nil && existingByHash == nil {
			// fmt.Printf("[Replace] %s (new hash=%s)\n", s.Path, s.Hash)
			ops = append(ops, ReplaceAssetOp{
				Existing: existingByPath,
				New:      s,
			})

			seen[existingByPath.ID] = true
			continue
		}

		// Create: entirely new
		if existingByPath == nil && existingByHash == nil {
			// fmt.Printf("[Create] New asset: %s\n", s.Path)
			ops = append(ops, CreateAssetOp{New: s})
		}
	}

	// Swap: two assets exchanged names
	processedSwap := map[string]bool{}
	for _, e1 := range existing {
		if seen[e1.ID] || processedSwap[e1.ID] {
			continue
		}

		s1 := scannedPathMap[e1.Path]
		if s1 == nil || s1.Hash == e1.Hash {
			continue
		}

		e2 := hashMap[s1.Hash]
		if e2 == nil || e2.ID == e1.ID {
			continue
		}

		s2 := scannedPathMap[e2.Path]
		if s2 == nil || s2.Hash != e1.Hash {
			continue
		}

		// fmt.Printf("[Swap] %s <=> %s\n", e1.Path, e2.Path)
		ops = append(ops, SwapAssetOp{
			ExistingA: e1,
			ExistingB: e2,
			NewA:      s2,
			NewB:      s1,
		})

		processedSwap[e1.ID] = true
		processedSwap[e2.ID] = true
		seen[e1.ID] = true
		seen[e2.ID] = true
	}

	// Delete: anything not seen yet
	for _, e := range existing {
		if !seen[e.ID] {
			// fmt.Printf("[Delete] Removed asset: %s\n", e.Path)
			ops = append(ops, DeleteAssetOp{Deleted: e})
		}
	}

	return ops
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAttachmentOp represents the case where a new attachment file is found on disk
// that does not exist in the database
type CreateAttachmentOp struct {
	// The new attachment to create
	New *models.Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Type implements the Op interface for CreateAttachmentOp
func (o CreateAttachmentOp) Type() OpType { return CreateOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAttachmentOp represents the case where an attachment file has been deleted on disk
// but still exists in the database
type DeleteAttachmentOp struct {
	// The attachment that no longer exists on disk but is still in the DB
	Deleted *models.Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Type implements the Op interface for DeleteAttachmentOp
func (o DeleteAttachmentOp) Type() OpType { return DeleteOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// reconcileAttachments compares the current scanned attachments on disk with the existing attachments
// in the database and returns a list of operations (Op) that describe how to transition the database
// state to match the disk state
//
// Cases handled:
//   - No-op
//   - Create
//   - Delete
func reconcileAttachments(scanned []*models.Attachment, existing []*models.Attachment) []Op {
	var ops []Op

	seen := map[string]bool{}
	pathMap := map[string]*models.Attachment{}

	for _, e := range existing {
		pathMap[e.Path] = e
	}

	for _, s := range scanned {
		existingByPath := pathMap[s.Path]

		// No-op
		if existingByPath != nil {
			seen[existingByPath.ID] = true
			continue
		}

		// Create
		ops = append(ops, CreateAttachmentOp{New: s})
	}

	// Delete
	for _, e := range existing {
		if !seen[e.ID] {
			ops = append(ops, DeleteAttachmentOp{Deleted: e})
		}
	}

	return ops
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NoLessonOp represents the case where the lesson exists and has not changed. This
// op can be used in the event that there are new assets/attachments being added to the lesson
// and need to be given the lesson ID
type NoLessonOp struct {
	// The new lesson is the assets found on disk. This is can be used to ensure all
	// assets and attachments have the correct lesson ID
	New *models.Lesson

	// The existing lesson that is already in the database
	Existing *models.Lesson
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Type implements the Op interface for NoLessonOp
func (o NoLessonOp) Type() OpType { return NoOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateLessonOp represents the case where a new lesson should be created
type CreateLessonOp struct {
	// The new lesson to create
	New *models.Lesson
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Type implements the Op interface for CreateLessonOp
func (o CreateLessonOp) Type() OpType { return CreateOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateLessonOp represents the case when an existing lesson has changed metadata,
// such as title or description
type UpdateLessonOp struct {
	// The existing lesson in the DB to update. We will give its ID to the new lesson
	// so it can update the existing record
	Existing *models.Lesson
	// The new lesson with changed metadata
	New *models.Lesson
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Type implements the Op interface for UpdateLessonOp
func (o UpdateLessonOp) Type() OpType { return UpdateOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateLessonOp represents the case where an lesson should be deleted
type DeleteLessonOp struct {
	// The existing lesson to delete
	Deleted *models.Lesson
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Type implements the Op interface for DeleteLessonOp
func (o DeleteLessonOp) Type() OpType { return DeleteOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// reconcileLessons compares the current scanned lessons on disk with the existing lessons
// in the database and returns a list of operations (Op) that describe how to transition the database
// state to match the disk state
//
// Cases handled:
//   - No-op
//   - Create
//   - Update
//   - Delete
func reconcileLessons(scannedLessons []*models.Lesson, existingLessons []*models.Lesson) []Op {
	var ops []Op
	seen := map[string]bool{}
	idx := map[string]*models.Lesson{}
	assetToGroup := map[string]*models.Lesson{}

	for _, existingLesson := range existingLessons {
		key := existingLesson.Module + ":" + fmt.Sprint(existingLesson.Prefix.Int16)
		idx[key] = existingLesson

		for _, a := range existingLesson.Assets {
			assetToGroup[a.Hash] = existingLesson
		}
	}

	for _, scannedLesson := range scannedLessons {
		key := scannedLesson.Module + ":" + fmt.Sprint(scannedLesson.Prefix.Int16)

		if existingLesson, ok := idx[key]; ok {
			// Preserve ID
			scannedLesson.ID = existingLesson.ID
			seen[key] = true

			// Update when title has changed
			if scannedLesson.Title != existingLesson.Title {
				// fmt.Printf("[Update Asset Group] %s:%d -> %s \n", scannedLesson.Module, scannedLesson.Prefix.Int16, scannedLesson.Title)
				ops = append(ops, UpdateLessonOp{Existing: existingLesson, New: scannedLesson})
			} else {
				// No-op
				// fmt.Printf("[No-Op Asset Group] %s:%d -> %s \n", scannedLesson.Module, scannedLesson.Prefix.Int16, scannedLesson.Title)
				ops = append(ops, NoLessonOp{New: scannedLesson, Existing: existingLesson})
			}
		} else {
			// Create
			// fmt.Printf("[Create Asset Group] %s:%d -> %s\n", scannedLesson.Module, scannedLesson.Prefix.Int16, scannedLesson.Title)
			ops = append(ops, CreateLessonOp{New: scannedLesson})
		}
	}

	// Delete any unseen lessons
	for _, existingLesson := range existingLessons {
		key := existingLesson.Module + ":" + fmt.Sprint(existingLesson.Prefix.Int16)
		if !seen[key] {
			// fmt.Printf("[Delete Asset Group] %s:%d -> %s\n", existingLesson.Module, existingLesson.Prefix.Int16, existingLesson.Title)
			ops = append(ops, DeleteLessonOp{Deleted: existingLesson})
		}
	}

	return ops
}
