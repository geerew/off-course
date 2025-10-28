package coursescan

import (
	"context"

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
type CourseScanProcessorFn func(context.Context, *CourseScan, *models.Scan) error

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScan scans a course and finds assets and attachments
type CourseScan struct {
	appFs              *appfs.AppFs
	db                 database.Database
	dao                *dao.DAO
	logger             *logger.Logger
	ffmpeg             *media.FFmpeg
	jobSignal          chan struct{}
	extractedKeyframes map[string][]float64 // path -> keyframes
}

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
		appFs:              config.AppFs,
		db:                 config.Db,
		dao:                dao.New(config.Db),
		logger:             config.Logger,
		ffmpeg:             config.FFmpeg,
		jobSignal:          make(chan struct{}, 1),
		extractedKeyframes: make(map[string][]float64),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Add inserts a course scan job into the db
func (s *CourseScan) Add(ctx context.Context, courseId string) (*models.Scan, error) {
	// Look up the course
	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.COURSE_TABLE_ID: courseId})
	course, err := s.dao.GetCourse(ctx, dbOpts)
	if err != nil {
		return nil, err
	}

	if course == nil {
		return nil, utils.ErrCourseNotFound
	}

	// Get the scan from the db and return that
	dbOpts = database.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_COURSE_ID: courseId})
	scan, err := s.dao.GetScan(ctx, dbOpts)
	if err != nil {
		return nil, err
	}

	if scan == nil {
		// No scan job exists, create a new one
		scan = &models.Scan{CourseID: course.ID}
		if err := s.dao.CreateScan(ctx, scan); err != nil {
			return nil, err
		}
	} else {
		// Scan job already exists
		s.logger.Debug().
			Str("job", scan.ID).
			Str("path", course.Path).
			Msg("Scan job already exists")

		return scan, nil
	}

	// Signal the worker to process the job
	select {
	case s.jobSignal <- struct{}{}:
	default:
	}

	s.logger.Info().
		Str("path", course.Path).
		Msg("Added scan job")

	return scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Worker processes jobs out of the DB sequentially
//
// TODO find a way to stop processing when the job is deleted
func (s *CourseScan) Worker(ctx context.Context, processorFn CourseScanProcessorFn, processingDone chan bool) {
	s.logger.Debug().Msg("Started course scanner worker")

	// Create an admin principal context for the course scan worker
	principal := types.Principal{
		UserID: "course-scan-worker",
		Role:   types.UserRoleAdmin,
	}
	ctx = context.WithValue(ctx, types.PrincipalContextKey, principal)

	for {
		<-s.jobSignal

		// Keep process jobs from the scans table until there are no more jobs
		for {
			nextScan, err := s.dao.NextWaitingScan(ctx)
			if err != nil {
				s.logger.Error().
					Err(err).
					Msg("Failed to look up the next scan job")

				break
			}

			// Nothing more to process
			if nextScan == nil {
				s.logger.Debug().Msg("Finished processing all scan jobs")
				break
			}

			s.logger.Info().
				Str("job", nextScan.ID).
				Str("path", nextScan.CoursePath).
				Msg("Processing scan job")

			err = processorFn(ctx, s, nextScan)
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("path", nextScan.CoursePath).
					Msg("Failed to process scan job")
			}

			// Cleanup
			dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_ID: nextScan.ID})
			if err := s.dao.DeleteScans(ctx, dbOpts); err != nil {
				s.logger.Error().
					Err(err).
					Str("job", nextScan.ID).
					Msg("Failed to delete scan job")

				break
			}
		}

		// Signal that processing is done
		if processingDone != nil {
			processingDone <- true
		}

		// Clear any pending signal that were sent while processing
		select {
		case <-s.jobSignal:
		default:
		}
	}
}
