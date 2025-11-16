package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScans_GetScans(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var scansResp []scanResponse
		err = json.Unmarshal(body, &scansResp)
		require.NoError(t, err)
		require.Zero(t, len(scansResp))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		for i := range 5 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.appDao.CreateCourse(ctx, course))

			_, err := router.app.CourseScan.Add(ctx, course.ID)
			require.NoError(t, err)
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var scansResp []scanResponse
		err = json.Unmarshal(body, &scansResp)
		require.NoError(t, err)
		require.Len(t, scansResp, 5)
	})

	t.Run("403 (not admin)", func(t *testing.T) {
		router, _ := setupUser(t)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, status)
		require.Equal(t, `{"message":"User is not an admin"}`, string(body))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScans_GetScan(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.appDao.CreateCourse(ctx, course))

		scanState, err := router.app.CourseScan.Add(ctx, course.ID)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/scans/"+course.ID, nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData scanResponse
		err = json.Unmarshal(body, &respData)
		require.NoError(t, err)
		require.Equal(t, scanState.ID, respData.ID)
		require.Equal(t, scanState.CourseID, respData.CourseID)
		require.Equal(t, scanState.GetStatus(), respData.Status)
	})

	t.Run("403 (not admin)", func(t *testing.T) {
		router, _ := setupUser(t)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, status)
		require.Equal(t, `{"message":"User is not an admin"}`, string(body))
	})

	t.Run("404 (not found)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		req := httptest.NewRequest(http.MethodGet, "/api/scans/test", nil)
		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScans_CreateScan(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.appDao.CreateCourse(ctx, course))

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(fmt.Sprintf(`{"courseID": "%s"}`, course.ID)))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		var respData scanResponse
		err = json.Unmarshal(body, &respData)
		require.NoError(t, err)
		require.Equal(t, course.ID, respData.CourseID)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "A course ID is required")
	})

	t.Run("400 (invalid course id)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "Invalid course ID")
	})

	t.Run("403 (not admin)", func(t *testing.T) {
		router, _ := setupUser(t)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, status)
		require.Equal(t, `{"message":"User is not an admin"}`, string(body))
	})

}
