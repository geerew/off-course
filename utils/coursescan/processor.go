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
	"github.com/geerew/off-course/dao"
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
	scannedResults, err := scanFiles(s, course.Path, course.ID)
	if err != nil {
		return err
	}

	scannedAssets := flattenAssets(scannedResults.assetsByChapterPrefix)

	// List the assets that already exist in the database for this course
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
			s.logger.Error("Failed to probe video file", loggerType,
				slog.String("path", asset.Path),
				slog.String("error", err.Error()),
			)
		}
	}

	updatedCourse := course.CardPath != scannedResults.cardPath
	if updatedCourse {
		course.CardPath = scannedResults.cardPath
	}

	return s.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		updatedAssets, err := applyAssetChanges(txCtx, s, course, assetOps, videoMetadataByPath)
		if err != nil {
			return err
		}

		updatedAttachments, err := applyAttachmentChanges(txCtx, s, scannedResults.assetsByChapterPrefix, scannedResults.attachmentsByChapterPrefix)
		if err != nil {
			return err
		}

		if updatedCourse || updatedAssets || updatedAttachments {
			course.InitialScan = true
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

type scannedResults struct {
	assetsByChapterPrefix      AssetsByChapterPrefix
	attachmentsByChapterPrefix AttachmentsByChapterPrefix
	cardPath                   string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanFiles scans the course directory for files. It will return a list of assets,
// attachments and a card path, if found.
func scanFiles(s *CourseScan, coursePath string, courseID string) (*scannedResults, error) {
	files, err := s.appFs.ReadDirFlat(coursePath, 2)
	if err != nil {
		return nil, err
	}

	buckets := make(map[string]map[int]*fileBucket)
	cardPath := ""

	for _, fp := range files {
		normalized := utils.NormalizeWindowsDrive(fp)
		filename := filepath.Base(normalized)
		dir := filepath.Dir(normalized)
		inRoot := dir == utils.NormalizeWindowsDrive(coursePath)

		chapter := ""
		if !inRoot {
			chapter = filepath.Base(dir)
		}

		parsed := parseFilename(filename)
		category := categorizeFile(parsed)

		if category == Ignore {
			// Ignore files that do not match any category
			fmt.Println("[Ignoring]", normalized)
			s.logger.Debug("Ignoring incompatible file", loggerType, slog.String("file", normalized))
			continue
		}

		if buckets[chapter] == nil {
			buckets[chapter] = make(map[int]*fileBucket)
		}

		var bucket *fileBucket
		if bucket = buckets[chapter][parsed.Prefix]; bucket == nil {
			bucket = &fileBucket{}
			buckets[chapter][parsed.Prefix] = bucket
		}

		// Build the buckets of assets, grouped assets, attachments, and descriptions
		switch category {
		case Card:
			fmt.Println("[Card]", normalized)

			if inRoot && cardPath == "" {
				cardPath = normalized
			}

		case Description:
			fmt.Println("[Description]", normalized)
			if bucket.descriptionPath == "" {
				bucket.descriptionPath = normalized
			} else {
				s.logger.Warn("Multiple description files found, ignoring", loggerType,
					slog.String("file", normalized),
					slog.String("existing", bucket.descriptionPath),
				)
			}

		case Asset:
			fmt.Println("[Asset]", normalized)

			stat, err := s.appFs.Fs.Stat(fp)
			if err != nil {
				return nil, err
			}

			bucket.assets = append(bucket.assets, &models.Asset{
				Title:    parsed.Title,
				Prefix:   sql.NullInt16{Int16: int16(parsed.Prefix), Valid: true},
				Chapter:  chapter,
				CourseID: courseID,
				Path:     normalized,
				Type:     *parsed.AssetType,
				FileSize: stat.Size(),
				ModTime:  stat.ModTime().UTC().Format(time.RFC3339Nano),
			})

		case GroupedAsset:
			fmt.Println("[Grouped Asset]", normalized)

			stat, err := s.appFs.Fs.Stat(fp)
			if err != nil {
				return nil, err
			}

			bucket.groupedAssets = append(bucket.groupedAssets, &models.Asset{
				Title:     parsed.Title,
				Prefix:    sql.NullInt16{Int16: int16(parsed.Prefix), Valid: true},
				SubPrefix: sql.NullInt16{Int16: int16(*parsed.SubPrefix), Valid: true},
				SubTitle:  parsed.SubTitle,
				Chapter:   chapter,
				CourseID:  courseID,
				Path:      normalized,
				Type:      *parsed.AssetType,
				FileSize:  stat.Size(),
				ModTime:   stat.ModTime().UTC().Format(time.RFC3339Nano),
			})

		case Attachment:
			fmt.Println("[Attachment]", normalized)

			bucket.attachments = append(bucket.attachments, &models.Attachment{
				Title: parsed.Title + filepath.Ext(normalized),
				Path:  normalized,
			})
		}
	}

	results := &scannedResults{
		assetsByChapterPrefix:      make(AssetsByChapterPrefix),
		attachmentsByChapterPrefix: make(AttachmentsByChapterPrefix),
		cardPath:                   cardPath,
	}

	for chapter, prefixMap := range buckets {
		results.assetsByChapterPrefix[chapter] = make(map[int][]*models.Asset)
		results.attachmentsByChapterPrefix[chapter] = make(map[int][]*models.Attachment)

		for prefix, bucket := range prefixMap {
			if len(bucket.groupedAssets) > 0 {
				// There are grouped assets, meaning we set demote non-grouped assets to attachments

				// Sort the grouped assets by sub-prefix and set
				sort.Slice(bucket.groupedAssets, func(i, j int) bool {
					return bucket.groupedAssets[i].SubPrefix.Int16 < bucket.groupedAssets[j].SubPrefix.Int16
				})
				results.assetsByChapterPrefix[chapter][prefix] = bucket.groupedAssets

				// Demote non-grouped assets to attachments
				for _, asset := range bucket.assets {
					results.attachmentsByChapterPrefix[chapter][prefix] = append(
						results.attachmentsByChapterPrefix[chapter][prefix],
						&models.Attachment{
							Title: asset.Title + filepath.Ext(asset.Path),
							Path:  asset.Path,
						},
					)
				}

				// Add attachments
				results.attachmentsByChapterPrefix[chapter][prefix] = append(
					results.attachmentsByChapterPrefix[chapter][prefix],
					bucket.attachments...)

				// Description (Set to the first asset in the group)
				if bucket.descriptionPath != "" {
					ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(bucket.descriptionPath), "."))
					if descriptionType := types.NewDescription(ext); descriptionType != nil && descriptionType.IsSupported() {
						results.assetsByChapterPrefix[chapter][prefix][0].DescriptionType = *descriptionType
						results.assetsByChapterPrefix[chapter][prefix][0].DescriptionPath = bucket.descriptionPath
					} else {
						s.logger.Warn("Ignoring incompatible description file", loggerType,
							slog.String("file", bucket.descriptionPath),
							slog.String("type", ext),
						)
					}
				}
			} else if len(bucket.assets) > 0 {
				// There are only non-grouped assets
				priorityIndex := pickBest(bucket.assets)
				results.assetsByChapterPrefix[chapter][prefix] = []*models.Asset{bucket.assets[priorityIndex]}

				// Demote the other assets to attachments
				for i, asset := range bucket.assets {
					if i == priorityIndex {
						continue
					}

					results.attachmentsByChapterPrefix[chapter][prefix] = append(
						results.attachmentsByChapterPrefix[chapter][prefix],
						&models.Attachment{
							Title: asset.Title + filepath.Ext(asset.Path),
							Path:  asset.Path,
						},
					)
				}

				// Add attachments
				results.attachmentsByChapterPrefix[chapter][prefix] = append(
					results.attachmentsByChapterPrefix[chapter][prefix],
					bucket.attachments...)

				// Description
				if bucket.descriptionPath != "" {
					// Get the ext of the description file
					ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(bucket.descriptionPath), "."))
					if descriptionType := types.NewDescription(ext); descriptionType != nil && descriptionType.IsSupported() {
						results.assetsByChapterPrefix[chapter][prefix][0].DescriptionType = *descriptionType
						results.assetsByChapterPrefix[chapter][prefix][0].DescriptionPath = bucket.descriptionPath
					} else {
						s.logger.Warn("Ignoring incompatible description file", loggerType,
							slog.String("file", bucket.descriptionPath),
							slog.String("type", ext),
						)
					}
				}
			}
		}
	}

	return results, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func flattenAssets(assetsByChapterPrefix AssetsByChapterPrefix) []*models.Asset {
	var out []*models.Asset
	for _, chapterMap := range assetsByChapterPrefix {
		for _, asset := range chapterMap {
			out = append(out, asset...)
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
			// Create an asset that was found on disk and does not exist in the database
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
			// Update an existing asset by giving the new asset the ID of the existing asset, then calling
			// update
			//
			// This happens when the metadata (title, path, prefix, description, etc) changes but the
			// contents of the asset have not
			//
			// Asset progress will be preserved
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
			// creating the new asset
			//
			// This happens when the contents of an existing asset have changed but the metadata
			// (title, path, prefix, etc) has not
			//
			// Asset progress will be lost
			if err := dao.Delete(ctx, s.dao, v.Existing, nil); err != nil {
				return false, err
			}

			if v.Existing.VideoMetadata != nil {
				course.Duration -= v.Existing.VideoMetadata.Duration
			}

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
			// taking the ID of the deleted asset, then deleting the nonexisting asset, and finally calling
			// update on the renamed asset to update its metadata (title, path, prefix, etc)
			//
			// This happens when an asset has been renamed to that of another, now deleted, asset
			//
			// Asset progress will be preserved
			if err := dao.Delete(ctx, s.dao, v.Deleted, nil); err != nil {
				return false, err
			}

			v.Renamed.ID = v.Existing.ID
			if err := s.dao.UpdateAsset(ctx, v.Renamed); err != nil {
				return false, err
			}

		case SwapAssetOp:
			// Swap two assets by first deleting the existing assets, then recreating the assets
			//
			// This happens when two existing assets swap paths
			//
			// Asset progress will be lost
			for _, existing := range []*models.Asset{v.ExistingA, v.ExistingB} {
				if err := dao.Delete(ctx, s.dao, existing, nil); err != nil {
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
			if err := dao.Delete(ctx, s.dao, v.Asset, nil); err != nil {
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
			if assets, exists := assetsByChapterPrefix[chapter][prefix]; exists && len(assets) > 0 {
				// Attach it to the first asset in the chapter with the prefix
				assetId := assets[0].ID

				for _, attachment := range potentialAttachments {
					attachment.AssetID = assetId
					attachmentsFlat = append(attachmentsFlat, attachment)
				}
			}
		}
	}

	assetIDs := []string{}
	for _, chapterMap := range assetsByChapterPrefix {
		for _, assets := range chapterMap {
			// Always just use the first asset in the chapter with the prefix
			assetIDs = append(assetIDs, assets[0].ID)
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
			if err := dao.Delete(ctx, s.dao, v.Attachment, nil); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileBucket struct {
	groupedAssets   []*models.Asset
	assets          []*models.Asset
	attachments     []*models.Attachment
	descriptionPath string
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
func pickBest(candidates []*models.Asset) int {
	bestIndex := 0
	bestRank := len(assetPriority)

	for i, a := range candidates {
		for rank, at := range assetPriority {
			if a.Type.Type() == at && rank < bestRank {
				bestRank = rank
				bestIndex = i
				break
			}
		}
	}
	return bestIndex
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
func parseFilename(filename string) *parsedFile {
	// Quick check for card
	if isCard(filename) {
		return &parsedFile{
			Prefix:    0, // Card has no prefix
			Title:     "card",
			SubPrefix: nil,
			SubTitle:  "",
			Ext:       strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), ".")),
			AssetType: types.NewAsset(""),
			IsCard:    true,
			Original:  filename,
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
		Prefix:    prefix,
		Title:     title,
		SubPrefix: subPrefix,
		SubTitle:  subTitle,
		Ext:       ext,
		AssetType: assetType,
		IsCard:    false,
		Original:  filename,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// FileCategory enumerates what parse+classify says we’ve got.
type FileCategory int

const (
	Ignore FileCategory = iota
	Card
	Description
	Asset
	GroupedAsset
	Attachment
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// categorize inspects a parsedFile and tells you which it is
func categorizeFile(p *parsedFile) FileCategory {
	// Ignore
	if p == nil {
		return Ignore
	}

	// Card
	if p.IsCard {
		return Card
	}

	// Description
	if p.SubPrefix == nil && strings.EqualFold(p.Title, "description") {
		if descriptionType := types.NewDescription(strings.ToLower(p.Ext)); descriptionType != nil {
			return Description
		}
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

// isCard returns true for card.(jpg|jpeg|png|webp|tiff)
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
			// TODO tmp

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
