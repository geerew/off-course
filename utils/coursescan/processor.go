package coursescan

// TODO support author.txt/md for course
// TODO support description.txt/md for course

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/media"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type AssetsByChapterPrefix map[string]map[int][]*models.Asset
type AttachmentsByChapterPrefix map[string]map[int][]*models.Attachment

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Processor scans a course to identify assets and attachments
//
// # It can be passed to coursescan.Worker
func Processor(ctx context.Context, s *CourseScan, scan *models.Scan) error {
	if scan == nil {
		return ErrNilScan
	}

	scan.Status.SetProcessing()
	if err := s.dao.UpdateScan(ctx, scan); err != nil {
		return err
	}

	course, err := fetchCourse(ctx, s, scan.CourseID)
	if err != nil || course == nil {
		return err
	}

	// Clear the maintenance mode at the end of the scan
	defer func() {
		if err := clearCourseMaintenance(ctx, s, course); err != nil {
			s.logger.Error("Failed to clear course from maintenance mode", loggerType,
				slog.String("path", course.Path),
			)
		}
	}()

	// Check if the course is available and set its status accordingly
	available, err := checkAndSetCourseAvailability(ctx, s, course)
	if err != nil || !available {
		return err
	}

	if err := enableCourseMaintenance(ctx, s, course); err != nil {
		return err
	}

	// Scan the course directory for files and populate assets and attachments. Also check if there is
	// a course card
	scanned, err := scanFiles(s, course.Path, course.ID)
	if err != nil {
		return err
	}

	// List the assets that already exist in the database for this course

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_COURSE_ID: course.ID})
	existingGroups, err := s.dao.ListAssetGroups(ctx, dbOpts)
	if err != nil {
		return err
	}

	scannedAttachments, scannedAssets := flatAttachmentsAndAssets(scanned.groups)
	existingAttachments, existingAssets := flatAttachmentsAndAssets(existingGroups)

	// Populate hashes of assets that have changed
	if err := populateHashesIfChanged(s.appFs.Fs, scannedAssets, existingAssets); err != nil {
		return err
	}

	// Reconcile what to do
	groupOps := reconcileAssetGroups(scanned.groups, existingGroups)
	assetOps := reconcileAssets(scannedAssets, existingAssets)
	attachmentOps := reconcileAttachments(scannedAttachments, existingAttachments)

	// FFprobe only assets that need it
	videoMetadataByPath := probeVideos(assetOps)

	updatedCourse := course.CardPath != scanned.cardPath
	if updatedCourse {
		course.CardPath = scanned.cardPath
	}

	return s.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		if updated, err := applyAssetGroupCreateUpdateOps(txCtx, s, groupOps); err != nil {
			return err
		} else if updated {
			updatedCourse = true
		}

		if updated, err := applyAttachmentOps(txCtx, s, attachmentOps); err != nil {
			return err
		} else if updated {
			updatedCourse = true
		}

		if updated, err := applyAssetOps(txCtx, s, course, assetOps, videoMetadataByPath); err != nil {
			return err
		} else if updated {
			updatedCourse = true
		}

		if updated, err := applyAssetGroupDeleteOps(txCtx, s, groupOps); err != nil {
			return err
		} else if updated {
			updatedCourse = true
		}

		if updatedCourse {
			course.InitialScan = true
			return s.dao.UpdateCourse(txCtx, course)
		}

		return nil
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// fetchCourse retrieves the course from the database
func fetchCourse(ctx context.Context, s *CourseScan, courseID string) (*models.Course, error) {
	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: courseID})
	course, err := s.dao.GetCourse(ctx, dbOpts)
	if err != nil {
		return nil, err
	}

	if course == nil {
		s.logger.Debug("Ignoring scan job as the course no longer exists",
			loggerType,
			slog.String("course_id", courseID),
		)
		return nil, nil
	}

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// checkAndSetCourseAvailability checks if the course is available and updates its status
// accordingly
func checkAndSetCourseAvailability(ctx context.Context, s *CourseScan, course *models.Course) (bool, error) {
	_, err := s.appFs.Fs.Stat(course.Path)
	if os.IsNotExist(err) {
		s.logger.Debug("Skipping unavailable course", loggerType, slog.String("path", course.Path))

		if course.Available {
			course.Available = false
			return false, s.dao.UpdateCourse(ctx, course)
		}

		return false, nil
	}

	if err != nil {
		return false, err
	}

	if !course.Available {
		course.Available = true
		return true, s.dao.UpdateCourse(ctx, course)
	}

	return true, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// enableCourseMaintenance sets the course to maintenance mode if it is not already
func enableCourseMaintenance(ctx context.Context, s *CourseScan, course *models.Course) error {
	if course.Maintenance {
		return nil
	}

	course.Maintenance = true
	if err := s.dao.UpdateCourse(ctx, course); err != nil {
		return err
	}

	s.logger.Debug("Set course to maintenance mode", loggerType, slog.String("path", course.Path))
	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// clearCourseMaintenance clears the course from maintenance mode if it is set
func clearCourseMaintenance(ctx context.Context, s *CourseScan, course *models.Course) error {
	if !course.Maintenance {
		return nil
	}

	course.Maintenance = false
	if err := s.dao.UpdateCourse(ctx, course); err != nil {
		return err
	}

	s.logger.Debug("Cleared course from maintenance mode", loggerType, slog.String("path", course.Path))
	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// assetGroupBucket accumulates files for a given module & prefix
type assetGroupBucket struct {
	groupedFiles []*parsedFile
	soloFiles    []*parsedFile
	attachFiles  []*parsedFile
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scannedResults holds all asset groups and the card image path
type scannedResults struct {
	groups   []*models.AssetGroup
	cardPath string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanFiles scans the course directory for files. It will return a list of grouped assets,
// and a card path, if found.
func scanFiles(s *CourseScan, coursePath, courseID string) (*scannedResults, error) {
	files, err := s.appFs.ReadDirFlat(coursePath, 2)
	if err != nil {
		return nil, err
	}

	// A bucket is the module => prefix => fileBucket
	buckets := map[string]map[int]*assetGroupBucket{}
	cardPath := ""

	// Scan files on disk and categorize them into buckets
	for _, fp := range files {
		normalizedPath := utils.NormalizeWindowsDrive(fp)
		filename := filepath.Base(normalizedPath)
		dir := filepath.Dir(normalizedPath)
		inRoot := dir == utils.NormalizeWindowsDrive(coursePath)

		module := ""
		if !inRoot {
			module = filepath.Base(dir)
		}

		parsed := parseFilename(normalizedPath, filename)
		category := categorizeFile(parsed)

		if category == Ignore {
			// s.logger.Debug("Ignoring incompatible file", loggerType, slog.String("file", normalized))
			// fmt.Println("[Ignoring]", normalizedPath)
			continue
		}

		if buckets[module] == nil {
			buckets[module] = map[int]*assetGroupBucket{}
		}

		prefix := parsed.Prefix
		if buckets[module][prefix] == nil {
			buckets[module][prefix] = &assetGroupBucket{}
		}

		bucket := buckets[module][prefix]

		// Build the buckets of asset groups of assets, attachments, and descriptions
		switch category {
		case Card:
			// fmt.Println("[Card]", normalizedPath)
			if inRoot && cardPath == "" {
				cardPath = normalizedPath
			}

		case GroupedAsset:
			// fmt.Println("[Grouped Asset]", normalizedPath)
			bucket.groupedFiles = append(bucket.groupedFiles, parsed)

		case Asset:
			// fmt.Println("[Solo Asset]", normalizedPath)
			bucket.soloFiles = append(bucket.soloFiles, parsed)

		case Attachment:
			// fmt.Println("[Attachment]", normalizedPath)
			bucket.attachFiles = append(bucket.attachFiles, parsed)
		}
	}

	var assetGroups []*models.AssetGroup
	for module, prefixMap := range buckets {
		for prefix, bucket := range prefixMap {
			sort.Slice(bucket.groupedFiles, func(i, j int) bool {
				return *bucket.groupedFiles[i].SubPrefix < *bucket.groupedFiles[j].SubPrefix
			})

			assetGroup := &models.AssetGroup{
				CourseID: courseID,
				Module:   module,
				Prefix:   sql.NullInt16{Int16: int16(prefix), Valid: true},
			}

			// Set the asset group title from the first grouped file (if any)
			if len(bucket.groupedFiles) > 0 {
				assetGroup.Title = bucket.groupedFiles[0].Title
			}

			if len(bucket.groupedFiles) > 0 {
				for _, parsedFile := range bucket.groupedFiles {
					asset, err := parsedFile.toAsset(s.appFs.Fs, module, courseID)
					if err != nil {
						return nil, err
					}

					asset.AssetGroupID = assetGroup.ID
					assetGroup.Assets = append(assetGroup.Assets, asset)
				}

				// Demote solo assets to attachments
				for _, parsedFile := range bucket.soloFiles {
					assetGroup.Attachments = append(assetGroup.Attachments, parsedFile.toAttachment())
				}
			} else if len(bucket.soloFiles) > 0 {
				if len(bucket.soloFiles) > 1 {
					// Multiple solo files, pick the best
					idx := pickBest(bucket.soloFiles)
					pf := bucket.soloFiles[idx]

					asset, err := pf.toAsset(s.appFs.Fs, module, courseID)
					if err != nil {
						return nil, err
					}
					asset.AssetGroupID = assetGroup.ID
					assetGroup.Assets = append(assetGroup.Assets, asset)

					// Set the asset group title from the selected solo file
					assetGroup.Title = pf.Title

					// Demote remaining solo files to attachments
					for i, other := range bucket.soloFiles {
						if i == idx {
							continue
						}
						assetGroup.Attachments = append(assetGroup.Attachments, other.toAttachment())
					}
				} else {
					asset, err := bucket.soloFiles[0].toAsset(s.appFs.Fs, module, courseID)
					if err != nil {
						return nil, err
					}
					asset.AssetGroupID = assetGroup.ID
					assetGroup.Assets = append(assetGroup.Assets, asset)

					assetGroup.Title = bucket.soloFiles[0].Title
				}

				// Attachments
				for _, parsedFile := range bucket.attachFiles {
					assetGroup.Attachments = append(assetGroup.Attachments, parsedFile.toAttachment())
				}

			}

			// Create the asset group if it has at least 1 asset
			if len(assetGroup.Assets) > 0 {
				assetGroups = append(assetGroups, assetGroup)
			}
		}
	}

	return &scannedResults{
		groups:   assetGroups,
		cardPath: cardPath,
	}, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// flatAttachmentsAndAssets returns a flat list of attachments and assets from a list of asset groups
func flatAttachmentsAndAssets(groups []*models.AssetGroup) ([]*models.Attachment, []*models.Asset) {
	var attachments []*models.Attachment
	var assets []*models.Asset
	for _, g := range groups {
		assets = append(assets, g.Assets...)
		attachments = append(attachments, g.Attachments...)
	}

	return attachments, assets
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// probeVideos probes videos assets that match the operations create, replace, or swap
func probeVideos(ops []Op) map[string]*models.VideoMetadata {
	var targets []*models.Asset
	for _, op := range ops {
		switch v := op.(type) {
		case CreateAssetOp:
			if v.New.Type.IsVideo() {
				targets = append(targets, v.New)
			}
		case ReplaceAssetOp:
			if v.New.Type.IsVideo() {
				targets = append(targets, v.New)
			}
		case SwapAssetOp:
			if v.NewA.Type.IsVideo() {
				targets = append(targets, v.NewA)
			}
			if v.NewB.Type.IsVideo() {
				targets = append(targets, v.NewB)
			}
		}
	}

	videoMetadataByPath := map[string]*models.VideoMetadata{}
	mediaProbe := media.MediaProbe{}

	for _, asset := range targets {
		if info, err := mediaProbe.ProbeVideo(asset.Path); err == nil {
			videoMetadataByPath[asset.Path] = &models.VideoMetadata{
				VideoMetadataInfo: models.VideoMetadataInfo{
					Duration:   info.Duration,
					Width:      info.Width,
					Height:     info.Height,
					Codec:      info.Codec,
					Resolution: info.Resolution,
				},
			}
		} else {
			// TODO log the error
			// s.logger.Error("Failed to probe video file", loggerType,
			// 	slog.String("path", asset.Path),
			// 	slog.String("error", err.Error()),
			// )
		}
	}

	return videoMetadataByPath
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// applyAssetGroupChanges applies asset group create/update operations. Delete is done after all
// asset and attachment operations have been applied
func applyAssetGroupCreateUpdateOps(
	ctx context.Context,
	s *CourseScan,
	ops []Op,
) (bool, error) {
	if len(ops) == 0 {
		return false, nil
	}

	for _, op := range ops {
		switch v := op.(type) {
		case CreateAssetGroupOp:
			// Create a new asset group
			if err := s.dao.CreateAssetGroup(ctx, v.New); err != nil {
				return false, err
			}

			// fmt.Println("[Created Asset Group]", v.New.ID, v.New.Module, v.New.Prefix.Int16)

			for _, asset := range v.New.Assets {
				asset.AssetGroupID = v.New.ID
			}

			for _, attachments := range v.New.Attachments {
				attachments.AssetGroupID = v.New.ID
			}

		case NoAssetGroupOp:
			// Ensure all new assets and attachments have the asset group ID. When a new
			// asset/attachment is added to an existing asset group, we need to ensure
			// that the asset group ID is set

			// fmt.Println("[No-Op Asset Group]", v.Existing.ID, v.Existing.Module, v.Existing.Prefix.Int16)

			for _, asset := range v.New.Assets {
				asset.AssetGroupID = v.Existing.ID
			}

			for _, attachments := range v.New.Attachments {
				attachments.AssetGroupID = v.Existing.ID
			}

		case UpdateAssetGroupOp:
			// Update an existing asset group by giving the new asset group the ID of the existing
			// asset group, then calling update
			v.New.ID = v.Existing.ID

			if err := s.dao.UpdateAssetGroup(ctx, v.New); err != nil {
				return false, err
			}

			// fmt.Println("[Updated Asset Group]", v.New.ID, v.New.Module, v.New.Prefix.Int16)

			for _, asset := range v.New.Assets {
				asset.AssetGroupID = v.New.ID
			}

			for _, attachments := range v.New.Attachments {
				attachments.AssetGroupID = v.New.ID
			}
		}
	}

	return true, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// applyAssetGroupDeleteOps applies asset group deletion operations
func applyAssetGroupDeleteOps(
	ctx context.Context,
	s *CourseScan,
	ops []Op,
) (bool, error) {
	if len(ops) == 0 {
		return false, nil
	}

	for _, op := range ops {
		switch v := op.(type) {

		case DeleteAssetGroupOp:
			// Delete an existing asset group
			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_GROUP_TABLE_ID: v.Deleted.ID})
			if err := s.dao.DeleteAssetGroups(ctx, dbOpts); err != nil {
				return false, err
			}

			// fmt.Println("[Deleted Asset Group]", v.Deleted.ID, v.Deleted.Module, v.Deleted.Prefix.Int16)
		}
	}

	return true, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// applyAssetOps applies the changes to the assets in the database by creating, renaming,
// replacing, swapping, or deleting them as needed
func applyAssetOps(
	ctx context.Context,
	s *CourseScan,
	course *models.Course,
	ops []Op,
	videoMetadataByPath map[string]*models.VideoMetadata,
) (bool, error) {
	if len(ops) == 0 {
		return false, nil
	}

	for _, op := range ops {
		switch v := op.(type) {
		case CreateAssetOp:
			// Create an asset that was found on disk and does not exist in the database

			// fmt.Println("[Create Asset]", v.New.Path, "->", v.New.Title, v.New.AssetGroupID)

			if err := s.dao.CreateAsset(ctx, v.New); err != nil {
				return false, err
			}

			// If the asset is a video, also create video metadata
			if metadata := videoMetadataByPath[v.New.Path]; metadata != nil && v.New.Type.IsVideo() {
				metadata.AssetID = v.New.ID

				if err := s.dao.CreateVideoMetadata(ctx, metadata); err != nil {
					return false, err
				}

				course.Duration += metadata.Duration
			}

		case UpdateAssetOp:
			// Update an existing asset by giving the new asset the ID of the existing asset and asset group,
			// then calling update
			//
			// This happens when the metadata (title, path, prefix, etc) changes but the
			// contents of the asset have not
			//
			// Asset progress will be preserved

			// fmt.Println("[Update Asset]", v.Existing.Path, "->", v.New.Title, v.New.AssetGroupID)

			v.New.ID = v.Existing.ID

			// When the update is because of a change to `description path`, the actual asset will not
			// have changed and therefore a hash will not have been generated. In this case, just take
			// the existing hash
			if v.New.Hash == "" {
				v.New.Hash = v.Existing.Hash
			}

			if err := s.dao.UpdateAsset(ctx, v.New); err != nil {
				return false, err
			}

		case ReplaceAssetOp:
			// Replace an existing asset with a new asset by first deleting the existing asset, then
			// creating the new asset. Take the existing asset group ID
			//
			// This happens when the contents of an existing asset have changed but the metadata
			// (title, path, prefix, etc) has not
			//
			// Asset progress will be lost

			// fmt.Println("[Replace Asset]", v.Existing.Path, v.Existing.AssetGroupID, "->", v.New.Path, v.New.AssetGroupID)

			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: v.Existing.ID})
			if err := s.dao.DeleteAssets(ctx, dbOpts); err != nil {
				return false, err
			}

			if v.Existing.VideoMetadata != nil {
				course.Duration -= v.Existing.VideoMetadata.Duration
			}

			v.New.AssetGroupID = v.Existing.AssetGroupID
			if err := s.dao.CreateAsset(ctx, v.New); err != nil {
				return false, err
			}

			// If the asset is a video, also create video metadata
			if metadata := videoMetadataByPath[v.New.Path]; metadata != nil && v.New.Type.IsVideo() {
				metadata.AssetID = v.New.ID

				if err := s.dao.CreateVideoMetadata(ctx, metadata); err != nil {
					return false, err
				}

				course.Duration += metadata.Duration
			}

		case OverwriteAssetOp:
			// Overwrite an existing but now deleted asset with another still existing asset by first
			// taking the ID of the existing asset, the asset group ID of the nonexisting asset, then
			// deleting the nonexisting asset, and finally calling update on the renamed asset to update
			// its metadata (title, path, prefix, etc)
			//
			// This happens when an asset has been renamed to that of another, now deleted, asset
			//
			// Asset progress will be preserved

			// fmt.Println("Overwrite Asset", v.Existing.Path, v.Existing.AssetGroupID, "->", v.Renamed.Path, v.Deleted.AssetGroupID)

			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: v.Deleted.ID})
			if err := s.dao.DeleteAssets(ctx, dbOpts); err != nil {
				return false, err
			}

			v.Renamed.ID = v.Existing.ID
			v.Renamed.AssetGroupID = v.Deleted.AssetGroupID

			if err := s.dao.UpdateAsset(ctx, v.Renamed); err != nil {
				return false, err
			}

		case SwapAssetOp:
			// Swap two assets by first deleting the existing assets, then recreating the assets.
			// The asset group IDs of the new assets will be swapped
			//
			// This happens when two existing assets swap paths
			//
			// Asset progress will be lost

			// fmt.Println("Swap Asset", v.ExistingA.Path, "<->", v.ExistingB.Path)

			for _, existing := range []*models.Asset{v.ExistingA, v.ExistingB} {
				dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: existing.ID})
				if err := s.dao.DeleteAssets(ctx, dbOpts); err != nil {
					return false, err
				}

				if existing.VideoMetadata != nil {
					course.Duration -= existing.VideoMetadata.Duration
				}
			}

			// Swap the new asset group IDs
			v.NewA.AssetGroupID = v.ExistingB.AssetGroupID
			v.NewB.AssetGroupID = v.ExistingA.AssetGroupID

			for _, newAsset := range []*models.Asset{v.NewA, v.NewB} {
				if err := s.dao.CreateAsset(ctx, newAsset); err != nil {
					return false, err
				}

				// If the asset is a video, also create video metadata
				if metadata := videoMetadataByPath[newAsset.Path]; metadata != nil && newAsset.Type.IsVideo() {
					metadata.AssetID = newAsset.ID

					if err := s.dao.CreateVideoMetadata(ctx, metadata); err != nil {
						return false, err
					}

					course.Duration += metadata.Duration
				}
			}

		case DeleteAssetOp:
			// Delete an asset that no longer exists on disk

			// fmt.Println("Delete Asset", v.Deleted.Path)

			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: v.Deleted.ID})
			if err := s.dao.DeleteAssets(ctx, dbOpts); err != nil {
				return false, err
			}

			if v.Deleted.VideoMetadata != nil {
				course.Duration -= v.Deleted.VideoMetadata.Duration
			}
		}
	}

	return true, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// applyAttachmentOps applies the changes to the attachments in the database by creating
// or deleting them as needed
func applyAttachmentOps(
	ctx context.Context,
	s *CourseScan,
	ops []Op,

) (bool, error) {
	if len(ops) == 0 {
		return false, nil
	}

	for _, op := range ops {
		switch v := op.(type) {
		case CreateAttachmentOp:
			if err := s.dao.CreateAttachment(ctx, v.New); err != nil {
				return false, err
			}
		case DeleteAttachmentOp:
			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ATTACHMENT_TABLE_ID: v.Deleted.ID})
			if err := s.dao.DeleteAttachments(ctx, dbOpts); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// A priority list for assets when picking a single asset
var assetPriority = []types.AssetType{
	types.AssetVideo,
	types.AssetHTML,
	types.AssetPDF,
	types.AssetMarkdown,
	types.AssetText,
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// pickBest returns the index in candidates of the highest‐priority asset
func pickBest(assets []*parsedFile) int {
	bestIdx := 0
	bestRank := len(assetPriority)
	for i, asset := range assets {
		for rank, at := range assetPriority {
			if asset.AssetType != nil && asset.AssetType.Type() == at {
				if rank < bestRank {
					bestRank = rank
					bestIdx = i
				}

				break
			}
		}
	}

	return bestIdx
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// FileCategory enumerates what parse+classify says we’ve got.
type FileCategory int

const (
	Ignore FileCategory = iota
	Card
	Asset
	GroupedAsset
	Attachment
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// categorizeFile inspects a parsedFile and tells you which it is
func categorizeFile(p *parsedFile) FileCategory {
	// Ignore
	if p == nil {
		return Ignore
	}

	// Card
	if p.IsCard {
		return Card
	}

	// Asset || grouped asset
	if p.AssetType != nil && p.Title != "" {
		if p.SubPrefix != nil {
			return GroupedAsset
		}

		return Asset
	}

	// Attachment
	return Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isCard returns true when the filename is a card
// TODO: make this a type
func isCard(filename string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))

	switch ext {
	case "jpg", "jpeg", "png", "webp", "tiff":
		name := strings.TrimSuffix(filename, "."+ext)
		return name == "card"
	}

	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// populateHashesIfChanged populates the hashes of the scanned assets if they have changed
func populateHashesIfChanged(fs afero.Fs, scanned []*models.Asset, existing []*models.Asset) error {
	existingMap := make(map[string]*models.Asset)
	for _, e := range existing {
		existingMap[e.Path] = e
	}

	for _, s := range scanned {
		e := existingMap[s.Path]
		if e == nil || e.FileSize != s.FileSize || e.ModTime != s.ModTime {
			hash, err := hashFilePartial(fs, s.Path, 1024*1024)
			if err != nil {
				return err
			}
			s.Hash = hash
		}
	}
	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// hashFilePartial computes the SHA-256 hash of a file by reading it in chunks
func hashFilePartial(fs afero.Fs, path string, chunkSize int64) (string, error) {
	file, err := fs.Open(path)
	if err != nil {
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
	}
	size := info.Size()

	hasher := sha256.New()
	readChunk := func(offset int64) error {
		buf := make([]byte, chunkSize)
		_, err := file.Seek(offset, io.SeekStart)
		if err != nil {
			return err
		}
		n, err := io.ReadFull(file, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			return err
		}
		hasher.Write(buf[:n])
		return nil
	}

	// Head
	if err := readChunk(0); err != nil {
		return "", err
	}

	// Middle
	if size > chunkSize*2 {
		middle := size / 2
		if err := readChunk(middle); err != nil {
			return "", err
		}
	}

	// Tail
	if size > chunkSize {
		tail := size - chunkSize
		if err := readChunk(tail); err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getInt safely dereferences *int
func getInt(ptr *int) int {
	if ptr != nil {
		return *ptr
	}
	return 0
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parsedFile represents a parsed file name with its components
type parsedFile struct {
	// 1, 2, 03, etc
	Prefix int

	// Text before `{`
	Title string

	// If {N ...}
	SubPrefix *int

	// If {... subtitle}
	SubTitle string

	// Lowercase extension (without dot)
	Ext string

	// Non-nill when type is one of video, html, pdf, markdown, text
	AssetType *types.Asset

	// True when the file is a card
	IsCard bool

	// The original filename
	Original string

	// Path to the file
	NormalizedPath string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// A regex for parsing a file name into a prefix, title, sub-prefix (optional), sub-title (optional) and extension
var filenameRegex = regexp.MustCompile(
	`^` +
		// Prefix
		//   - Ex: 1, 01, 123
		`(?P<Prefix>\d+)` +

		// Optional spacer followed by an optional title
		//   - Spacer: zero or more spaces, zero or more hyphens, zero or more spaces
		//   - Title: any chars up to “{” (non‐greedy)
		//
		//   - Ex:
		// 		`- title`,
		// 		`  -  title`,
		// 		`title`,
		// 		`--- title`
		`(?:(?:\s+|\s*-+\s*)(?P<Title>[^{]*?))?` +

		// Optional sub‐group in braces { … }
		//   - SubPrefix: any number of digits (non‐greedy)
		//   - Spacer: zero or more spaces, zero or more hyphens, zero or more spaces
		//   - SubTitle: any chars up to `}` (non‐greedy)
		//
		//  - Ex:
		// 		`{2}`,
		// 		`{2 - subtitle}`,
		// 		`{ 2 - subtitle}`,
		// 		`{ 2 }`,
		// 		`{2 - subtitle}`,
		`(?:\s*\{(?:(?P<SubPrefix>\d+)(?:\s*(?:-+\s*)?(?P<SubTitle>[^}]*))?)?\})?` +

		// Optional extension
		//   - Ex: `.jpg`, `.png`, `.pdf`
		`(?:\.(?P<Ext>\w+))?` +

		// End of string
		`$`,
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseFilename parses a filename into its constituent parts
func parseFilename(normalizedPath, filename string) *parsedFile {
	// Quick check for card
	if isCard(filename) {
		return &parsedFile{
			Title:          "card",
			Ext:            strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), ".")),
			AssetType:      types.NewAsset(""),
			IsCard:         true,
			Original:       filename,
			NormalizedPath: normalizedPath,
		}
	}

	matches := filenameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return nil
	}

	// Prefix
	prefix, err := strconv.Atoi(matches[filenameRegex.SubexpIndex("Prefix")])
	if err != nil {
		return nil
	}

	// Title without leading hyphens and spaces
	title := strings.TrimLeft(strings.TrimSpace(matches[filenameRegex.SubexpIndex("Title")]), " -")

	// Ext
	ext := strings.ToLower(matches[filenameRegex.SubexpIndex("Ext")])

	// Asset type
	var assetType *types.Asset
	if ext != "" {
		assetType = types.NewAsset(ext)
	}

	var subPrefix *int
	if sp := matches[filenameRegex.SubexpIndex("SubPrefix")]; sp != "" {
		if v, err := strconv.Atoi(sp); err == nil {
			subPrefix = &v
		}
	}

	subTitle := strings.Trim(strings.TrimSpace(matches[filenameRegex.SubexpIndex("SubTitle")]), " -")

	return &parsedFile{
		Prefix:         prefix,
		Title:          title,
		SubPrefix:      subPrefix,
		SubTitle:       subTitle,
		Ext:            ext,
		AssetType:      assetType,
		IsCard:         false,
		Original:       filename,
		NormalizedPath: normalizedPath,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// toAsset converts a parsedFile to a models.Asset
func (p *parsedFile) toAsset(fs afero.Fs, module, courseID string) (*models.Asset, error) {
	stat, err := fs.Stat(p.NormalizedPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", p.NormalizedPath, err)
	}
	mtime := stat.ModTime().UTC().Format(time.RFC3339Nano)

	asset := &models.Asset{
		CourseID:  courseID,
		Module:    module,
		Title:     p.Title,
		Prefix:    sql.NullInt16{Int16: int16(p.Prefix), Valid: true},
		SubPrefix: sql.NullInt16{Int16: int16(getInt(p.SubPrefix)), Valid: p.SubPrefix != nil},
		SubTitle:  p.SubTitle,
		Type:      *p.AssetType,
		Path:      p.NormalizedPath,
		FileSize:  stat.Size(),
		ModTime:   mtime,
	}

	return asset, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// toAttachment converts a parsedFile to a models.Attachment
func (p *parsedFile) toAttachment() *models.Attachment {
	return &models.Attachment{
		Title: p.Title + filepath.Ext(p.Original),
		Path:  p.NormalizedPath,
	}
}
