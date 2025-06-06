package api

import (
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/session"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// TODO Add unit tests for the auth routes

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type authAPI struct {
	dao            *dao.DAO
	sessionManager *session.SessionManager
	r              *Router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initFsRoutes initializes the filesystem routes
func (r *Router) initAuthRoutes() {
	authAPI := authAPI{
		dao:            r.dao,
		sessionManager: r.sessionManager,
		r:              r,
	}

	authGroup := r.api.Group("/auth")

	authGroup.Get("/signup-status", authAPI.signupStatus)
	authGroup.Post("/bootstrap", authAPI.bootstrap)
	authGroup.Post("/register", authAPI.register)
	authGroup.Post("/login", authAPI.login)
	authGroup.Post("/logout", authAPI.logout)

	authGroup.Get("/me", authAPI.getMe)
	authGroup.Put("/me", authAPI.updateMe)
	authGroup.Delete("/me", authAPI.deleteMe)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) signupStatus(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(signupStatusResponse{
		Enabled: api.r.config.SignupEnabled,
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) register(c *fiber.Ctx) error {
	if !api.r.config.SignupEnabled {
		return errorResponse(c, fiber.StatusForbidden, "Sign-up is disabled", nil)
	}

	userReq := &userRequest{}

	if err := c.BodyParser(userReq); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if userReq.Username == "" || userReq.Password == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Username and/or password cannot be empty", nil)
	}

	user := &models.User{
		Username:     userReq.Username,
		DisplayName:  userReq.Username, // Set the display name to the username by default
		PasswordHash: auth.GeneratePassword(userReq.Password),
	}

	// The first user will always be an admin
	bootstrapAdmin, ok := c.Locals("bootstrapAdmin").(bool)
	if ok && bootstrapAdmin {
		user.Role = types.UserRoleAdmin
	} else {
		user.Role = types.UserRoleUser
	}

	err := api.dao.CreateUser(c.UserContext(), user)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Username already exists", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating user", err)
	}

	err = api.sessionManager.SetSession(c, user.ID, user.Role)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error setting session", err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) bootstrap(c *fiber.Ctx) error {
	c.Locals("bootstrapAdmin", true)
	err := api.register(c)

	if err == nil {
		api.r.setBootstrapped()
	}

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) login(c *fiber.Ctx) error {
	userReq := &userRequest{}

	if err := c.BodyParser(userReq); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if userReq.Username == "" || userReq.Password == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Username and/or password cannot be empty", nil)
	}

	user := &models.User{}
	err := api.dao.GetUser(c.UserContext(), user, &database.Options{Where: squirrel.Eq{models.USER_TABLE_USERNAME: userReq.Username}})
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid username and/or password", nil)
	}

	if !auth.ComparePassword(user.PasswordHash, userReq.Password) {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid username and/or password", nil)
	}

	err = api.sessionManager.SetSession(c, user.ID, user.Role)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error setting session", err)
	}

	return c.SendStatus(fiber.StatusOK)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) logout(c *fiber.Ctx) error {
	err := api.sessionManager.DeleteSession(c)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting session", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) getMe(c *fiber.Ctx) error {
	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	user := &models.User{Base: models.Base{ID: principal.UserID}}
	if err := api.dao.GetUser(ctx, user, nil); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting user information", err)
	}

	return c.Status(fiber.StatusOK).JSON(&userResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) updateMe(c *fiber.Ctx) error {
	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	user := &models.User{Base: models.Base{ID: principal.UserID}}
	if err := api.dao.GetUser(ctx, user, nil); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting user information", err)
	}

	userReq := &userRequest{}
	if err := c.BodyParser(userReq); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if userReq.DisplayName == "" && userReq.Password == "" {
		return errorResponse(c, fiber.StatusBadRequest, "No data to update", nil)
	}

	if userReq.DisplayName != "" {
		user.DisplayName = userReq.DisplayName
	}

	if userReq.Password != "" {
		if !auth.ComparePassword(user.PasswordHash, userReq.CurrentPassword) {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid current password", nil)
		}

		user.PasswordHash = auth.GeneratePassword(userReq.Password)
	}

	err = api.dao.UpdateUser(ctx, user)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error updating user", err)
	}

	return c.Status(fiber.StatusOK).JSON(&userResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api authAPI) deleteMe(c *fiber.Ctx) error {
	principal, ctx, err := principalCtx(c)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Missing principal", nil)
	}

	user := &models.User{Base: models.Base{ID: principal.UserID}}
	if err := api.dao.GetUser(ctx, user, nil); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting user information", err)
	}

	userReq := &userRequest{}
	if err := c.BodyParser(userReq); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if !auth.ComparePassword(user.PasswordHash, userReq.CurrentPassword) {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid password", nil)
	}

	if user.Role == types.UserRoleAdmin {
		// Count the number of admin users and fail if there is only one
		adminCount, err := dao.Count(ctx, api.dao, &models.User{}, &database.Options{
			Where: squirrel.Eq{models.USER_TABLE_ROLE: types.UserRoleAdmin},
		})

		if err != nil {
			return errorResponse(c, fiber.StatusInternalServerError, "Error counting admin users", err)
		}

		if adminCount == 1 {
			return errorResponse(c, fiber.StatusBadRequest, "Unable to delete the last admin user", nil)
		}
	}

	err = dao.Delete(ctx, api.dao, user, nil)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting user", err)
	}

	err = api.sessionManager.DeleteUserSessions(user.ID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting user sessions", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
