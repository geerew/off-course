package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// createTestRouter creates a test router
func createTestRouter(t *testing.T) *Router {
	return createTestRouterWithDataDir(t, "./oc_data")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// createTestRouterWithDataDir creates a router with a specific data directory
func createTestRouterWithDataDir(t *testing.T, dataDir string) *Router {
	t.Helper()

	// Create a test logger
	testLogger := logger.New(&logger.Config{
		Level:         logger.LevelInfo,
		ConsoleOutput: false, // Disable console output for tests
	})

	appFs := appfs.New(afero.NewMemMapFs())

	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: dataDir,
		AppFs:   appFs,
		Testing: true,
	})
	require.NoError(t, err)
	require.NotNil(t, dbManager)

	// Get FFmpeg instance
	ffmpeg := getCachedFFmpeg(t)

	courseScan := coursescan.New(&coursescan.CourseScanConfig{
		Db:     dbManager.DataDb,
		AppFs:  appFs,
		Logger: testLogger.WithCourseScan(),
		FFmpeg: ffmpeg,
	})

	// Router config
	config := &RouterConfig{
		DbManager:     dbManager,
		AppFs:         appFs,
		CourseScan:    courseScan,
		Logger:        testLogger.WithAPI(),
		SignupEnabled: true,
		DataDir:       dataDir,
	}

	// Create router
	router := &Router{
		config: config,
		dao:    dao.New(config.DbManager.DataDb),
		logDao: dao.New(config.DbManager.LogsDb),
		logger: config.Logger,
	}

	router.createSessionStore()

	router.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Only add CORS middleware, no authentication
	router.App.Use(corsMiddleWare())
	router.initRoutes()

	return router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestRecovery_ResetPassword(t *testing.T) {
	t.Run("200 (success)", func(t *testing.T) {
		ctx := context.Background()
		tempDir := t.TempDir()

		// Create router with temp directory
		router := createTestRouterWithDataDir(t, tempDir)

		// Create admin user
		user := &models.User{
			Username:     "testadmin",
			DisplayName:  "Test Admin",
			PasswordHash: security.RandomString(10),
			Role:         types.UserRoleAdmin,
		}
		require.NoError(t, router.dao.CreateUser(ctx, user))

		// Generate recovery token in the same directory
		recoveryToken, err := auth.GenerateRecoveryToken("testadmin", "newpassword123", tempDir)
		require.NoError(t, err)

		// Make request
		reqBody := map[string]string{
			"token": recoveryToken.Token,
		}
		jsonData, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/recovery", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		// Verify response
		var response map[string]interface{}
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)
		require.Equal(t, "Password reset successfully", response["message"])
		require.Equal(t, "testadmin", response["username"])

		// Verify password was updated
		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_USERNAME: "testadmin"})
		updatedUser, err := router.dao.GetUser(ctx, dbOpts)
		require.NoError(t, err)
		require.True(t, auth.ComparePassword(updatedUser.PasswordHash, "newpassword123"))
	})

	t.Run("401 (invalid token)", func(t *testing.T) {
		tempDir := t.TempDir()
		router := createTestRouterWithDataDir(t, tempDir)

		// Make request with invalid token
		reqBody := map[string]string{
			"token": "invalid-token",
		}
		jsonData, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/recovery", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("401 (no token file)", func(t *testing.T) {
		tempDir := t.TempDir()
		router := createTestRouterWithDataDir(t, tempDir)

		// Make request with token but no file
		reqBody := map[string]string{
			"token": "some-token",
		}
		jsonData, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/recovery", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("400 (missing token)", func(t *testing.T) {
		router := createTestRouter(t)

		// Make request without token
		req := httptest.NewRequest(http.MethodPost, "/api/admin/recovery", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("404 (user not found)", func(t *testing.T) {
		tempDir := t.TempDir()
		router := createTestRouterWithDataDir(t, tempDir)

		// Generate recovery token for non-existent user
		recoveryToken, err := auth.GenerateRecoveryToken("nonexistent", "password", tempDir)
		require.NoError(t, err)

		// Make request
		reqBody := map[string]string{
			"token": recoveryToken.Token,
		}
		jsonData, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/recovery", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("403 (user not admin)", func(t *testing.T) {
		ctx := context.Background()
		tempDir := t.TempDir()
		router := createTestRouterWithDataDir(t, tempDir)

		// Create regular user (not admin)
		user := &models.User{
			Username:     "testuser",
			DisplayName:  "Test User",
			PasswordHash: security.RandomString(10),
			Role:         types.UserRoleUser,
		}
		require.NoError(t, router.dao.CreateUser(ctx, user))

		// Generate recovery token
		recoveryToken, err := auth.GenerateRecoveryToken("testuser", "newpassword123", tempDir)
		require.NoError(t, err)

		// Make request
		reqBody := map[string]string{
			"token": recoveryToken.Token,
		}
		jsonData, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/recovery", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, status)
	})
}
