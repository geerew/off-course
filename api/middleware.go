package api

// TODO When 2 windows are open with bootstrap, and the first one bootstraps, the second should
//   1) error on submit or
//   2) redirect to the login page on refresh

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerMiddleware logs the request and response details
func loggerMiddleware(config *RouterConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		started := time.Now()
		err := c.Next()

		attrs := make([]any, 0, 15)

		attrs = append(attrs, slog.Any("type", types.LogTypeRequest))

		if !started.IsZero() {
			attrs = append(attrs, slog.Float64("execTime", float64(time.Since(started))/float64(time.Millisecond)))
		}

		if err != nil {
			attrs = append(
				attrs,
				slog.String("error", err.Error()),
			)
		}

		method := strings.ToUpper(c.Method())

		attrs = append(
			attrs,
			slog.Int("status", c.Response().StatusCode()),
			slog.String("method", method),
		)

		// Get the port
		host := string(c.Request().Host())
		if strings.Contains(host, ":") {
			attrs = append(attrs, slog.String("port", strings.Split(host, ":")[1]))
		}

		// Determine if the response is an error
		isErrorResponse := err != nil

		var jsonBody map[string]interface{}
		if jsonErr := json.Unmarshal(c.Response().Body(), &jsonBody); jsonErr == nil {
			if val, exists := jsonBody["message"]; exists {
				attrs = append(attrs, slog.String("message", val.(string)))
			}

			if !isErrorResponse {
				if val, exists := jsonBody["error"]; exists {
					isErrorResponse = true
					attrs = append(
						attrs, slog.String("error", val.(string)))
				}
			}
		}

		message := method + " " + c.OriginalURL()

		if isErrorResponse {
			config.Logger.Error(message, attrs...)
		} else {
			config.Logger.Info(message, attrs...)
		}

		return err
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// corsMiddleWare creates a CORS middleware
func corsMiddleWare() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, DELETE, HEAD, PATCH",
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// bootstrapMiddleware checks if the app is bootstrapped. If not, it redirects
// to /auth/bootstrap
//
// Bootstrapping is the process of setting up the app for the first time. It involves
// the creation of 1 admin user, which the /auth/bootstrap endpoint handles
func bootstrapMiddleware(r *Router) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If not bootstrapped, for everything through /auth/bootstrap or
		// /api/auth/bootstrap
		if !r.isBootstrapped() {

			path := c.Path()

			if r.isDevUIPath(path) || r.isProdUIPath(path) || r.isFavicon(path) {
				return c.Next()
			}

			// API check
			if strings.HasPrefix(path, "/api/") {
				if strings.HasPrefix(path, "/api/auth/bootstrap") {
					c.Locals("bootstrapping", true)
					return c.Next()
				} else {
					return c.SendStatus(fiber.StatusForbidden)
				}
			}

			// UI check
			if strings.HasPrefix(path, "/auth/bootstrap") {
				c.Locals("bootstrapping", true)
				return c.Next()
			} else {
				return c.Redirect("/auth/bootstrap/")
			}
		}

		return c.Next()
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// authMiddleware authenticates the request
func authMiddleware(r *Router) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if bootstrapping, _ := c.Locals("bootstrapping").(bool); bootstrapping {
			return c.Next()
		}

		path := c.Path()
		isAPI := strings.HasPrefix(path, "/api/")
		isAuthUI := strings.HasPrefix(path, "/auth/")
		isMe := strings.HasPrefix(path, "/api/auth/me")
		isLogout := strings.HasPrefix(path, "/api/auth/logout")

		if r.isDevUIPath(path) || r.isProdUIPath(path) || r.isFavicon(path) || isLogout {
			return c.Next()
		}

		session, err := r.sessionManager.Get(c)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if session.Fresh() {

			if isAPI {
				if strings.HasPrefix(path, "/api/auth/login") || (r.config.SignupEnabled && strings.HasPrefix(path, "/api/auth/register")) ||
					strings.HasPrefix(path, "/api/auth/signup-status") {
					return c.Next()
				}
				return c.SendStatus(fiber.StatusForbidden)
			}

			if isAuthUI {
				if !r.config.SignupEnabled && strings.HasPrefix(path, "/auth/register") {
					return c.Redirect("/auth/login")
				}

				return c.Next()
			}

			return c.Redirect("/auth/login")
		}

		if isAuthUI {
			return c.Redirect("/")
		}

		if strings.HasPrefix(path, "/api/auth/") && !isMe {
			return c.SendStatus(fiber.StatusOK)
		}

		userID, ok1 := session.Get("id").(string)
		userRole, ok2 := session.Get("role").(string)
		if !ok1 || !ok2 || userID == "" || userRole == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		if userRole != "admin" && r.isProtectedUIPage(path) {
			return c.Redirect("/")
		}

		role := types.UserRole(userRole)
		principal := types.Principal{
			UserID: userID,
			Role:   role,
		}

		// for your UI/router logic:
		c.Locals(types.PrincipalContextKey, principal)

		return c.Next()
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// devAuthMiddleware sets the id, role (for use in development only)
func devAuthMiddleware(id string, role types.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		principal := types.Principal{
			UserID: id,
			Role:   role,
		}

		// for your UI/router logic:
		c.Locals(types.PrincipalContextKey, principal)
		return c.Next()
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isProdUIPath checks if the request is for a sveltekit asset when running in production mode
func (r *Router) isProdUIPath(path string) bool {
	if r.config.IsProduction && strings.HasPrefix(path, "/_app/") {
		return true
	}

	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isDevUIPath checks if the request is for a sveltekit path when running in dev mode
func (r *Router) isDevUIPath(path string) bool {
	if !r.config.IsProduction &&
		(strings.HasPrefix(path, "/node_modules/") ||
			strings.HasPrefix(path, "/.svelte-kit/") ||
			strings.HasPrefix(path, "/src/") ||
			strings.HasPrefix(path, "/@")) {
		return true
	}

	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isProtectedUIPage checks if the request is intended for a protected UI page
func (r *Router) isProtectedUIPage(path string) bool {
	return strings.HasPrefix(path, "/admin")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isFavicon checks if the request is for a favicon
func (r *Router) isFavicon(path string) bool {
	return strings.HasPrefix(path, "/favicon.")
}
