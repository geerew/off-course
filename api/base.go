package api

import (
	"context"
	"log/slog"
	"net"
	"time"

	"sync/atomic"

	"github.com/Masterminds/squirrel"
	"github.com/fatih/color"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/session"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	fibersession "github.com/gofiber/fiber/v2/middleware/session"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Router defines a router
type Router struct {
	App            *fiber.App
	api            fiber.Router
	config         *RouterConfig
	dao            *dao.DAO
	logDao         *dao.DAO
	bootstrapped   int32
	sessionManager *session.SessionManager
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RouterConfig defines the configuration for the router
type RouterConfig struct {
	DbManager     *database.DatabaseManager
	Logger        *slog.Logger
	AppFs         *appfs.AppFs
	CourseScan    *coursescan.CourseScan
	HttpAddr      string
	IsProduction  bool
	SignupEnabled bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new router
func NewRouter(config *RouterConfig) *Router {
	r := &Router{
		config: config,
		dao:    dao.New(config.DbManager.DataDb),
		logDao: dao.New(config.DbManager.LogsDb),
	}

	r.createSessionStore()

	r.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	r.initMiddleware()
	r.initRoutes()

	return r
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// devRouter creates a new router for use in development. The main difference is the lack of
// middleware and the use of a predefined user id and role
func devRouter(config *RouterConfig, id string, role types.UserRole) *Router {
	r := &Router{
		config: config,
		dao:    dao.New(config.DbManager.DataDb),
		logDao: dao.New(config.DbManager.LogsDb),
	}

	r.createSessionStore()

	r.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	r.App.Use(devAuthMiddleware(id, role))
	r.initRoutes()

	return r
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Serve serves the API and UI
func (r *Router) Serve() error {
	r.initBootstrap()

	ln, err := net.Listen("tcp", r.config.HttpAddr)
	if err != nil {
		return err
	}

	utils.Infof(
		"%s %s",
		"Server started at", color.CyanString("http://%s\n", r.config.HttpAddr),
	)

	return r.App.Listener(ln)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Private
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initMiddleware initializes the middleware
func (r *Router) initMiddleware() {
	// Middleware
	r.App.Use(loggerMiddleware(r.config))
	r.App.Use(corsMiddleWare())
	r.App.Use(bootstrapMiddleware(r))
	r.App.Use(authMiddleware(r))

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initRoutes initializes the routes
func (r *Router) initRoutes() {
	// UI
	r.bindUi()

	// API
	r.api = r.App.Group("/api")
	r.initAuthRoutes()
	r.initFsRoutes()
	r.initCourseRoutes()
	r.initScanRoutes()
	r.initTagRoutes()
	r.initUserRoutes()
	r.initLogRoutes()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// createSessionStore creates the session store
func (r *Router) createSessionStore() {
	config := fibersession.Config{
		KeyLookup:      "cookie:session",
		Expiration:     7 * (24 * time.Hour),
		CookieHTTPOnly: true,
	}

	sqliteStorage := session.NewSqliteStorage(r.config.DbManager.DataDb, 10*time.Second)

	r.sessionManager = session.New(r.config.DbManager.DataDb, config, sqliteStorage)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initBootstrap determines if the application is bootstrapped by checking if there is
// an admin user
func (r *Router) initBootstrap() {
	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_ROLE: types.UserRoleAdmin})
	count, err := r.dao.CountUsers(context.Background(), dbOpts)
	if err != nil {
		utils.Errf("Failed to count users: %s\n", err.Error())
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

// isBootstrapped checks if the application is bootstrapped
func (r *Router) isBootstrapped() bool {
	return atomic.LoadInt32(&r.bootstrapped) == 1
}
