package api

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scansAPI struct {
	logger     *logger.Logger
	appFs      *appfs.AppFs
	dao        *dao.DAO
	courseScan *coursescan.CourseScan
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initScanRoutes initializes the scan routes
func (r *Router) initScanRoutes() {
	scansAPI := scansAPI{
		logger:     r.logger.WithAPI(),
		appFs:      r.config.AppFs,
		courseScan: r.config.CourseScan,
		dao:        r.dao,
	}

	scanGroup := r.api.Group("/scans")
	scanGroup.Get("/", protectedRoute, scansAPI.getScans)
	scanGroup.Get("/:courseId", protectedRoute, scansAPI.getScan)
	scanGroup.Post("", protectedRoute, scansAPI.createScan)
	scanGroup.Delete("/:id", protectedRoute, scansAPI.deleteScan)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scansAPI) getScans(c *fiber.Ctx) error {
	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	builderOpts := builderOptions{
		DefaultOrderBy: defaultScansOrderBy,
		Paginate:       true,
	}

	dbOpts, err := optionsBuilder(c, builderOpts, principal.UserID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing query", err)
	}

	scans, err := api.dao.ListScans(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up scan", err)
	}

	pResult, err := dbOpts.Pagination.BuildResult(scanResponseHelper(scans, principal.Role == types.UserRoleAdmin))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scansAPI) getScan(c *fiber.Ctx) error {
	courseId := c.Params("courseId")

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_COURSE_ID: courseId})
	scan, err := api.dao.GetScan(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up scan", err)
	}

	if scan == nil {
		return errorResponse(c, fiber.StatusNotFound, "Scan not found", nil)
	}

	return c.Status(fiber.StatusOK).JSON(scanResponseHelper([]*models.Scan{scan}, principal.Role == types.UserRoleAdmin)[0])
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

	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	scan, err := api.courseScan.Add(ctx, req.CourseID)
	if err != nil {
		if err == utils.ErrCourseNotFound {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid course ID", nil)
		}
		return errorResponse(c, fiber.StatusInternalServerError, "Error creating scan job", err)
	}

	return c.Status(fiber.StatusCreated).JSON(scanResponseHelper([]*models.Scan{scan}, principal.Role == types.UserRoleAdmin)[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api scansAPI) deleteScan(c *fiber.Ctx) error {
	id := c.Params("id")

	_, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_ID: id})
	if err := api.dao.DeleteScans(ctx, dbOpts); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting scan", err)
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}
