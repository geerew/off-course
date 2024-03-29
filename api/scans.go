package api

import (
	"database/sql"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scans struct {
	appFs         *appFs.AppFs
	scanDao       *daos.ScanDao
	courseScanner *jobs.CourseScanner
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scanResponse struct {
	ID        string           `json:"id"`
	CourseID  string           `json:"courseId"`
	Status    types.ScanStatus `json:"status"`
	CreatedAt types.DateTime   `json:"createdAt"`
	UpdatedAt types.DateTime   `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindScansApi(router fiber.Router, appFs *appFs.AppFs, db database.Database, courseScanner *jobs.CourseScanner) {
	api := scans{
		appFs:         appFs,
		scanDao:       daos.NewScanDao(db),
		courseScanner: courseScanner,
	}

	subGroup := router.Group("/scans")

	subGroup.Get("/:courseId", api.getScan)
	subGroup.Post("", api.createScan)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scans) getScan(c *fiber.Ctx) error {
	courseId := c.Params("courseId")

	scan, err := api.scanDao.Get(courseId)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up scan")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up scan - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(scanResponseHelper([]*models.Scan{scan})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scans) createScan(c *fiber.Ctx) error {
	scan := &models.Scan{}

	if err := c.BodyParser(scan); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "error parsing data - " + err.Error(),
		})
	}

	if scan.CourseID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "a course ID is required",
		})
	}

	scan, err := api.courseScanner.Add(scan.CourseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid course ID",
			})
		}

		log.Err(err).Msg("error creating scan job")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error creating scan job - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(scanResponseHelper([]*models.Scan{scan})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func scanResponseHelper(scans []*models.Scan) []*scanResponse {
	responses := []*scanResponse{}
	for _, scan := range scans {
		responses = append(responses, &scanResponse{
			ID:        scan.ID,
			CourseID:  scan.CourseID,
			Status:    scan.Status,
			CreatedAt: scan.CreatedAt,
			UpdatedAt: scan.UpdatedAt,
		})
	}

	return responses
}
