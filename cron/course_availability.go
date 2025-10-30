package cron

import (
	"context"
	"os"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
)

type courseAvailability struct {
	db        database.Database
	dao       *dao.DAO
	appFs     *appfs.AppFs
	logger    *logger.Logger
	batchSize int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (ca *courseAvailability) run() error {
	perPage := 100
	page := 1
	totalPages := 1

	// Stats
	totalScanned := 0
	madeAvailable := 0
	madeUnavailable := 0
	updatedCount := 0

	ca.logger.Info().
		Int("per_page", perPage).
		Int("batch_size", ca.batchSize).
		Msg("cron: updating course availability started")

	coursesBatch := make([]*models.Course, 0, ca.batchSize)

	// Create an admin principal context for the cron job
	principal := types.Principal{
		UserID: "availability-cron",
		Role:   types.UserRoleAdmin,
	}
	ctx := context.WithValue(context.Background(), types.PrincipalContextKey, principal)

	for page <= totalPages {
		p := pagination.New(page, perPage)
		dbOpts := database.NewOptions().WithPagination(p)

		// Fetch a batch of courses
		courses, err := ca.dao.ListCourses(ctx, dbOpts)
		if err != nil {
			ca.logger.Error().Err(err).Int("page", page).Msg("cron: failed to fetch courses")
			return err
		}
		totalScanned += len(courses)

		// Update total pages after the first fetch
		if page == 1 {
			totalPages = p.TotalPages()
		}

		// Process each course in the batch
		for _, course := range courses {
			if _, err := ca.appFs.Fs.Stat(course.Path); err != nil {
				if os.IsNotExist(err) {
					if course.Available {
						// The course is currently marked as available but is now unavailable
						course.Available = false
						madeUnavailable++
						coursesBatch = append(coursesBatch, course)
					}
				} else {
					// Failed to check the availability of the course
					ca.logger.Error().
						Err(err).
						Str("course", course.Title).
						Str("path", course.Path).
						Msg("cron: failed to stat course")
					return err
				}
			} else if !course.Available {
				// The course is currently marked as unavailable but is now available
				course.Available = true
				madeAvailable++
				coursesBatch = append(coursesBatch, course)
			}

			// Update the courses if we hit the batch size
			if len(coursesBatch) == ca.batchSize {
				if err := ca.writeAll(ctx, coursesBatch); err != nil {
					ca.logger.Error().Err(err).Int("batch_len", len(coursesBatch)).Msg("cron: failed to write availability batch")
					return err
				}
				updatedCount += len(coursesBatch)
				ca.logger.Debug().Int("updated", len(coursesBatch)).Msg("cron: availability batch written")
				coursesBatch = coursesBatch[:0]
			}
		}

		page++
	}

	// Update any remaining courses
	if len(coursesBatch) > 0 {
		if err := ca.writeAll(ctx, coursesBatch); err != nil {
			ca.logger.Error().Err(err).Int("batch_len", len(coursesBatch)).Msg("cron: failed to write final availability batch")
			return err
		}
		updatedCount += len(coursesBatch)
		ca.logger.Debug().Int("updated", len(coursesBatch)).Msg("cron: final availability batch written")
	}

	ca.logger.Info().
		Int("scanned", totalScanned).
		Int("updated", updatedCount).
		Int("made_available", madeAvailable).
		Int("made_unavailable", madeUnavailable).
		Msg("cron: updating course availability completed")

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (ca *courseAvailability) writeAll(ctx context.Context, courses []*models.Course) error {
	// Update the courses in a transaction
	err := ca.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		for _, course := range courses {
			if err := ca.dao.UpdateCourse(txCtx, course); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}
