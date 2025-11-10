package coursescan

import (
	"context"
	"sort"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanProcessorFn is a function that processes a course scan job
type CourseScanProcessorFn func(context.Context, *CourseScan, *ScanState) error

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScan scans a course and finds assets and attachments
type CourseScan struct {
	appFs  *appfs.AppFs
	db     database.Database
	dao    *dao.DAO
	logger *logger.Logger
	ffmpeg *media.FFmpeg

	// In-memory scan state storage
	scans utils.CMap[string, *ScanState]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	// scanPollInterval is how often the worker polls for waiting scans
	scanPollInterval = 1 * time.Second
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanConfig is the config for a CourseScan
type CourseScanConfig struct {
	Db     database.Database
	AppFs  *appfs.AppFs
	Logger *logger.Logger
	FFmpeg *media.FFmpeg
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new CourseScan
func New(config *CourseScanConfig) *CourseScan {
	return &CourseScan{
		appFs:  config.AppFs,
		db:     config.Db,
		dao:    dao.New(config.Db),
		logger: config.Logger,
		ffmpeg: config.FFmpeg,
		scans:  utils.NewCMap[string, *ScanState](),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Add creates a course scan job and adds it to the CMap
func (s *CourseScan) Add(ctx context.Context, courseId string) (*ScanState, error) {
	// Look up the course to get path and title
	dbOpts := dao.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: courseId})
	course, err := s.dao.GetCourse(ctx, dbOpts)
	if err != nil {
		return nil, err
	}

	if course == nil {
		return nil, utils.ErrCourseNotFound
	}

	// Check if a scan already exists for this course
	existingScan := s.GetScanByCourseID(courseId)
	if existingScan != nil {
		// Scan job already exists
		s.logger.Debug().
			Str("course_id", courseId).
			Str("course_path", course.Path).
			Str("scan_id", existingScan.ID).
			Msg("Scan job already exists")

		return existingScan, nil
	}

	// Create a new scan state
	scanState := NewScanState(courseId, course.Path, course.Title)

	// Add to CMap
	s.scans.Set(scanState.ID, scanState)

	s.logger.Info().
		Str("course_id", courseId).
		Str("course_path", course.Path).
		Str("scan_id", scanState.ID).
		Msg("Added scan job")

	return scanState, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScanByCourseID finds a scan by course ID
func (s *CourseScan) GetScanByCourseID(courseID string) *ScanState {
	var found *ScanState
	s.scans.Range(func(scanID string, scanState *ScanState) bool {
		if scanState.CourseID == courseID {
			found = scanState
			return false // Stop iteration
		}
		return true // Continue iteration
	})
	return found
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScan retrieves a scan by its ID
func (s *CourseScan) GetScan(scanID string) (*ScanState, bool) {
	return s.scans.Get(scanID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAllScans returns all scans in the CMap, sorted by status (processing first) then by createdAt (oldest first)
// This ensures deterministic ordering for the frontend
func (s *CourseScan) GetAllScans() []*ScanState {
	var allScans []*ScanState
	s.scans.Range(func(scanID string, scanState *ScanState) bool {
		allScans = append(allScans, scanState)
		return true // Continue iteration
	})

	sortScans(allScans)

	return allScans
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CancelScan cancels a scan by its ID if it exists
func (s *CourseScan) CancelScan(scanID string) bool {
	scanState, exists := s.scans.Get(scanID)
	if !exists {
		return false
	}
	scanState.Cancel()
	return true
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RemoveScan removes a scan from the CMap by its ID
func (s *CourseScan) RemoveScan(scanID string) {
	s.scans.Remove(scanID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CancelAndRemoveScan cancels a scan (if running) and removes it from the CMap
func (s *CourseScan) CancelAndRemoveScan(scanID string) bool {
	scanState, exists := s.scans.Get(scanID)
	if !exists {
		return false
	}
	// Cancel the scan if it's running
	scanState.Cancel()
	// Remove from CMap
	s.scans.Remove(scanID)
	return true
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CancelAndRemoveScansByCourseID cancels and removes all scans for a given course ID
// This is used when a course is deleted to ensure any ongoing scans are stopped
func (s *CourseScan) CancelAndRemoveScansByCourseID(courseID string) {
	var scanIDsToRemove []string
	s.scans.Range(func(scanID string, scanState *ScanState) bool {
		if scanState.CourseID == courseID {
			// Cancel the scan if it's running
			scanState.Cancel()
			scanIDsToRemove = append(scanIDsToRemove, scanID)
		}
		return true // Continue iteration
	})
	// Remove all found scans
	for _, scanID := range scanIDsToRemove {
		s.scans.Remove(scanID)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Worker polls the CMap for waiting scans and processes them sequentially
func (s *CourseScan) Worker(ctx context.Context, processorFn CourseScanProcessorFn) {
	s.logger.Debug().Msg("Started course scanner worker")

	ticker := time.NewTicker(scanPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug().Msg("Course scanner worker stopped")
			return
		case <-ticker.C:
			// Poll for waiting scans
			var waitingScans []*ScanState
			s.scans.Range(func(scanID string, scanState *ScanState) bool {
				if scanState.GetStatus() == types.ScanStatusWaiting {
					waitingScans = append(waitingScans, scanState)
				}

				return true
			})

			// Sort scans to ensure deterministic processing order
			// Processing scans first, then by createdAt (oldest first)
			// Since we're filtering to waiting scans, they'll all be waiting, but this ensures consistent ordering
			sortScans(waitingScans)

			// Process each waiting scan
			for _, scanState := range waitingScans {
				// Check if scan was cancelled or removed before processing
				if scanState.IsCancelled() {
					s.logger.Debug().
						Str("course_id", scanState.CourseID).
						Str("scan_id", scanState.ID).
						Msg("Skipping cancelled scan")
					// Remove from CMap if it still exists
					s.scans.Remove(scanState.ID)
					continue
				}

				// Double-check scan still exists in CMap (could have been deleted)
				if _, exists := s.scans.Get(scanState.ID); !exists {
					s.logger.Debug().
						Str("course_id", scanState.CourseID).
						Str("scan_id", scanState.ID).
						Msg("Skipping scan that was removed from CMap")
					continue
				}

				// Create a cancellable context for this scan
				scanCtx, cancel := context.WithCancel(ctx)
				scanState.SetCancel(cancel)

				// Check cancellation one more time after setting cancel function
				if scanState.IsCancelled() {
					cancel() // Cancel the context we just created
					s.logger.Debug().
						Str("course_id", scanState.CourseID).
						Str("scan_id", scanState.ID).
						Msg("Skipping scan cancelled after context creation")
					s.scans.Remove(scanState.ID)
					continue
				}

				s.logger.Info().
					Str("course_id", scanState.CourseID).
					Str("course_path", scanState.CoursePath).
					Str("scan_id", scanState.ID).
					Msg("Processing scan job")

				err := processorFn(scanCtx, s, scanState)
				if err != nil {
					s.logger.Error().
						Err(err).
						Str("course_id", scanState.CourseID).
						Str("course_path", scanState.CoursePath).
						Str("scan_id", scanState.ID).
						Msg("Failed to process scan job")
				}

				// Cleanup: remove scan from CMap
				s.scans.Remove(scanState.ID)
			}
		}
	}
}

// sortScans sorts scans by status (processing first) then by createdAt (oldest first), then by ID
// This ensures deterministic ordering for both Worker and GetAllScans
// If multiple scans have the same status and createdAt, they're sorted by ID (lexicographic) as a tiebreaker
func sortScans(scans []*ScanState) {
	sort.Slice(scans, func(i, j int) bool {
		iStatus := scans[i].GetStatus()
		jStatus := scans[j].GetStatus()

		// Processing scans come first
		iProcessing := iStatus == types.ScanStatusProcessing
		jProcessing := jStatus == types.ScanStatusProcessing

		if iProcessing && !jProcessing {
			return true // i comes first
		}
		if !iProcessing && jProcessing {
			return false // j comes first
		}

		// Same status - sort by createdAt (oldest first)
		if scans[i].CreatedAt.Before(scans[j].CreatedAt) {
			return true
		}
		if scans[j].CreatedAt.Before(scans[i].CreatedAt) {
			return false
		}

		// Same status and createdAt - use ID as tiebreaker for deterministic ordering
		return scans[i].ID < scans[j].ID
	})
}
