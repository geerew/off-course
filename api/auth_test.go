package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAuth_Register(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		router.setBootstrapped()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_USERNAME: "test"})
		record, err := router.appDao.GetUser(ctx, dbOpts)
		require.NoError(t, err)
		require.NotEqual(t, "password", record.PasswordHash)
		require.Equal(t, types.UserRoleUser, record.Role)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		// Missing both
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Missing password
		req = httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Missing username
		req = httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"password": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Both empty
		req = httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "", "password": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")
	})

	t.Run("400 (existing user)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		// Same case
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username already exists")

		// Different case
		req = httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username": "TEST", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username already exists")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAuth_Bootstrap(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		// Clear the admin user to make it unbootstrapped
		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_USERNAME: "admin"})
		err := router.appDao.DeleteUsers(ctx, dbOpts)
		require.NoError(t, err)
		router.InitBootstrap()

		// Generate a bootstrap token using the app's data directory and filesystem
		bootstrapToken, err := auth.GenerateBootstrapToken(router.app.Config.DataDir, router.app.AppFs.Fs)
		require.NoError(t, err)

		// Create user with token
		req := httptest.NewRequest(http.MethodPost, "/api/auth/bootstrap/"+bootstrapToken.Token, strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err2 := requestHelper(t, router, req)
		require.NoError(t, err2)
		require.Equal(t, http.StatusCreated, status)

		dbOpts2 := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_USERNAME: "test"})
		record, err3 := router.appDao.GetUser(ctx, dbOpts2)
		require.NoError(t, err3)
		require.NotEqual(t, "password", record.PasswordHash)
		require.Equal(t, types.UserRoleAdmin, record.Role)
		require.True(t, router.IsBootstrapped())
	})

	t.Run("401 (invalid token)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		// Clear the admin user to make it unbootstrapped
		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_USERNAME: "admin"})
		err := router.appDao.DeleteUsers(context.Background(), dbOpts)
		require.NoError(t, err)
		router.InitBootstrap()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/bootstrap/invalid-token", strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err2 := requestHelper(t, router, req)
		require.NoError(t, err2)
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("403 (already bootstrapped)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		// Generate a bootstrap token using the app's data directory and filesystem
		bootstrapToken, err := auth.GenerateBootstrapToken(router.app.Config.DataDir, router.app.AppFs.Fs)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/bootstrap/"+bootstrapToken.Token, strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAuth_Login(t *testing.T) {
	t.Run("200 (success)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		user := &models.User{
			Username:     "test",
			DisplayName:  "Test",
			PasswordHash: auth.GeneratePassword("abcd1234"),
			Role:         types.UserRoleAdmin,
		}
		require.NoError(t, router.appDao.CreateUser(ctx, user))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.USER_TABLE_ID: user.ID})
		_, err := router.appDao.GetUser(ctx, dbOpts)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "test", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		// Missing both
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Missing password
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Missing username
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"password": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")

		// Both empty
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "", "password": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username and/or password cannot be empty")
	})

	t.Run("401 (invalid user)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		user := &models.User{
			Username:     "test",
			DisplayName:  "Test",
			PasswordHash: auth.GeneratePassword("abcd1234"),
			Role:         types.UserRoleAdmin,
		}
		require.NoError(t, router.appDao.CreateUser(ctx, user))

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "invalid", "password": "abcd1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, status)
		require.Contains(t, string(body), "Invalid username and/or password")
	})

	t.Run("401 (invalid password)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		user := &models.User{
			Username:     "test",
			DisplayName:  "Test",
			PasswordHash: auth.GeneratePassword("abcd1234"),
			Role:         types.UserRoleAdmin,
		}
		require.NoError(t, router.appDao.CreateUser(ctx, user))

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username": "test", "password": "invalid" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, status)
		require.Contains(t, string(body), "Invalid username and/or password")
	})
}
