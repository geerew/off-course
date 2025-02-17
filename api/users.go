package api

import (
	"database/sql"
	"strings"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type userAPI struct {
	r *Router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initFsRoutes initializes the filesystem routes
func (r *Router) initUserRoutes() {
	userAPI := userAPI{r: r}

	userGroup := r.api.Group("/users")

	userGroup.Get("", protectedRoute, userAPI.getUsers)
	userGroup.Post("", protectedRoute, userAPI.createUser)
	userGroup.Put("/:id", protectedRoute, userAPI.updateUser)
	userGroup.Delete("/:id", protectedRoute, userAPI.deleteUser)

	// TODO Add route to revoke all sessions for a user
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api userAPI) getUsers(c *fiber.Ctx) error {
	orderBy := c.Query("orderBy", models.USER_TABLE+"."+models.BASE_CREATED_AT+" desc")

	options := &database.Options{
		OrderBy:    strings.Split(orderBy, ","),
		Pagination: pagination.NewFromApi(c),
	}

	users := []*models.User{}
	err := api.r.dao.List(c.UserContext(), &users, options)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up users", err)
	}

	pResult, err := options.Pagination.BuildResult(userResponseHelper(users))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api userAPI) createUser(c *fiber.Ctx) error {
	userReq := &userRequest{}

	if err := c.BodyParser(userReq); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if userReq.Username == "" || userReq.Password == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A username and password are required", nil)
	}

	// Default the role to a user when not provided
	if userReq.Role == "" {
		userReq.Role = types.UserRoleUser.String()
	}

	user := &models.User{
		Username:     userReq.Username,
		DisplayName:  userReq.Username,
		PasswordHash: auth.GeneratePassword(userReq.Password),
		Role:         types.NewUserRole(userReq.Role),
	}

	if userReq.DisplayName != "" {
		user.DisplayName = userReq.DisplayName
	}

	err := api.r.dao.CreateUser(c.UserContext(), user)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Username already exists", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating user", err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Revokes all sessions for the user when the role is updated
func (api userAPI) updateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	userReq := &userRequest{}
	if err := c.BodyParser(userReq); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	if userReq.DisplayName == "" && userReq.Password == "" && userReq.Role == "" {
		return errorResponse(c, fiber.StatusBadRequest, "No data to update", nil)
	}

	user := &models.User{Base: models.Base{ID: id}}
	err := api.r.dao.GetById(c.UserContext(), user)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "User not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up user", err)
	}

	if userReq.DisplayName != "" {
		user.DisplayName = userReq.DisplayName
	}

	if userReq.Password != "" {
		user.PasswordHash = auth.GeneratePassword(userReq.Password)
	}

	if userReq.Role != "" {
		// Do nothing when the role is the same
		if user.Role.String() == userReq.Role {
			userReq.Role = ""
		} else {
			user.Role = types.NewUserRole(userReq.Role)
		}
	}

	err = api.r.dao.UpdateUser(c.UserContext(), user)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error updating user", err)
	}

	// Revoke all sessions for the give id
	if userReq.Role != "" {
		err := api.r.sessionManager.DeleteUserSessions(id)
		if err != nil {
			return errorResponse(c, fiber.StatusInternalServerError, "Error deleting user sessions", err)
		}
	}

	return c.Status(fiber.StatusOK).JSON(&userResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api userAPI) deleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	user := &models.User{Base: models.Base{ID: id}}
	err := api.r.dao.Delete(c.UserContext(), user, nil)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting user", err)
	}

	err = api.r.sessionManager.DeleteUserSessions(id)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting user sessions", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
