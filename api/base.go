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
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/coursescan"
	storage "github.com/geerew/off-course/utils/session_storage"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Router defines a router
type Router struct {
	App          *fiber.App
	api          fiber.Router
	config       *RouterConfig
	dao          *dao.DAO
	logDao       *dao.DAO
	bootstrapped int32
	sessionStore *session.Store
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RouterConfig defines the configuration for the router
type RouterConfig struct {
	DbManager    *database.DatabaseManager
	Logger       *slog.Logger
	AppFs        *appFs.AppFs
	CourseScan   *coursescan.CourseScan
	HttpAddr     string
	IsProduction bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new router
func NewRouter(config *RouterConfig) *Router {
	r := &Router{
		config: config,
		dao:    dao.NewDAO(config.DbManager.DataDb),
		logDao: dao.NewDAO(config.DbManager.LogsDb),
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
		dao:    dao.NewDAO(config.DbManager.DataDb),
		logDao: dao.NewDAO(config.DbManager.LogsDb),
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
	sessionStorage := storage.NewSqlite(r.config.DbManager.DataDb.DB(), "sessions")

	r.sessionStore = session.New(session.Config{
		Storage:        sessionStorage,
		KeyLookup:      "cookie:session",
		Expiration:     7 * (24 * time.Hour),
		CookieHTTPOnly: true,
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initBootstrap determines if the application is bootstrapped by checking if there is
// an admin user
func (r *Router) initBootstrap() {
	options := &database.Options{
		Where: squirrel.Eq{models.USER_TABLE + "." + models.USER_ROLE: types.UserRoleAdmin},
	}
	count, err := r.dao.Count(context.Background(), &models.User{}, options)
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
