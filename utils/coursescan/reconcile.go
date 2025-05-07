package coursescan

import (
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type OpType string

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	CreateOp  OpType = "create"
	RenameOp  OpType = "rename"
	SwapOp    OpType = "swap"
	ReplaceOp OpType = "replace"
	DeleteOp  OpType = "delete"
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

// RenameAssetOp represents an asset that should be updated due to a rename
// (same hash, different path)
type RenameAssetOp struct {
	Existing *models.Asset
	New      *models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Type implements the Op interface for RenameAssetOp
func (o RenameAssetOp) Type() OpType { return RenameOp }

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
//   - Rename: Different path, same hash → update existing asset path
//   - Replace: Same path, different content (hash) → delete + create
//   - Create: New file with unknown path and hash → create new asset
//   - Swap: Two files exchanged names → delete both and recreate in swapped form
//   - Delete: File no longer on disk → remove from DB
func reconcileAssets(scanned []*models.Asset, existing []*models.Asset) []Op {
	var ops []Op

	seen := map[string]bool{}
	pathMap := map[string]*models.Asset{}
	hashMap := map[string]*models.Asset{}
	scannedPathMap := map[string]*models.Asset{}
	scannedHashMap := map[string]*models.Asset{}

	for _, e := range existing {
		pathMap[e.Path] = e
		hashMap[e.Hash] = e
	}

	for _, s := range scanned {
		scannedPathMap[s.Path] = s
		scannedHashMap[s.Hash] = s
	}

	// First pass: handle create, rename, replace, and noop
	for _, s := range scanned {
		// fmt.Printf("[Scan] Processing: %s (hash=%s)\n", s.Path, s.Hash)

		existingByPath := pathMap[s.Path]
		existingByHash := hashMap[s.Hash]

		// Case 1: No-op
		if existingByPath != nil &&
			existingByPath.FileSize == s.FileSize &&
			existingByPath.ModTime == s.ModTime {
			// fmt.Printf("[No-Op] Match on path+mod+size: %s\n", s.Path)
			s.ID = existingByPath.ID
			seen[existingByPath.ID] = true
			continue
		}

		// Case 2: Rename
		if existingByHash != nil && pathMap[s.Path] == nil {
			// fmt.Printf("[Rename] %s → %s\n", existingByHash.Path, s.Path)
			ops = append(ops, RenameAssetOp{
				Existing: existingByHash,
				New:      s,
			})

			seen[existingByHash.ID] = true
			continue
		}

		// Case 3: Replace
		if existingByPath != nil && existingByHash == nil {
			// fmt.Printf("[Replace] %s (new hash=%s)\n", s.Path, s.Hash)
			ops = append(ops, ReplaceAssetOp{
				Existing: existingByPath,
				New:      s,
			})

			seen[existingByPath.ID] = true
			continue
		}

		// Case 4: Create
		if existingByPath == nil && existingByHash == nil {
			// fmt.Printf("[Create] New asset: %s\n", s.Path)
			ops = append(ops, CreateAssetOp{New: s})
		}
	}

	// Second pass: detect swaps
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

		// Perform swap
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

	// Final pass: deletes
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
