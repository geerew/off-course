package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
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

		paginationResp, _ := unmarshalHelper[scanResponse](t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		for i := range 5 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.appDao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			require.NoError(t, router.appDao.CreateScan(ctx, scan))
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, coursesResp := unmarshalHelper[scanResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, coursesResp, 5)
	})

	t.Run("200 (sort)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		scans := []*models.Scan{}
		for i := range 5 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i+1), Path: fmt.Sprintf("/course %d", i+1)}
			require.NoError(t, router.appDao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			require.NoError(t, router.appDao.CreateScan(ctx, scan))
			scans = append(scans, scan)
			time.Sleep(1 * time.Millisecond)
		}

		// CREATED_AT ASC
		q := "sort:\"" + models.SCAN_TABLE_CREATED_AT + " asc\""
		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/?q="+url.QueryEscape(q), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, scanResp := unmarshalHelper[scanResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, scanResp, 5)
		require.Equal(t, scans[0].ID, scanResp[0].ID)

		// CREATED_AT DESC
		q = "sort:\"" + models.SCAN_TABLE_CREATED_AT + " desc\""
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/?q="+url.QueryEscape(q), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, scanResp = unmarshalHelper[scanResponse](t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, scanResp, 5)
		require.Equal(t, scans[4].ID, scanResp[0].ID)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		scans := []*models.Scan{}
		for i := range 17 {
			course := &models.Course{Title: fmt.Sprintf("course %d", i), Path: fmt.Sprintf("/course %d", i)}
			require.NoError(t, router.appDao.CreateCourse(ctx, course))

			scan := &models.Scan{CourseID: course.ID}
			require.NoError(t, router.appDao.CreateScan(ctx, scan))
			scans = append(scans, scan)

			time.Sleep(1 * time.Millisecond)
		}

		// Page 1 (10 scans)
		params := url.Values{
			"q":                          {"sort:\"" + models.SCAN_TABLE_CREATED_AT + " asc\""},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, scanResp := unmarshalHelper[scanResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 10)
		require.Equal(t, scans[0].ID, scanResp[0].ID)
		require.Equal(t, scans[9].ID, scanResp[9].ID)

		// Page 2 (7 scans)
		params = url.Values{
			"q":                          {"sort:\"" + models.SCAN_TABLE_CREATED_AT + " asc\""},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		status, body, err = requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, scanResp = unmarshalHelper[scanResponse](t, body)
		require.Equal(t, 17, int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 7)
		require.Equal(t, scans[10].ID, scanResp[0].ID)
		require.Equal(t, scans[16].ID, scanResp[6].ID)
	})

	t.Run("403 (not admin)", func(t *testing.T) {
		router, _ := setupUser(t)

		status, body, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, status)
		require.Equal(t, `{"message":"User is not an admin"}`, string(body))
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		// Drop the courses table
		_, err := router.app.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.SCAN_TABLE)
		require.NoError(t, err)

		status, _, err := requestHelper(t, router, httptest.NewRequest(http.MethodGet, "/api/scans/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScans_GetScan(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router, ctx := setupAdmin(t)

		course := &models.Course{Title: "Course 1", Path: "/Course 1"}
		require.NoError(t, router.appDao.CreateCourse(ctx, course))

		scan := &models.Scan{CourseID: course.ID}
		require.NoError(t, router.appDao.CreateScan(ctx, scan))

		req := httptest.NewRequest(http.MethodGet, "/api/scans/"+course.ID, nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData scanResponse
		err = json.Unmarshal(body, &respData)
		require.NoError(t, err)
		require.Equal(t, scan.ID, respData.ID)
		require.Equal(t, scan.CourseID, respData.CourseID)
		require.Equal(t, scan.Status, respData.Status)
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

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		_, err := router.app.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.SCAN_TABLE)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/scans/test", nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "Error looking up scan")
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

	t.Run("500 (internal error)", func(t *testing.T) {
		router, _ := setupAdmin(t)

		_, err := router.app.DbManager.DataDb.Exec("DROP TABLE IF EXISTS " + models.COURSE_TABLE)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/scans/", strings.NewReader(`{"courseID": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "Error creating scan job")
	})
}
