package api

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type recoveryAPI struct {
	dao     *dao.DAO
	logger  *logger.Logger
	dataDir string
	appFs   *appfs.AppFs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// recoveryRequest represents the request body for recovery endpoint
type recoveryRequest struct {
	Token string `json:"token" validate:"required"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initRecoveryRoutes initializes the recovery routes
func (r *Router) initRecoveryRoutes() {
	recoveryAPI := recoveryAPI{
		dao:     r.dao,
		logger:  r.config.Logger,
		dataDir: r.config.DataDir,
		appFs:   r.config.AppFs,
	}

	// Recovery endpoint - no authentication required (validates via token file)
	r.api.Post("/admin/recovery", recoveryAPI.resetPassword)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api recoveryAPI) resetPassword(c *fiber.Ctx) error {
	// Parse request
	req := &recoveryRequest{}
	if err := c.BodyParser(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing request", err)
	}

	if req.Token == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Token is required", nil)
	}

	// Validate recovery token
	recoveryToken, err := auth.ValidateRecoveryToken(api.appFs, req.Token, api.dataDir)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid or expired recovery token", nil)
	}

	// Get user by username
	ctx := context.Background()
	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_USERNAME: recoveryToken.Username})
	user, err := api.dao.GetUser(ctx, dbOpts)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up user", err)
	}

	if user == nil {
		return errorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// Verify user is admin
	if user.Role != types.UserRoleAdmin {
		return errorResponse(c, fiber.StatusForbidden, "User is not an admin", nil)
	}

	// Update password
	user.PasswordHash = recoveryToken.PasswordHash
	err = api.dao.UpdateUser(ctx, user)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error updating password", err)
	}

	// Delete the recovery token file (best-effort)
	_ = auth.DeleteRecoveryToken(api.appFs, api.dataDir)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Password reset successfully",
		"username": recoveryToken.Username,
	})
}
