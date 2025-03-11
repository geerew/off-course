package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUsers_GetUsers(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[userResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t, "admin", types.UserRoleAdmin)

		for i := range 5 {
			users := &models.User{
				Username:     fmt.Sprintf("user %d", i),
				DisplayName:  fmt.Sprintf("User %d", i),
				PasswordHash: security.RandomString(10),
				Role:         types.UserRoleUser,
			}
			require.NoError(t, router.dao.CreateUser(ctx, users))
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, userResp := unmarshalHelper[userResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, userResp, 5)
	})

	t.Run("200 (sort)", func(t *testing.T) {
		router, ctx := setup(t, "admin", types.UserRoleAdmin)

		users := []*models.User{}
		for i := range 5 {
			user := &models.User{
				Username:     fmt.Sprintf("user %d", i),
				DisplayName:  fmt.Sprintf("User %d", i),
				PasswordHash: security.RandomString(10),
				Role:         types.UserRoleUser,
			}
			require.NoError(t, router.dao.CreateUser(ctx, user))
			users = append(users, user)
			time.Sleep(1 * time.Millisecond)
		}

		// CREATED_AT ASC
		q := "sort:\"" + models.USER_TABLE + "." + models.BASE_CREATED_AT + " asc\""
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/?q="+url.QueryEscape(q), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, userResp := unmarshalHelper[userResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, userResp, 5)
		require.Equal(t, users[0].ID, userResp[0].ID)

		// CREATED_AT DESC
		q = "sort:\"" + models.USER_TABLE + "." + models.BASE_CREATED_AT + " desc\""
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/?q="+url.QueryEscape(q), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, userResp = unmarshalHelper[userResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, userResp, 5)
		require.Equal(t, users[4].ID, userResp[0].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		router, ctx := setup(t, "admin", types.UserRoleAdmin)

		users := []*models.User{}
		for i := range 17 {
			user := &models.User{
				Username:     fmt.Sprintf("user %d", i),
				DisplayName:  fmt.Sprintf("User %d", i),
				PasswordHash: security.RandomString(10),
				Role:         types.UserRoleUser,
			}
			require.NoError(t, router.dao.CreateUser(ctx, user))
			users = append(users, user)
		}

		// Get the first page (10 users)
		params := url.Values{
			"q":                          {"sort:\"" + models.USER_TABLE + "." + models.BASE_CREATED_AT + " asc\""},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, userResp := unmarshalHelper[userResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)
		require.Equal(t, users[0].ID, userResp[0].ID)
		require.Equal(t, users[9].ID, userResp[9].ID)

		// Get the second page (7 users)
		params = url.Values{
			"q":                          {"sort:\"" + models.USER_TABLE + "." + models.BASE_CREATED_AT + " asc\""},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, userResp = unmarshalHelper[userResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)
		require.Equal(t, users[10].ID, userResp[0].ID)
		require.Equal(t, users[16].ID, userResp[6].ID)
	})

	t.Run("200 (filter)", func(t *testing.T) {
		router, ctx := setup(t, "admin", types.UserRoleAdmin)

		defaultSort := " sort:\"" + models.USER_TABLE_USERNAME + " asc\""

		users := []*models.User{}
		for i := range 5 {
			role := types.UserRoleUser
			if i%2 == 0 {
				role = types.UserRoleAdmin
			}

			user := &models.User{
				Username:     fmt.Sprintf("user %d", i+1),
				DisplayName:  fmt.Sprintf("User %d", i+1),
				PasswordHash: security.RandomString(10),
				Role:         role,
			}
			require.NoError(t, router.dao.CreateUser(ctx, user))
			users = append(users, user)
		}

		// No filter
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := unmarshalHelper[userResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 5)

		// Username
		q := "user 1 OR user 4 " + defaultSort
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/?q="+url.QueryEscape(q), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, usersResp := unmarshalHelper[userResponse](t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 2)
		require.Equal(t, users[0].ID, usersResp[0].ID)
		require.Equal(t, users[3].ID, usersResp[1].ID)

		// Role
		q = "role:admin " + defaultSort
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/?q="+url.QueryEscape(q), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, usersResp = unmarshalHelper[userResponse](t, body)
		require.Equal(t, 3, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 3)
		require.Equal(t, users[0].ID, usersResp[0].ID)
		require.Equal(t, users[2].ID, usersResp[1].ID)
		require.Equal(t, users[4].ID, usersResp[2].ID)

		// Complex filter
		q = "user 1 OR user 4 AND role:admin " + defaultSort
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/?q="+url.QueryEscape(q), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, usersResp = unmarshalHelper[userResponse](t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 1)
		require.Equal(t, users[0].ID, usersResp[0].ID)
	})

	t.Run("403 (not admin)", func(t *testing.T) {
		router, _ := setup(t, "user", types.UserRoleUser)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, status)
		require.Equal(t, `{"message":"User is not an admin"}`, string(body))
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		// Drop the users table
		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.USER_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/users/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUsers_CreateUser(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		req := httptest.NewRequest(http.MethodPost, "/api/users/", strings.NewReader(`{"username": "admin", "password": "1234"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		req := httptest.NewRequest(http.MethodPost, "/api/users/", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		// Missing username
		req := httptest.NewRequest(http.MethodPost, "/api/users/", strings.NewReader(`{"username": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "A username and password are required")

		// Missing password
		req = httptest.NewRequest(http.MethodPost, "/api/users/", strings.NewReader(`{"username": "admin", "password": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "A username and password are required")
	})

	t.Run("400 (existing user)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		req := httptest.NewRequest(http.MethodPost, "/api/users/", strings.NewReader(`{"username": "admin", "password": "1234" }`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Username already exists")
	})

	t.Run("403 (not admin)", func(t *testing.T) {
		router, _ := setup(t, "user", types.UserRoleUser)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodPost, "/api/users/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, status)
		require.Equal(t, `{"message":"User is not an admin"}`, string(body))
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.USER_TABLE)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/users/", strings.NewReader(`{"username": "admin", "password": "1234"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "Error creating user")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUsers_UpdateUser(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setup(t, "admin", types.UserRoleAdmin)

		user := &models.User{
			Username:     "test",
			DisplayName:  "Test",
			PasswordHash: security.RandomString(10),
			Role:         types.UserRoleUser,
		}
		require.NoError(t, router.dao.CreateUser(ctx, user))

		// Update display name
		req := httptest.NewRequest(http.MethodPut, "/api/users/"+user.ID, strings.NewReader(`{"displayName": "Bob"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		userResult := &models.User{Base: models.Base{ID: user.ID}}
		require.NoError(t, router.dao.GetById(ctx, userResult))
		require.Equal(t, "Bob", userResult.DisplayName)

		// Update password
		req = httptest.NewRequest(http.MethodPut, "/api/users/"+user.ID, strings.NewReader(`{"password": "1234"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		require.NoError(t, router.dao.GetById(ctx, userResult))
		require.Equal(t, "Bob", userResult.DisplayName)
		require.True(t, auth.ComparePassword(userResult.PasswordHash, "1234"))

		// Update role
		req = httptest.NewRequest(http.MethodPut, "/api/users/"+user.ID, strings.NewReader(`{"role": "admin"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err = requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		require.NoError(t, router.dao.GetById(ctx, userResult))
		require.Equal(t, "Bob", userResult.DisplayName)
		require.True(t, auth.ComparePassword(userResult.PasswordHash, "1234"))
		require.Equal(t, types.UserRoleAdmin, userResult.Role)
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		req := httptest.NewRequest(http.MethodPut, "/api/users/invalid", strings.NewReader(`invalid`))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (nothing to update)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		req := httptest.NewRequest(http.MethodPut, "/api/users/invalid", strings.NewReader(`{"invalid": "invalid"}`))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "No data to update")
	})

	t.Run("404 (user not found)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		req := httptest.NewRequest(http.MethodPut, "/api/users/invalid", strings.NewReader(`{"displayName": "Admin"}`))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
		require.Contains(t, string(body), "User not found")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.USER_TABLE)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/users/invalid", strings.NewReader(`{"username": "admin", "password": "1234"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "Error looking up user")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUsers_DeleteUser(t *testing.T) {
	t.Run("204 (deleted)", func(t *testing.T) {
		router, ctx := setup(t, "admin", types.UserRoleAdmin)

		users := []*models.User{}
		for i := range 5 {
			u := &models.User{
				Username:     fmt.Sprintf("user %d", i),
				DisplayName:  fmt.Sprintf("User %d", i),
				PasswordHash: security.RandomString(10),
				Role:         types.UserRoleUser,
			}
			require.NoError(t, router.dao.CreateUser(ctx, u))
			users = append(users, u)
		}

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/users/"+users[1].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		user := &models.User{Base: models.Base{ID: users[1].ID}}
		err = router.dao.GetById(ctx, user)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("204 (not found)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/users/invalid", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)
	})

	t.Run("403 (not admin)", func(t *testing.T) {
		router, _ := setup(t, "user", types.UserRoleUser)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/users/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, status)
		require.Equal(t, `{"message":"User is not an admin"}`, string(body))
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setup(t, "admin", types.UserRoleAdmin)

		_, err := router.config.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.USER_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodDelete, "/api/users/invalid", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}
