package api

import (
	"bufio"
	"context"
	"encoding/json"
	"time"

	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/coursescan"
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
	principal, _, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	// Get all scans from CMap (ephemeral, no pagination needed)
	scanStates := api.r.app.CourseScan.GetAllScans()
	responses := scanResponseHelper(scanStates, principal.Role == types.UserRoleAdmin)

	return c.Status(fiber.StatusOK).JSON(responses)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *scansAPI) getScan(c *fiber.Ctx) error {
	courseId := c.Params("courseId")

	principal, _, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	scanState := api.r.app.CourseScan.GetScanByCourseID(courseId)
	if scanState == nil {
		return errorResponse(c, fiber.StatusNotFound, "Scan not found", nil)
	}

	return c.Status(fiber.StatusOK).JSON(scanResponseHelper([]*coursescan.ScanState{scanState}, principal.Role == types.UserRoleAdmin)[0])
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

	scanState, err := api.r.app.CourseScan.Add(ctx, req.CourseID)
	if err != nil {
		if err == utils.ErrCourseNotFound {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid course ID", nil)
		}
		return errorResponse(c, fiber.StatusInternalServerError, "Error creating scan job", err)
	}

	return c.Status(fiber.StatusCreated).JSON(scanResponseHelper([]*coursescan.ScanState{scanState}, principal.Role == types.UserRoleAdmin)[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api scansAPI) deleteScan(c *fiber.Ctx) error {
	id := c.Params("id")

	_, _, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	// Cancel and remove scan from CMap
	if !api.r.app.CourseScan.CancelAndRemoveScan(id) {
		return errorResponse(c, fiber.StatusNotFound, "Scan not found", nil)
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

		ticker := time.NewTicker(1 * time.Second)
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
				// Get all scans from CMap
				scanStates := api.r.app.CourseScan.GetAllScans()
				scanResponses := scanResponseHelper(scanStates, isAdmin)

				updated := false
				current := make(map[string]struct{})

				for _, scanResp := range scanResponses {
					current[scanResp.ID] = struct{}{}
					last, exists := lastSeen[scanResp.ID]
					if !exists || last == nil || last.Status != scanResp.Status || last.Message != scanResp.Message {
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
