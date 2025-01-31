package api

import (
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
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api userAPI) getUsers(c *fiber.Ctx) error {
	orderBy := c.Query("orderBy", models.USER_TABLE+".created_at desc")

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
