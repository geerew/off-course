package api

import (
	"github.com/geerew/off-course/cron"
	"github.com/geerew/off-course/version"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type versionAPI struct {
	r *Router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initVersionRoutes initializes the version routes
func (r *Router) initVersionRoutes() {
	versionAPI := versionAPI{
		r: r,
	}

	g := r.apiGroup("version")

	g.Get("", versionAPI.getVersion)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getVersion returns the application version
func (api *versionAPI) getVersion(c *fiber.Ctx) error {
	currentVersion := version.GetVersion()
	response := fiber.Map{
		"version": currentVersion,
	}

	// Add latest release if available and different from current version
	if cron.ReleaseChecker != nil {
		latestRelease := cron.ReleaseChecker.GetLatestRelease()
		if latestRelease != "" && latestRelease != currentVersion {
			response["latestRelease"] = latestRelease
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
