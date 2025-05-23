package coursescan

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/types"
)

var (
	loggerType = slog.Any("type", types.LogTypeCourseScan)
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanProcessorFn is a function that processes a course scan job
type CourseScanProcessorFn func(context.Context, *CourseScan, *models.Scan) error

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScan scans a course and finds assets and attachments
type CourseScan struct {
	appFs     *appfs.AppFs
	db        database.Database
	dao       *dao.DAO
	logger    *slog.Logger
	jobSignal chan struct{}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseScanConfig is the config for a CourseScan
type CourseScanConfig struct {
	Db     database.Database
	AppFs  *appfs.AppFs
	Logger *slog.Logger
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new CourseScan
func New(config *CourseScanConfig) *CourseScan {
	return &CourseScan{
		appFs:     config.AppFs,
		db:        config.Db,
		dao:       dao.New(config.Db),
		logger:    config.Logger,
		jobSignal: make(chan struct{}, 1),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Add inserts a course scan job into the db
func (s *CourseScan) Add(ctx context.Context, courseId string) (*models.Scan, error) {
	// Look up the course
	course := &models.Course{}
	options := &database.Options{
		Where:            squirrel.Eq{models.COURSE_TABLE_ID: courseId},
		ExcludeRelations: []string{models.COURSE_RELATION_PROGRESS},
	}
	if err := s.dao.GetCourse(ctx, course, options); err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrInvalidId
		}

		return nil, err
	}

	// Get the scan from the db and return that
	scan := &models.Scan{}
	if err := s.dao.GetScan(ctx, scan, &database.Options{Where: squirrel.Eq{models.SCAN_TABLE_COURSE_ID: courseId}}); err != nil {
		if err == sql.ErrNoRows {
			// No scan job exists, create a new one
			scan = &models.Scan{CourseID: course.ID}
			if err := s.dao.CreateScan(ctx, scan); err != nil {
				return nil, err
			}
		} else {
			// Error
			return nil, err
		}
	} else {
		// Scan job already exists
		s.logger.Debug(
			"Scan job already exists",
			loggerType,
			slog.String("job", scan.ID),
			slog.String("path", course.Path),
		)

		return scan, nil
	}

	// Signal the worker to process the job
	select {
	case s.jobSignal <- struct{}{}:
	default:
	}

	s.logger.Info(
		"Added scan job",
		loggerType,
		slog.String("path", course.Path),
	)

	return scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Worker processes jobs out of the DB sequentially
func (s *CourseScan) Worker(ctx context.Context, processorFn CourseScanProcessorFn, processingDone chan bool) {
	s.logger.Debug("Started course scanner worker", loggerType)

	for {
		<-s.jobSignal

		// Keep process jobs from the scans table until there are no more jobs
		for {
			nextScan := &models.Scan{}
			err := s.dao.NextWaitingScan(ctx, nextScan)
			if err != nil {
				// Nothing more to process
				if err == sql.ErrNoRows {
					s.logger.Debug("Finished processing all scan jobs", loggerType)
					break
				}

				// Error
				s.logger.Error(
					"Failed to look up the next scan job",
					loggerType,
					slog.String("error", err.Error()),
				)

				break
			}

			s.logger.Info(
				"Processing scan job",
				loggerType,
				slog.String("job", nextScan.ID),
				slog.String("path", nextScan.CoursePath),
			)

			err = processorFn(ctx, s, nextScan)
			if err != nil {
				s.logger.Error(
					"Failed to process scan job",
					loggerType,
					slog.String("error", err.Error()),
					slog.String("path", nextScan.CoursePath),
				)
			}

			// Cleanup
			if err := dao.Delete(ctx, s.dao, nextScan, nil); err != nil {
				s.logger.Error(
					"Failed to delete scan job",
					loggerType,
					slog.String("error", err.Error()),
					slog.String("job", nextScan.ID),
				)

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
