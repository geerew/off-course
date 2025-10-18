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
	"github.com/geerew/off-course/utils/media/probe"
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

	utils.Infof("CourseScan: Starting scan for course ID: %s\n", scan.CourseID)

	scan.Status.SetProcessing()
	if err := s.dao.UpdateScan(ctx, scan); err != nil {
		return err
	}

	course, err := fetchCourse(ctx, s, scan.CourseID)
	if err != nil || course == nil {
		return err
	}

	utils.Infof("CourseScan: Found course '%s' at path: %s\n", course.Title, course.Path)

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
	utils.Infof("CourseScan: Scanning course directory: %s\n", course.Path)
	scanned, err := scanFiles(s, course.Path, course.ID)
	if err != nil {
		return err
	}

	utils.Infof("CourseScan: Found %d lessons, card path: %s\n", len(scanned.lessons), scanned.cardPath)

	// List the assets that already exist in the database for this course

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_COURSE_ID: course.ID})
	existingGroups, err := s.dao.ListLessons(ctx, dbOpts)
	if err != nil {
		return err
	}

	scannedAttachments, scannedAssets := flatAttachmentsAndAssets(scanned.lessons)
	existingAttachments, existingAssets := flatAttachmentsAndAssets(existingGroups)

	// Populate hashes of assets that have changed
	if err := populateHashesIfChanged(s.appFs.Fs, scannedAssets, existingAssets); err != nil {
		return err
	}

	// Reconcile what to do
	utils.Infof("CourseScan: Reconciling changes - %d scanned assets, %d existing assets\n", len(scannedAssets), len(existingAssets))
	groupOps := reconcileLessons(scanned.lessons, existingGroups)
	assetOps := reconcileAssets(scannedAssets, existingAssets)
	attachmentOps := reconcileAttachments(scannedAttachments, existingAttachments)

	utils.Infof("CourseScan: Generated %d lesson ops, %d asset ops, %d attachment ops\n", len(groupOps), len(assetOps), len(attachmentOps))

	// FFprobe only assets that need it
	utils.Infof("CourseScan: Probing video assets for metadata and keyframes\n")
	assetMetadataByPath := probeVideos(s, assetOps)

	updatedCourse := course.CardPath != scanned.cardPath
	if updatedCourse {
		course.CardPath = scanned.cardPath
	}

	// Clean up extracted keyframes after processing and log completion
	defer func() {
		s.extractedKeyframes = make(map[string][]float64)
		utils.Infof("CourseScan: Completed scan for course ID: %s\n", scan.CourseID)
	}()

	utils.Infof("CourseScan: Applying database changes in transaction\n")
	return s.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		if updated, err := applyLessonCreateUpdateOps(txCtx, s, groupOps); err != nil {
			return err
		} else if updated {
			updatedCourse = true
		}

		if updated, err := applyAttachmentOps(txCtx, s, attachmentOps); err != nil {
			return err
		} else if updated {
			updatedCourse = true
		}

		if updated, err := applyAssetOps(txCtx, s, course, assetOps, assetMetadataByPath); err != nil {
			return err
		} else if updated {
			updatedCourse = true
		}

		if updated, err := applyLessonDeleteOps(txCtx, s, groupOps); err != nil {
			return err
		} else if updated {
			updatedCourse = true
		}

		if updatedCourse {
			course.InitialScan = true
			utils.Infof("CourseScan: Updating course metadata\n")
			return s.dao.UpdateCourse(txCtx, course)
		}

		utils.Infof("CourseScan: No course updates needed\n")
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

// lessonBucket accumulates files for a given module & prefix
type lessonBucket struct {
	groupedFiles []*parsedFile
	soloFiles    []*parsedFile
	attachFiles  []*parsedFile
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scannedResults holds all lessons and the card image path
type scannedResults struct {
	lessons  []*models.Lesson
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
	buckets := map[string]map[int]*lessonBucket{}
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
			buckets[module] = map[int]*lessonBucket{}
		}

		prefix := parsed.Prefix
		if buckets[module][prefix] == nil {
			buckets[module][prefix] = &lessonBucket{}
		}

		bucket := buckets[module][prefix]

		// Build the buckets of lessons of assets, attachments, and descriptions
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

	var lessons []*models.Lesson
	for module, prefixMap := range buckets {
		for prefix, bucket := range prefixMap {
			sort.Slice(bucket.groupedFiles, func(i, j int) bool {
				return *bucket.groupedFiles[i].SubPrefix < *bucket.groupedFiles[j].SubPrefix
			})

			lesson := &models.Lesson{
				CourseID: courseID,
				Module:   module,
				Prefix:   sql.NullInt16{Int16: int16(prefix), Valid: true},
			}

			// Set the lesson title from the first grouped file (if any)
			if len(bucket.groupedFiles) > 0 {
				lesson.Title = bucket.groupedFiles[0].Title
			}

			if len(bucket.groupedFiles) > 0 {
				for _, parsedFile := range bucket.groupedFiles {
					asset, err := parsedFile.toAsset(s.appFs.Fs, module, courseID)
					if err != nil {
						return nil, err
					}

					asset.LessonID = lesson.ID
					lesson.Assets = append(lesson.Assets, asset)
				}

				// Demote solo assets to attachments
				for _, parsedFile := range bucket.soloFiles {
					lesson.Attachments = append(lesson.Attachments, parsedFile.toAttachment())
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
					asset.LessonID = lesson.ID
					lesson.Assets = append(lesson.Assets, asset)

					// Set the lesson title from the selected solo file
					lesson.Title = pf.Title

					// Demote remaining solo files to attachments
					for i, other := range bucket.soloFiles {
						if i == idx {
							continue
						}
						lesson.Attachments = append(lesson.Attachments, other.toAttachment())
					}
				} else {
					asset, err := bucket.soloFiles[0].toAsset(s.appFs.Fs, module, courseID)
					if err != nil {
						return nil, err
					}
					asset.LessonID = lesson.ID
					lesson.Assets = append(lesson.Assets, asset)

					lesson.Title = bucket.soloFiles[0].Title
				}

				// Attachments
				for _, parsedFile := range bucket.attachFiles {
					lesson.Attachments = append(lesson.Attachments, parsedFile.toAttachment())
				}

			}

			// Create the lesson if it has at least 1 asset
			if len(lesson.Assets) > 0 {
				lessons = append(lessons, lesson)
			}
		}
	}

	return &scannedResults{
		lessons:  lessons,
		cardPath: cardPath,
	}, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// flatAttachmentsAndAssets returns a flat list of attachments and assets from a list of lessons
func flatAttachmentsAndAssets(lessons []*models.Lesson) ([]*models.Attachment, []*models.Asset) {
	var attachments []*models.Attachment
	var assets []*models.Asset
	for _, g := range lessons {
		assets = append(assets, g.Assets...)
		attachments = append(attachments, g.Attachments...)
	}

	return attachments, assets
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// probeVideos probes videos assets that match the operations create, replace, or swap
func probeVideos(s *CourseScan, ops []Op) map[string]*models.AssetMetadata {
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

	utils.Infof("CourseScan: Found %d video assets to probe\n", len(targets))

	assetMetadataByPath := map[string]*models.AssetMetadata{}
	mediaProbe := probe.MediaProbe{FFmpeg: s.ffmpeg}

	for _, asset := range targets {
		if info, err := mediaProbe.ProbeVideo(asset.Path); err == nil {
			assetMetadataByPath[asset.Path] = &models.AssetMetadata{
				VideoMetadata: &models.VideoMetadata{
					DurationSec: info.DurationSec,
					Container:   info.File.Container,
					MIMEType:    info.File.MIMEType,
					SizeBytes:   info.File.SizeBytes,
					OverallBPS:  info.File.OverallBPS,
					VideoCodec:  info.Video.Codec,
					Width:       info.Video.Width,
					Height:      info.Video.Height,
					FPSNum:      info.Video.FPSNum,
					FPSDen:      info.Video.FPSDen,
				},
				AudioMetadata: &models.AudioMetadata{
					Language:      info.Audio.Language,
					Codec:         info.Audio.Codec,
					Profile:       info.Audio.Profile,
					Channels:      info.Audio.Channels,
					ChannelLayout: info.Audio.ChannelLayout,
					SampleRate:    info.Audio.SampleRate,
					BitRate:       info.Audio.BitRate,
				},
			}

			// Extract keyframes for HLS transcoding
			if keyframes, err := mediaProbe.ExtractKeyframesForVideo(asset.Path); err == nil {
				// Store keyframes in a separate structure for later processing
				// We'll handle this in the asset processing phase
				s.extractedKeyframes[asset.Path] = keyframes
				utils.Infof("CourseScan: Extracted %d keyframes for video: %s\n", len(keyframes), asset.Path)
			} else {
				// Log keyframe extraction failure but don't fail the scan
				// Using utils.Errf for non-critical errors that shouldn't stop processing
				utils.Errf("Failed to extract keyframes for video: %s - %v\n", asset.Path, err)
			}
		} else {
			// Log video probe failure but don't fail the scan
			utils.Errf("Failed to probe video file: %s - %v\n", asset.Path, err)
		}
	}

	return assetMetadataByPath
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// applyLessonChanges applies lesson create/update operations. Delete is done after all
// asset and attachment operations have been applied
func applyLessonCreateUpdateOps(
	ctx context.Context,
	s *CourseScan,
	ops []Op,
) (bool, error) {
	if len(ops) == 0 {
		return false, nil
	}

	for _, op := range ops {
		switch v := op.(type) {
		case CreateLessonOp:
			// Create a new lesson
			if err := s.dao.CreateLesson(ctx, v.New); err != nil {
				return false, err
			}

			// fmt.Println("[Created Asset Group]", v.New.ID, v.New.Module, v.New.Prefix.Int16)

			for _, asset := range v.New.Assets {
				asset.LessonID = v.New.ID
			}

			for _, attachments := range v.New.Attachments {
				attachments.LessonID = v.New.ID
			}

		case NoLessonOp:
			// Ensure all new assets and attachments have the lesson ID. When a new
			// asset/attachment is added to an existing lesson, we need to ensure
			// that the lesson ID is set

			// fmt.Println("[No-Op Asset Group]", v.Existing.ID, v.Existing.Module, v.Existing.Prefix.Int16)

			for _, asset := range v.New.Assets {
				asset.LessonID = v.Existing.ID
			}

			for _, attachments := range v.New.Attachments {
				attachments.LessonID = v.Existing.ID
			}

		case UpdateLessonOp:
			// Update an existing lesson by giving the new lesson the ID of the existing
			// lesson, then calling update
			v.New.ID = v.Existing.ID

			if err := s.dao.UpdateLesson(ctx, v.New); err != nil {
				return false, err
			}

			// fmt.Println("[Updated Asset Group]", v.New.ID, v.New.Module, v.New.Prefix.Int16)

			for _, asset := range v.New.Assets {
				asset.LessonID = v.New.ID
			}

			for _, attachments := range v.New.Attachments {
				attachments.LessonID = v.New.ID
			}
		}
	}

	return true, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// applyLessonDeleteOps applies lesson deletion operations
func applyLessonDeleteOps(
	ctx context.Context,
	s *CourseScan,
	ops []Op,
) (bool, error) {
	if len(ops) == 0 {
		return false, nil
	}

	for _, op := range ops {
		switch v := op.(type) {

		case DeleteLessonOp:
			// Delete an existing lesson
			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.LESSON_TABLE_ID: v.Deleted.ID})
			if err := s.dao.DeleteLessons(ctx, dbOpts); err != nil {
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
	assetMetadataByPath map[string]*models.AssetMetadata,
) (bool, error) {
	if len(ops) == 0 {
		return false, nil
	}

	for _, op := range ops {
		switch v := op.(type) {
		case CreateAssetOp:
			// Create an asset that was found on disk and does not exist in the database

			// fmt.Println("[Create Asset]", v.New.Path, "->", v.New.Title, v.New.LessonID)

			if err := s.dao.CreateAsset(ctx, v.New); err != nil {
				return false, err
			}

			// If the asset is a video, also create video metadata
			if metadata := assetMetadataByPath[v.New.Path]; metadata != nil && v.New.Type.IsVideo() {
				metadata.AssetID = v.New.ID

				if err := s.dao.CreateAssetMetadata(ctx, metadata); err != nil {
					return false, err
				}

				// Store keyframes if they were extracted
				if keyframes, exists := s.extractedKeyframes[v.New.Path]; exists {
					// Validate keyframes before storage
					if len(keyframes) > 0 {
						assetKeyframes := &models.AssetKeyframes{
							AssetID:    v.New.ID,
							Keyframes:  keyframes,
							IsComplete: true,
						}

						if err := s.dao.CreateAssetKeyframes(ctx, assetKeyframes); err != nil {
							// Log error but don't fail the scan
							utils.Errf("Failed to store keyframes for asset %s (%s): %v\n", v.New.ID, v.New.Path, err)
						}
					}

					// Clean up the keyframes from memory after processing
					delete(s.extractedKeyframes, v.New.Path)
				}

				course.Duration += metadata.VideoMetadata.DurationSec
			}

		case UpdateAssetOp:
			// Update an existing asset by giving the new asset the ID of the existing asset and lesson,
			// then calling update
			//
			// This happens when the metadata (title, path, prefix, etc) changes but the
			// contents of the asset have not
			//
			// Asset progress will be preserved

			// fmt.Println("[Update Asset]", v.Existing.Path, "->", v.New.Title, v.New.LessonID)

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
			// creating the new asset. Take the existing lesson ID
			//
			// This happens when the contents of an existing asset have changed but the metadata
			// (title, path, prefix, etc) has not
			//
			// Asset progress will be lost

			// fmt.Println("[Replace Asset]", v.Existing.Path, v.Existing.LessonID, "->", v.New.Path, v.New.LessonID)

			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: v.Existing.ID})
			if err := s.dao.DeleteAssets(ctx, dbOpts); err != nil {
				return false, err
			}

			if v.Existing.AssetMetadata != nil {
				course.Duration -= v.Existing.AssetMetadata.VideoMetadata.DurationSec
			}

			v.New.LessonID = v.Existing.LessonID
			if err := s.dao.CreateAsset(ctx, v.New); err != nil {
				return false, err
			}

			// If the asset is a video, also create video metadata
			if metadata := assetMetadataByPath[v.New.Path]; metadata != nil && v.New.Type.IsVideo() {
				metadata.AssetID = v.New.ID

				if err := s.dao.CreateAssetMetadata(ctx, metadata); err != nil {
					return false, err
				}

				course.Duration += metadata.VideoMetadata.DurationSec
			}

		case OverwriteAssetOp:
			// Overwrite an existing but now deleted asset with another still existing asset by first
			// taking the ID of the existing asset, the lesson ID of the nonexisting asset, then
			// deleting the nonexisting asset, and finally calling update on the renamed asset to update
			// its metadata (title, path, prefix, etc)
			//
			// This happens when an asset has been renamed to that of another, now deleted, asset
			//
			// Asset progress will be preserved

			// fmt.Println("Overwrite Asset", v.Existing.Path, v.Existing.LessonID, "->", v.Renamed.Path, v.Deleted.LessonID)

			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: v.Deleted.ID})
			if err := s.dao.DeleteAssets(ctx, dbOpts); err != nil {
				return false, err
			}

			v.Renamed.ID = v.Existing.ID
			v.Renamed.LessonID = v.Deleted.LessonID

			if err := s.dao.UpdateAsset(ctx, v.Renamed); err != nil {
				return false, err
			}

		case SwapAssetOp:
			// Swap two assets by first deleting the existing assets, then recreating the assets.
			// The lesson IDs of the new assets will be swapped
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

				if existing.AssetMetadata != nil {
					course.Duration -= existing.AssetMetadata.VideoMetadata.DurationSec
				}
			}

			// Swap the new lesson IDs
			v.NewA.LessonID = v.ExistingB.LessonID
			v.NewB.LessonID = v.ExistingA.LessonID

			for _, newAsset := range []*models.Asset{v.NewA, v.NewB} {
				if err := s.dao.CreateAsset(ctx, newAsset); err != nil {
					return false, err
				}

				// If the asset is a video, also create video metadata
				if metadata := assetMetadataByPath[newAsset.Path]; metadata != nil && newAsset.Type.IsVideo() {
					metadata.AssetID = newAsset.ID

					if err := s.dao.CreateAssetMetadata(ctx, metadata); err != nil {
						return false, err
					}

					course.Duration += metadata.VideoMetadata.DurationSec
				}
			}

		case DeleteAssetOp:
			// Delete an asset that no longer exists on disk

			// fmt.Println("Delete Asset", v.Deleted.Path)

			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: v.Deleted.ID})
			if err := s.dao.DeleteAssets(ctx, dbOpts); err != nil {
				return false, err
			}

			if v.Deleted.AssetMetadata != nil {
				course.Duration -= v.Deleted.AssetMetadata.VideoMetadata.DurationSec
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
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
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

	// Non-nill when type is one of video, pdf, markdown, text
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
		Weight:    1,
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
