package api

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scansAPI struct {
	logger     *slog.Logger
	appFs      *appfs.AppFs
	dao        *dao.DAO
	courseScan *coursescan.CourseScan
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initScanRoutes initializes the scan routes
func (r *Router) initScanRoutes() {
	scansAPI := scansAPI{
		logger:     r.config.Logger,
		appFs:      r.config.AppFs,
		courseScan: r.config.CourseScan,
		dao:        r.dao,
	}

	scanGroup := r.api.Group("/scans")
	scanGroup.Get("/", protectedRoute, scansAPI.getScans)
	scanGroup.Get("/:courseId", protectedRoute, scansAPI.getScan)
	scanGroup.Post("", protectedRoute, scansAPI.createScan)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scansAPI) getScans(c *fiber.Ctx) error {
	scans := []*models.Scan{}
	err := api.dao.List(c.UserContext(), &scans, nil)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up scan", err)
	}

	return c.Status(fiber.StatusOK).JSON(scanResponseHelper(scans))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scansAPI) getScan(c *fiber.Ctx) error {
	courseId := c.Params("courseId")

	scan := &models.Scan{}
	options := &database.Options{
		Where: squirrel.Eq{fmt.Sprintf("%s.%s", models.SCAN_TABLE, models.SCAN_COURSE_ID): courseId},
	}

	err := api.dao.Get(c.UserContext(), scan, options)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Scan not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up scan", err)
	}

	return c.Status(fiber.StatusOK).JSON(scanResponseHelper([]*models.Scan{scan})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scansAPI) createScan(c *fiber.Ctx) error {
	req := &ScanRequest{}
	if err := c.BodyParser(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if req.CourseID == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A course ID is required", nil)
	}

	scan, err := api.courseScan.Add(c.UserContext(), req.CourseID)
	if err != nil {
		if err == utils.ErrInvalidId {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid course ID", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating scan job", err)
	}

	return c.Status(fiber.StatusCreated).JSON(scanResponseHelper([]*models.Scan{scan})[0])
}
