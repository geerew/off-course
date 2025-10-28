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
	ca.logger.Debug().Str("type", "cron").Msg("Updating course availability")
	perPage := 100
	page := 1
	totalPages := 1

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
			ca.logger.Error().Err(err).Msg("Failed to fetch courses")
			return err
		}

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
						coursesBatch = append(coursesBatch, course)
					}
				} else {
					// Failed to check the availability of the course
					ca.logger.Error().
						Err(err).
						Str("course", course.Title).
						Str("path", course.Path).
						Msg("Failed to stat course")

					return err
				}
			} else if !course.Available {
				// The course is currently marked as unavailable but is now available
				course.Available = true
				coursesBatch = append(coursesBatch, course)
			}

			// Update the courses if we hit the batch size
			if len(coursesBatch) == ca.batchSize {
				ca.writeAll(ctx, coursesBatch)
				coursesBatch = coursesBatch[:0]
			}
		}

		page++
	}

	// Update any remaining courses
	if len(coursesBatch) > 0 {
		ca.writeAll(ctx, coursesBatch)
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (ca *courseAvailability) writeAll(ctx context.Context, courses []*models.Course) {
	// Update the courses in a transaction
	err := ca.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		for _, course := range courses {
			err := ca.dao.UpdateCourse(txCtx, course)
			if err != nil {
				return err
			}

		}

		return nil
	})

	if err != nil {
		ca.logger.Error().Err(err).Msg("Failed to update course availability")
	}
}
