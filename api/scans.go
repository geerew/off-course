package api

import (
	"bufio"
	"context"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scansAPI struct {
	r *Router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initScanRoutes initializes the scan routes
func (r *Router) initScanRoutes() {
	scansAPI := scansAPI{
		r: r,
	}

	g := r.apiGroup("scans")

	g.Get("/", protectedRoute, scansAPI.getScans)
	g.Get("/stream", protectedRoute, scansAPI.streamScans)
	g.Get("/:courseId", protectedRoute, scansAPI.getScan)
	g.Post("", protectedRoute, scansAPI.createScan)
	g.Delete("/:id", protectedRoute, scansAPI.deleteScan)
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

	scans, err := api.r.appDao.ListScans(ctx, dbOpts)
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

	dbOpts := dao.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_COURSE_ID: courseId})
	scan, err := api.r.appDao.GetScan(ctx, dbOpts)
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

	scan, err := api.r.app.CourseScan.Add(ctx, req.CourseID)
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

	dbOpts := dao.NewOptions().WithWhere(squirrel.Eq{models.SCAN_TABLE_ID: id})
	if err := api.r.appDao.DeleteScans(ctx, dbOpts); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting scan", err)
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// streamScans streams scan updates using Server-Sent Events (SSE)
func (api *scansAPI) streamScans(c *fiber.Ctx) error {
	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Track last seen scan data to detect changes
	isAdmin := principal.Role == types.UserRoleAdmin
	streamCtx, cancel := context.WithCancel(ctx)
	lastSeen := make(map[string]*scanResponse)

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer cancel()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		writeEvent := func(payload interface{}) error {
			eventBytes, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			if _, err = w.WriteString("data: " + string(eventBytes) + "\n\n"); err != nil {
				return err
			}
			return w.Flush()
		}

		writeComment := func(text string) error {
			if _, err := w.WriteString(":" + text + "\n\n"); err != nil {
				return err
			}
			return w.Flush()
		}

		// Initial comment to confirm connection
		if err := writeComment(" connected"); err != nil {
			return
		}

		for {
			select {
			case <-streamCtx.Done():
				return
			case <-ticker.C:
				dbOpts := dao.NewOptions()
				scans, err := api.r.appDao.ListScans(streamCtx, dbOpts)
				if err != nil {
					_ = writeEvent(map[string]interface{}{
						"type":    "error",
						"message": "Failed to fetch scans",
					})
					continue
				}

				updated := false
				current := make(map[string]struct{})
				scanResponses := scanResponseHelper(scans, isAdmin)

				for _, scanResp := range scanResponses {
					current[scanResp.ID] = struct{}{}
					last, exists := lastSeen[scanResp.ID]
					if !exists || last == nil || last.Status != scanResp.Status || last.Message != scanResp.Message || last.UpdatedAt != scanResp.UpdatedAt {
						if err := writeEvent(map[string]interface{}{
							"type": "scan_update",
							"data": scanResp,
						}); err != nil {
							return
						}
						lastSeen[scanResp.ID] = scanResp
						updated = true
					}
				}

				for id := range lastSeen {
					if _, exists := current[id]; !exists {
						if err := writeEvent(map[string]interface{}{
							"type": "scan_deleted",
							"data": map[string]string{"id": id},
						}); err != nil {
							return
						}
						delete(lastSeen, id)
						updated = true
					}
				}

				// Send heartbeat to keep connection alive
				if !updated {
					if err := writeComment(" keep-alive"); err != nil {
						return
					}
				}
			}
		}
	}))

	return nil
}
