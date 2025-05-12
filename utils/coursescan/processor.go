package coursescan

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
	"strconv"
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

type AssetsByChapterPrefix map[string]map[int]*models.Asset
type AttachmentsByChapterPrefix map[string]map[int][]*models.Attachment

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Processor scans a course to identify assets and attachments
//
// It can be passed to coursescan.Worker
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

	available, err := checkAndSetCourseAvailability(ctx, s, course)
	if err != nil || !available {
		return err
	}

	assetsByChapterPrefix, attachmentsByChapterPrefix, cardPath, err := scanFiles(s, course.Path, course.ID)
	if err != nil {
		return err
	}

	scannedAssets := flattenAssets(assetsByChapterPrefix)

	var existingAssets []*models.Asset
	assetOptions := &database.Options{
		Where:            squirrel.Eq{models.ASSET_TABLE_COURSE_ID: course.ID},
		ExcludeRelations: []string{models.ASSET_RELATION_PROGRESS},
	}
	if err = s.dao.ListAssets(ctx, &existingAssets, assetOptions); err != nil {
		return err
	}

	// Populate hashes if changed
	if err := populateHashesIfChanged(s.appFs.Fs, scannedAssets, existingAssets); err != nil {
		return err
	}

	// Reconcile assets
	assetOps := reconcileAssets(scannedAssets, existingAssets)

	// FFprobe only assets that need it
	videoMetadataByPath := map[string]*models.VideoMetadata{}
	mediaProbe := media.MediaProbe{}
	for _, asset := range collectFFProbeTargets(assetOps) {
		if info, err := mediaProbe.ProbeVideo(asset.Path); err == nil {
			videoMetadataByPath[asset.Path] = &models.VideoMetadata{
				Duration:   info.Duration,
				Width:      info.Width,
				Height:     info.Height,
				Codec:      info.Codec,
				Resolution: info.Resolution,
			}
		} else {
			// TODO handle error properly
		}
	}

	updatedCourse := course.CardPath != cardPath
	if updatedCourse {
		course.CardPath = cardPath
	}

	return s.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		updatedAssets, err := applyAssetChanges(txCtx, s, course, assetOps, videoMetadataByPath)
		if err != nil {
			return err
		}

		updatedAttachments, err := applyAttachmentChanges(txCtx, s, assetsByChapterPrefix, attachmentsByChapterPrefix)
		if err != nil {
			return err
		}

		if updatedCourse || updatedAssets || updatedAttachments {
			return s.dao.UpdateCourse(txCtx, course)
		}

		return nil
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// fetchCourse retrieves the course from the database
func fetchCourse(ctx context.Context, s *CourseScan, courseID string) (*models.Course, error) {
	course := &models.Course{}
	options := &database.Options{
		Where:            squirrel.Eq{models.COURSE_TABLE_ID: courseID},
		ExcludeRelations: []string{models.COURSE_RELATION_PROGRESS},
	}

	err := s.dao.GetCourse(ctx, course, options)
	if err == sql.ErrNoRows {
		s.logger.Debug("Ignoring scan job as the course no longer exists",
			loggerType,
			slog.String("course_id", courseID),
		)
		return nil, nil
	}

	return course, err
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

// scanFiles scans the course directory for files. It will return a list of assets,
// attachments and a card path, if found.
func scanFiles(s *CourseScan, coursePath string, courseID string) (AssetsByChapterPrefix, AttachmentsByChapterPrefix, string, error) {
	files, err := s.appFs.ReadDirFlat(coursePath, 2)
	if err != nil {
		return nil, nil, "", err
	}

	assetsByChapterPrefix := make(AssetsByChapterPrefix)
	attachmentsByChapterPrefix := make(AttachmentsByChapterPrefix)
	cardPath := ""

	for _, fp := range files {
		normalized := utils.NormalizeWindowsDrive(fp)
		filename := filepath.Base(normalized)
		dir := filepath.Dir(normalized)
		inRoot := dir == utils.NormalizeWindowsDrive(coursePath)

		if inRoot && isCard(filename) {
			if cardPath == "" {
				cardPath = normalized
			}
			continue
		}

		chapter := ""
		if !inRoot {
			chapter = filepath.Base(dir)
		}

		parsed := parseFilename(filename)
		if parsed == nil {
			s.logger.Debug("Ignoring incompatible file", loggerType, slog.String("file", normalized))
			continue
		}

		if _, ok := assetsByChapterPrefix[chapter]; !ok {
			assetsByChapterPrefix[chapter] = make(map[int]*models.Asset)
		}

		if _, ok := attachmentsByChapterPrefix[chapter]; !ok {
			attachmentsByChapterPrefix[chapter] = make(map[int][]*models.Attachment)
		}

		// Add attachment
		if parsed.asset == nil {
			attachmentsByChapterPrefix[chapter][parsed.prefix] = append(
				attachmentsByChapterPrefix[chapter][parsed.prefix],
				&models.Attachment{
					Title: parsed.title,
					Path:  normalized,
				},
			)

			continue
		}

		prefix := parsed.prefix
		existingAsset := assetsByChapterPrefix[chapter][prefix]

		stat, err := s.appFs.Fs.Stat(normalized)
		if err != nil {
			return nil, nil, "", err
		}

		scannedAsset := &models.Asset{
			Title:    parsed.title,
			Prefix:   sql.NullInt16{Int16: int16(prefix), Valid: true},
			Chapter:  chapter,
			CourseID: courseID,
			Path:     normalized,
			Type:     *parsed.asset,
			FileSize: stat.Size(),
			ModTime:  stat.ModTime().UTC().Format(time.RFC3339Nano),
		}

		// Add new asset
		if existingAsset == nil {
			assetsByChapterPrefix[chapter][prefix] = scannedAsset
			continue
		}

		// Apply asset priority: video > html > pdf
		if scannedAsset.Type.IsVideo() && !existingAsset.Type.IsVideo() ||
			scannedAsset.Type.IsHTML() && existingAsset.Type.IsPDF() {

			scannedAsset.Hash, err = hashFilePartial(s.appFs.Fs, normalized, 1024*1024)
			if err != nil {
				return nil, nil, "", err
			}

			// Downgrade asset to attachment
			attachmentsByChapterPrefix[chapter][prefix] = append(
				attachmentsByChapterPrefix[chapter][prefix],
				&models.Attachment{
					Title: existingAsset.Title + filepath.Ext(existingAsset.Path),
					Path:  existingAsset.Path,
				},
			)

			assetsByChapterPrefix[chapter][prefix] = scannedAsset
		} else {
			// Add the new asset as an attachment
			attachmentsByChapterPrefix[chapter][prefix] = append(
				attachmentsByChapterPrefix[chapter][prefix],
				&models.Attachment{
					Title: parsed.title,
					Path:  normalized,
				},
			)
		}
	}

	return assetsByChapterPrefix, attachmentsByChapterPrefix, cardPath, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func flattenAssets(assetsByChapterPrefix AssetsByChapterPrefix) []*models.Asset {
	var out []*models.Asset
	for _, chapterMap := range assetsByChapterPrefix {
		for _, asset := range chapterMap {
			out = append(out, asset)
		}
	}
	return out
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// collectFFProbeTargets collects the assets that need to be probed by ffprobe based on the
// operations performed on them
func collectFFProbeTargets(ops []Op) []*models.Asset {
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
	return targets
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// applyAssetChanges applies the changes to the assets in the database by creating, renaming,
// replacing, swapping, or deleting them as needed
func applyAssetChanges(
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
			// Create an asset that was found on disk
			fmt.Println("Creating asset", v.New.Path)
			if err := s.dao.CreateAsset(ctx, v.New); err != nil {
				return false, err
			}

			if metadata := videoMetadataByPath[v.New.Path]; metadata != nil && v.New.Type.IsVideo() {
				metadata.AssetID = v.New.ID
				if err := s.dao.CreateVideoMetadata(ctx, metadata); err != nil {
					return false, err
				}
				course.Duration += metadata.Duration
			}

		case RenameAssetOp:
			// Rename an existing asset by giving the new asset the ID of the existing one then call
			// update. This will result in the existing asset being updated with the new prefix, title,
			// path, etc. It also means existing progress will be preserved
			fmt.Println("Renaming asset", v.Existing.Path, "to", v.New.Path)
			v.New.ID = v.Existing.ID
			if err := s.dao.UpdateAsset(ctx, v.New); err != nil {
				return false, err
			}

		case ReplaceAssetOp:
			// Replace an existing asset with a new one by first deleting the existing one, then
			// creating the new one. This happens with an existing asset has been updated, for
			// example a better quality video. Existing progress will be lost, which is perfect
			// because the duration may have changed as well
			fmt.Println("Replacing asset", v.Existing.Path, "with", v.New.Path)
			if err := s.dao.Delete(ctx, v.Existing, nil); err != nil {
				return false, err
			}

			if v.Existing.VideoMetadata != nil {
				course.Duration -= v.Existing.VideoMetadata.Duration
			}

			if err := s.dao.CreateAsset(ctx, v.New); err != nil {
				return false, err
			}

			if metadata := videoMetadataByPath[v.New.Path]; metadata != nil && v.New.Type.IsVideo() {
				metadata.AssetID = v.New.ID
				if err := s.dao.CreateVideoMetadata(ctx, metadata); err != nil {
					return false, err
				}
				course.Duration += metadata.Duration
			}

		case OverwriteRenameOp:
			// Overwrite an existing asset with a new one by first deleting the existing one, then
			// renaming the new one. This happens when a file has been renamed to that of another
			// (existing) asset. The existing asset will be deleted and the new one will take its place.
			// Progress for the rename asset will be preserved
			fmt.Println("Overwriting asset", v.Renamed.Path, "with", v.Deleted.Path)
			if err := s.dao.Delete(ctx, v.Deleted, nil); err != nil {
				return false, err
			}

			if err := s.dao.UpdateAsset(ctx, v.Renamed); err != nil {
				return false, err
			}

		case SwapAssetOp:
			// Swap two assets by first deleting the existing ones, then creating the new ones. This
			// happens when two files have swapped paths on disk. Existing progress will be lost
			fmt.Println("Swapping assets", v.ExistingA.Path, "and", v.ExistingB.Path)
			for _, existing := range []*models.Asset{v.ExistingA, v.ExistingB} {
				if err := s.dao.Delete(ctx, existing, nil); err != nil {
					return false, err
				}

				if existing.VideoMetadata != nil {
					course.Duration -= existing.VideoMetadata.Duration
				}
			}

			for _, newAsset := range []*models.Asset{v.NewA, v.NewB} {
				if err := s.dao.CreateAsset(ctx, newAsset); err != nil {
					return false, err
				}

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
			fmt.Println("Deleting asset", v.Asset.Path)
			if err := s.dao.Delete(ctx, v.Asset, nil); err != nil {
				return false, err
			}

			if v.Asset.VideoMetadata != nil {
				course.Duration -= v.Asset.VideoMetadata.Duration
			}
		}
	}

	return true, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// applyAttachmentChanges applies the changes to the attachments in the database by creating
// or deleting them as needed
func applyAttachmentChanges(
	ctx context.Context,
	s *CourseScan,
	assetsByChapterPrefix AssetsByChapterPrefix,
	attachmentsByChapterPrefix AttachmentsByChapterPrefix,
) (bool, error) {
	attachmentsFlat := []*models.Attachment{}
	for chapter, attachmentMap := range attachmentsByChapterPrefix {
		for prefix, potentialAttachments := range attachmentMap {
			// Only add attachments when there is an asset
			if asset, exists := assetsByChapterPrefix[chapter][prefix]; exists {
				for _, attachment := range potentialAttachments {
					attachment.AssetID = asset.ID
					attachmentsFlat = append(attachmentsFlat, attachment)
				}
			}
		}
	}

	assetIDs := []string{}
	for _, chapterMap := range assetsByChapterPrefix {
		for _, asset := range chapterMap {
			assetIDs = append(assetIDs, asset.ID)
		}
	}

	existing := []*models.Attachment{}
	if err := s.dao.ListAttachments(ctx, &existing, &database.Options{
		Where: squirrel.Eq{models.ATTACHMENT_TABLE_ASSET_ID: assetIDs},
	}); err != nil {
		return false, err
	}

	ops := reconcileAttachments(attachmentsFlat, existing)
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
			if err := s.dao.Delete(ctx, v.Attachment, nil); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parsedFilename that holds information following a filename being parsed
type parsedFilename struct {
	prefix int
	title  string
	asset  *types.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// A regex for parsing a file name into a prefix, title, and extension
//
// Valid patterns:
//
//	 `<prefix>`
//	 `<prefix>.<ext>`
//	 `<prefix> <title>`
//	 `<prefix>-<title>`
//	 `<prefix> - <title>`
//	 `<prefix> <title>.<ext>`
//	 `<prefix>-<title>.<ext>`
//	 `<prefix> - <title>.<ext>`
//
//	- <prefix> is required and must be a number
//	- A dash (-) is optional
//	- <title> is optional and can be any non-empty string
//	- <ext> is optional
var filenameRegex = regexp.MustCompile(`^\s*(?P<Prefix>[0-9]+)((?:\s+-+\s+|\s+-+|\s+|-+\s*)(?P<Title>[^.][^.]*)?)?(?:\.(?P<Ext>\w+))?$`)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseFilename parses a file name and determines if it represents an asset, attachment, or
// neither
//
// Asset: `<prefix> <title>.<ext>` (where <ext> is a valid `types.AssetType`)
// Attachment: `<prefix>`, `<prefix> <title>` or `<prefix> <title>.<ext>` (where <ext> is not a valid `types.AssetType`)
func parseFilename(filename string) *parsedFilename {
	pfn := &parsedFilename{}

	matches := filenameRegex.FindStringSubmatch(filename)
	if len(matches) == 0 {
		return nil
	}

	prefix, err := strconv.Atoi(matches[filenameRegex.SubexpIndex("Prefix")])
	if err != nil {
		return nil
	}

	pfn.prefix = prefix
	pfn.title = matches[filenameRegex.SubexpIndex("Title")]

	// When title is empty, consider this an attachment
	if pfn.title == "" {
		pfn.title = filename
		return pfn
	}

	// Where there is no extension, consider this an attachment
	ext := matches[filenameRegex.SubexpIndex("Ext")]
	if ext == "" {
		return pfn
	}

	pfn.asset = types.NewAsset(ext)

	// When the extension is not supported, consider this an attachment
	if pfn.asset == nil {
		pfn.title = pfn.title + "." + ext
	}

	return pfn
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isCard determines if a given file name represents a card based on its name and extension
func isCard(filename string) bool {
	// Get the extension. If there is no extension, return false
	ext := filepath.Ext(filename)
	if ext == "" {
		return false
	}

	fileWithoutExt := filename[:len(filename)-len(ext)]
	if fileWithoutExt != "card" {
		return false
	}

	// Check if the extension is supported
	switch ext[1:] {
	case
		"jpg",
		"jpeg",
		"png",
		"webp",
		"tiff":
		return true
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
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
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
