package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media"
	"github.com/geerew/off-course/utils/media/hls"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	// Cache FFmpeg instance for course scanning
	cachedFFmpeg *media.FFmpeg
	ffmpegOnce   sync.Once
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getCachedFFmpeg returns a cached FFmpeg instance for course scanning
func getCachedFFmpeg(t *testing.T) *media.FFmpeg {
	ffmpegOnce.Do(func() {
		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			// Skip if FFmpeg unavailable
			t.Skip("FFmpeg not available for testing")
		}
		cachedFFmpeg = ffmpeg
	})
	return cachedFFmpeg
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setupAdmin creates a test router with an admin user
func setupAdmin(t *testing.T) (*Router, context.Context) {
	return setup(t, "admin", types.UserRoleAdmin)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setupUser creates a test router with a regular user
func setupUser(t *testing.T) (*Router, context.Context) {
	return setup(t, "user", types.UserRoleUser)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setupNoAuth creates a test router without authentication
//
// Note: role doesn't matter when no auth
func setupNoAuth(t *testing.T) (*Router, context.Context) {
	return setup(t, "", types.UserRoleUser)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setup creates a test router
func setup(t *testing.T, id string, role types.UserRole) (*Router, context.Context) {
	t.Helper()

	// Create a test logger
	testLogger := logger.New(&logger.Config{
		Level:         logger.LevelInfo,
		ConsoleOutput: false, // Disable console output for tests
	})

	appFs := appfs.New(afero.NewMemMapFs())

	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: "./oc_data",
		AppFs:   appFs,
		Testing: true,
	})

	require.NoError(t, err)
	require.NotNil(t, dbManager)

	// Get FFmpeg instance
	ffmpeg := getCachedFFmpeg(t)

	// Initialize HLS settings
	transcoder, err := hls.NewTranscoder(&hls.TranscoderConfig{
		CachePath: "./oc_data",
		AppFs:     appFs,
		Logger:    testLogger.WithHLS(),
		Dao:       dao.New(dbManager.DataDb),
	})
	require.NoError(t, err)
	require.NotNil(t, transcoder)

	courseScan := coursescan.New(&coursescan.CourseScanConfig{
		Db:     dbManager.DataDb,
		AppFs:  appFs,
		Logger: testLogger.WithCourseScan(),
		FFmpeg: ffmpeg,
	})

	// Router
	config := &RouterConfig{
		DbManager:     dbManager,
		AppFs:         appFs,
		CourseScan:    courseScan,
		Logger:        testLogger.WithAPI(),
		SignupEnabled: true,
	}

	// Configure middleware based on whether we have auth
	if id != "" {
		// Use dev auth for normal tests
		config.Middleware = []MiddlewareFactory{
			func(r *Router) fiber.Handler { return devAuthMiddleware(id, role) },
		}
	} else {
		// Use CORS-only for recovery tests
		config.Middleware = []MiddlewareFactory{
			func(r *Router) fiber.Handler { return corsMiddleWare() },
		}
	}

	router := NewRouter(config)

	// Create user only if we have auth
	if id != "" {
		user := models.User{
			Base: models.Base{
				ID: id,
			},
			Username:     id,
			Role:         role,
			PasswordHash: "password",
			DisplayName:  "Test User",
		}
		require.NoError(t, router.dao.CreateUser(context.Background(), &user))
	}

	// Initialize bootstrap
	router.InitBootstrap()

	ctx := context.Background()

	// Add principal to context only if we have auth
	if id != "" {
		principal := types.Principal{
			UserID: id,
			Role:   role,
		}
		ctx = context.WithValue(ctx, types.PrincipalContextKey, principal)
	}

	return router, ctx
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// requestHelper sends a request to the router and returns the status code and body
func requestHelper(t *testing.T, router *Router, req *http.Request) (int, []byte, error) {
	t.Helper()

	resp, err := router.App.Test(req)
	if err != nil {
		return -1, nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// unmarshalHelper unmarshals a body into a pagination result and a slice of items
func unmarshalHelper[T any](t *testing.T, body []byte) (pagination.PaginationResult, []T) {
	t.Helper()

	var respData pagination.PaginationResult
	err := json.Unmarshal(body, &respData)
	require.NoError(t, err)

	var resp []T
	for _, item := range respData.Items {
		var r T
		require.Nil(t, json.Unmarshal(item, &r))
		resp = append(resp, r)
	}

	return respData, resp
}
