package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"sync/atomic"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/app"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/session"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	fibersession "github.com/gofiber/fiber/v2/middleware/session"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MiddlewareFactory defines a function that creates middleware with access to the router
type MiddlewareFactory func(r *Router) fiber.Handler

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Router defines a router
type Router struct {
	fiberApp       *fiber.App
	app            *app.App
	appDao         *dao.DAO
	logDao         *dao.DAO
	bootstrapped   int32
	sessionManager *session.SessionManager
	logger         *logger.Logger
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewRouter creates a new router from an App instance
func NewRouter(app *app.App) *Router {
	r := &Router{
		app:    app,
		appDao: dao.New(app.DbManager.DataDb),
		logDao: dao.New(app.DbManager.LogsDb),
		logger: app.Logger.WithAPI(),
	}

	r.createSessionStore()

	r.fiberApp = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	r.initMiddleware()
	r.initRoutes()

	return r
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Serve serves the API and UI
func (r *Router) Serve() error {
	ln, err := net.Listen("tcp", r.app.Config.HttpAddr)
	if err != nil {
		return err
	}

	r.logger.Info().Str("url", fmt.Sprintf("http://%s", r.app.Config.HttpAddr)).Msg("Server started")

	return r.fiberApp.Listener(ln)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Private
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initMiddleware initializes the middleware
func (r *Router) initMiddleware() {
	// Middleware
	r.fiberApp.Use(requestLoggingMiddleware(r.logger))
	r.fiberApp.Use(corsMiddleWare())
	r.fiberApp.Use(bootstrapMiddleware(r))
	r.fiberApp.Use(authMiddleware(r))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initRoutes initializes the routes
func (r *Router) initRoutes() {
	// UI
	r.bindUi()

	// API routes
	r.initAuthRoutes()
	r.initFsRoutes()
	r.initCourseRoutes()
	r.initScanRoutes()
	r.initTagRoutes()
	r.initUserRoutes()
	r.initLogRoutes()
	r.initRecoveryRoutes()
	r.initHlsRoutes()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// createSessionStore creates the session store
func (r *Router) createSessionStore() {
	config := fibersession.Config{
		KeyLookup:      "cookie:session",
		Expiration:     7 * (24 * time.Hour),
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	}

	sqliteStorage := session.NewSqliteStorage(r.app.DbManager.DataDb, 10*time.Second)

	r.sessionManager = session.New(r.app.DbManager.DataDb, config, sqliteStorage)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InitBootstrap determines if the app is bootstrapped by checking if there is
// an admin user
func (r *Router) InitBootstrap() {
	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_ROLE: types.UserRoleAdmin})
	count, err := r.appDao.CountUsers(context.Background(), dbOpts)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count users")
	}

	if count != 0 {
		atomic.StoreInt32(&r.bootstrapped, 1)
	} else {
		atomic.StoreInt32(&r.bootstrapped, 0)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setBootstrapped sets the application as bootstrapped
func (r *Router) setBootstrapped() {
	atomic.StoreInt32(&r.bootstrapped, 1)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsBootstrapped checks if the application is bootstrapped
func (r *Router) IsBootstrapped() bool {
	return atomic.LoadInt32(&r.bootstrapped) == 1
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// apiGroup returns the API router group
func (r *Router) apiGroup(groupPath string) fiber.Router {
	return r.fiberApp.Group("/api/" + groupPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetTestMiddleware replaces the middleware stack with test middleware.
// This should only be used in tests.
func (r *Router) SetTestMiddleware(factories ...MiddlewareFactory) {
	// Clear existing middleware by creating a new Fiber app with same config
	r.fiberApp = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Default error handler that returns JSON
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"message": err.Error(),
			})
		},
	})

	// Apply test middleware
	for _, factory := range factories {
		r.fiberApp.Use(factory(r))
	}

	// Re-initialize routes (they depend on the Fiber app)
	r.initRoutes()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Test is a test helper that wraps FiberApp.Test() for testing purposes
func (r *Router) Test(req *http.Request, msTimeout ...int) (*http.Response, error) {
	return r.fiberApp.Test(req, msTimeout...)
}
