package coursescan

import (
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type OpType string

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
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

// CreateAssetOp represents a new asset that should be created in the database
type CreateAssetOp struct {
	New *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for CreateAssetOp
func (o CreateAssetOp) Type() OpType { return CreateOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAssetOp represents an asset that should be updated in the database
type UpdateAssetOp struct {
	Existing *models.Asset
	New      *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for UpdateAssetOp
func (o UpdateAssetOp) Type() OpType { return UpdateOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// OverwriteAssetOp represents a case where an asset has been renamed to take the place of an
// existing but now deleted asset (same hash, different path and a deleted path)
type OverwriteAssetOp struct {
	Deleted  *models.Asset
	Existing *models.Asset
	Renamed  *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for OverwriteAssetOp
func (o OverwriteAssetOp) Type() OpType { return OverwriteOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ReplaceAssetOp represents a case where an asset at a known path has changed
// content (different hash), and should be deleted and replaced
type ReplaceAssetOp struct {
	Existing *models.Asset
	New      *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for ReplaceAssetOp
func (o ReplaceAssetOp) Type() OpType { return ReplaceOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SwapAssetOp represents two assets that have swapped file paths on disk.
// Because paths are unique, both must be deleted and re-created in reverse
type SwapAssetOp struct {
	ExistingA *models.Asset
	ExistingB *models.Asset
	NewA      *models.Asset
	NewB      *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for SwapAssetOp
func (o SwapAssetOp) Type() OpType { return SwapOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAssetOp represents an asset that exists in the DB but not on disk
// and should therefore be removed
type DeleteAssetOp struct {
	Asset *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for DeleteAssetOp
func (o DeleteAssetOp) Type() OpType { return DeleteOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAttachmentOp represents a new attachment file that should be created
type CreateAttachmentOp struct {
	New *models.Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Type implements the Op interface for CreateAttachmentOp
func (o CreateAttachmentOp) Type() OpType { return CreateOp }

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAttachmentOp represents an attachment file that should be removed
type DeleteAttachmentOp struct {
	Attachment *models.Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Type implements the Op interface for DeleteAttachmentOp
func (o DeleteAttachmentOp) Type() OpType { return DeleteOp }

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

		// Update: same path, same size and mod time, but different description path
		if existingByPath != nil && existingByPath.FileSize == s.FileSize && existingByPath.ModTime == s.ModTime {
			if existingByPath.DescriptionPath != s.DescriptionPath {
				// fmt.Printf("[Update] %s (new description=%s)\n", s.Path, s.DescriptionPath)
				ops = append(ops, UpdateAssetOp{
					Existing: existingByPath,
					New:      s,
				})

				seen[existingByPath.ID] = true
				continue
			}
		}

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
			ops = append(ops, DeleteAssetOp{Asset: e})
		}
	}

	return ops
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// reconcileAttachments compares the current scanned attachments on disk with the existing attachments
// in the database and returns a list of operations (Op) that describe how to transition the database
// state to match the disk state
//
// Cases handled:
//   - No-op: Same path → do nothing
//   - Create: New file with unknown path → create new attachment
//   - Delete: File no longer on disk → remove from DB
func reconcileAttachments(scanned []*models.Attachment, existing []*models.Attachment) []Op {
	var ops []Op

	seen := map[string]bool{}
	pathMap := map[string]*models.Attachment{}

	for _, e := range existing {
		pathMap[e.Path] = e
	}

	for _, s := range scanned {
		existingByPath := pathMap[s.Path]

		// Case 1: Do nothing
		if existingByPath != nil {
			seen[existingByPath.ID] = true
			continue
		}

		// Case 2: Create
		ops = append(ops, CreateAttachmentOp{New: s})
	}

	// Case 3: Delete
	for _, e := range existing {
		if !seen[e.ID] {
			ops = append(ops, DeleteAttachmentOp{Attachment: e})
		}
	}

	return ops
}
