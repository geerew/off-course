package api

import (
	"context"
	"fmt"
	"net"
	"time"

	"sync/atomic"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media/hls"
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
	App            *fiber.App
	api            fiber.Router
	config         *RouterConfig
	dao            *dao.DAO
	logDao         *dao.DAO
	bootstrapped   int32
	sessionManager *session.SessionManager
	logger         *logger.Logger
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RouterConfig defines the configuration for the router
type RouterConfig struct {
	DbManager     *database.DatabaseManager
	Logger        *logger.Logger
	AppFs         *appfs.AppFs
	CourseScan    *coursescan.CourseScan
	Transcoder    *hls.Transcoder
	HttpAddr      string
	IsProduction  bool
	SignupEnabled bool
	DataDir       string

	// Optional custom middleware stack; if empty, defaults are applied
	Middleware []MiddlewareFactory
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new router
func NewRouter(config *RouterConfig) *Router {
	r := &Router{
		config: config,
		dao:    dao.New(config.DbManager.DataDb),
		logDao: dao.New(config.DbManager.LogsDb),
		logger: config.Logger,
	}

	r.createSessionStore()

	r.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	if len(config.Middleware) == 0 {
		r.initMiddleware()
	} else {
		for _, f := range config.Middleware {
			r.App.Use(f(r))
		}
	}

	r.initRoutes()

	return r
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Serve serves the API and UI
func (r *Router) Serve() error {
	r.InitBootstrap()

	ln, err := net.Listen("tcp", r.config.HttpAddr)
	if err != nil {
		return err
	}

	r.logger.Info().Str("url", fmt.Sprintf("http://%s", r.config.HttpAddr)).Msg("Server started")

	return r.App.Listener(ln)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Private
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initMiddleware initializes the middleware
func (r *Router) initMiddleware() {
	// Middleware
	r.App.Use(requestLoggingMiddleware(r.logger))
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
	}

	sqliteStorage := session.NewSqliteStorage(r.config.DbManager.DataDb, 10*time.Second)

	r.sessionManager = session.New(r.config.DbManager.DataDb, config, sqliteStorage)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InitBootstrap determines if the application is bootstrapped by checking if there is
// an admin user
func (r *Router) InitBootstrap() {
	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_ROLE: types.UserRoleAdmin})
	count, err := r.dao.CountUsers(context.Background(), dbOpts)
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
