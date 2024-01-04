package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_GetAssets(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 0, int(respData.TotalItems))
		assert.Len(t, respData.Items, 0)
	})

	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Create 2 courses with 5 assets each with 2 attachments each
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)
		models.NewTestAttachments(t, db, assets, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 10, int(respData.TotalItems))

		// Unmarshal
		assetsResponse := []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		// Assert values. By default orderBy is desc so the last inserted course should be first
		assert.Equal(t, assets[9].ID, assetsResponse[0].ID)
		assert.Equal(t, assets[9].Title, assetsResponse[0].Title)
		assert.Equal(t, assets[9].Path, assetsResponse[0].Path)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Create 2 courses with 5 assets each (10 assets total)
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20asc", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 10, int(respData.TotalItems))

		// Unmarshal
		assetsResponse := []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		assert.Equal(t, assets[0].ID, assetsResponse[0].ID)
		assert.Equal(t, assets[0].Title, assetsResponse[0].Title)
		assert.Equal(t, assets[0].Path, assetsResponse[0].Path)
	})

	t.Run("200 (expand)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Create 5 courses with 5 assets each
		courses := models.NewTestCourses(t, db, 3)
		assets := models.NewTestAssets(t, db, courses, 5)
		models.NewTestAttachments(t, db, assets, 4)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/?orderBy=created_at%20asc&expand=true", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 15, int(respData.TotalItems))

		// Unmarshal
		assetsResponse := []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		assert.Equal(t, assets[0].ID, assetsResponse[0].ID)
		assert.Equal(t, assets[0].Title, assetsResponse[0].Title)
		assert.Equal(t, assets[0].Path, assetsResponse[0].Path)

		// Assert the attachments
		require.NotNil(t, assetsResponse[0].Attachments)
		assert.Len(t, assetsResponse[0].Attachments, 4)
	})

	t.Run("200 (pagination)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Create 3 courses with 6 assets each (18 assets total)
		courses := models.NewTestCourses(t, db, 3)
		assets := models.NewTestAssets(t, db, courses, 6)

		// Get the first 10 assets
		params := url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData pagination.PaginationResult
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 18, int(respData.TotalItems))
		assert.Len(t, respData.Items, 10)

		// Unmarshal
		assetsResponse := []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		assert.Equal(t, assets[0].ID, assetsResponse[0].ID)
		assert.Equal(t, assets[0].Title, assetsResponse[0].Title)
		assert.Equal(t, assets[0].Path, assetsResponse[0].Path)

		// Get the next 8 assets
		params = url.Values{
			"orderBy":                    {"created_at asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"10"},
		}
		resp, err = f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/?"+params.Encode(), nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ = io.ReadAll(resp.Body)

		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		assert.Equal(t, 18, int(respData.TotalItems))
		assert.Len(t, respData.Items, 8)

		// Unmarshal
		assetsResponse = []assetResponse{}
		for _, item := range respData.Items {
			var asset assetResponse
			require.Nil(t, json.Unmarshal(item, &asset))
			assetsResponse = append(assetsResponse, asset)
		}

		assert.Equal(t, assets[10].ID, assetsResponse[0].ID)
		assert.Equal(t, assets[10].Title, assetsResponse[0].Title)
		assert.Equal(t, assets[10].Path, assetsResponse[0].Path)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Drop the assets table
		_, err := db.DB().NewDropTable().Model(&models.Asset{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_GetAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Create 2 courses with 5 assets with 2 attachments
		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[6].ID, nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData assetResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)

		assert.Equal(t, assets[6].ID, respData.ID)
		assert.Equal(t, assets[6].Title, respData.Title)
		assert.Equal(t, assets[6].Path, respData.Path)
		assert.Equal(t, assets[6].CourseID, respData.CourseID)
	})

	t.Run("200 (expand)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 5)
		assets := models.NewTestAssets(t, db, courses, 5)
		models.NewTestAttachments(t, db, assets, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[6].ID+"?expand=true", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var respData assetResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)

		assert.Equal(t, assets[6].ID, respData.ID)
		assert.Equal(t, assets[6].Title, respData.Title)
		assert.Equal(t, assets[6].Path, respData.Path)
		assert.Equal(t, assets[6].CourseID, respData.CourseID)

		// Assert the attachments
		require.NotNil(t, respData.Attachments)
		assert.Len(t, respData.Attachments, 2)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Drop the table
		_, err := db.DB().NewDropTable().Model(&models.Asset{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/test", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_UpdateAsset(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Create 1 courses with 1 asset
		course := models.NewTestCourses(t, db, 1)[0]
		asset := models.NewTestAssets(t, db, []*models.Course{course}, 1)[0]

		// Store the original asset
		origAsset, err := models.GetAssetById(context.Background(), db, nil, asset.ID)
		require.Nil(t, err)

		// Update the asset
		asset.Title = "new title"
		asset.Path = "/new/path"
		asset.Progress = 45
		asset.Completed = true

		data, err := json.Marshal(toAssetResponse([]*models.Asset{asset})[0])
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/"+asset.ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		var respData assetResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)

		assert.Equal(t, origAsset.ID, respData.ID)
		assert.Equal(t, origAsset.CourseID, respData.CourseID)
		assert.Equal(t, origAsset.Title, respData.Title)
		assert.Equal(t, origAsset.Path, respData.Path)

		// Assert the updated values
		assert.Equal(t, 45, respData.Progress)
		assert.True(t, respData.Completed)
		assert.NotNil(t, respData.CompletedAt)
		assert.NotEqual(t, origAsset.UpdatedAt.String(), respData.UpdatedAt.String())
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`bob`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Drop the table
		_, err := db.DB().NewDropTable().Model(&models.Asset{}).Exec(context.Background())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/assets/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAssets_ServeAsset(t *testing.T) {
	t.Run("200 (full video)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		// Create the asset path
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, assets[1].Path, []byte("video"), os.ModePerm))

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[1].ID+"/serve", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "video", string(body))
	})

	t.Run("200 (stream video)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		// Create the asset path
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, assets[1].Path, []byte("video"), os.ModePerm))

		// Create the request and add the range header
		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[1].ID+"/serve", nil)
		req.Header.Add("Range", "bytes=0-")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusPartialContent, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "video", string(body))
	})

	t.Run("400 (invalid path)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 2)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[1].ID+"/serve", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "asset does not exist")
	})

	t.Run("400 (invalid video range)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		courses := models.NewTestCourses(t, db, 2)
		assets := models.NewTestAssets(t, db, courses, 5)

		// Create the asset path
		require.Nil(t, appFs.Fs.MkdirAll(filepath.Dir(assets[1].Path), os.ModePerm))
		require.Nil(t, afero.WriteFile(appFs.Fs, assets[1].Path, []byte("video"), os.ModePerm))

		// Create the request and add the range header
		req := httptest.NewRequest(http.MethodGet, "/api/assets/"+assets[1].ID+"/serve", nil)
		req.Header.Add("Range", "bytes=10-1")

		resp, err := f.Test(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "range start cannot be greater than end")
	})

	t.Run("404 (not found)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/test/serve", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		appFs, db, _, _, teardown := setup(t)
		defer teardown(t)

		f := fiber.New()
		bindAssetsApi(f.Group("/api"), appFs, db)

		// Drop the table
		_, err := db.DB().NewDropTable().Model(&models.Asset{}).Exec(context.Background())
		require.Nil(t, err)

		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/assets/test/serve", nil))
		assert.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
