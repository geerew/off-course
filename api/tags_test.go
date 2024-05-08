package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Slice of 20 tags for testing (programming languages)
var test_tags = []string{
	"JavaScript", "Python", "Java", "Ruby", "PHP",
	"TypeScript", "C#", "C++", "C", "Swift",
	"Kotlin", "Rust", "Go", "Perl", "Scala",
	"R", "Objective-C", "Shell", "PowerShell", "Haskell",
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTags_GetTags(t *testing.T) {
	t.Run("200 (empty)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, _ := tagsUnmarshalHelper(t, body)
		require.Zero(t, int(paginationResp.TotalItems))
		require.Zero(t, len(paginationResp.Items))
	})

	t.Run("200 (found)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		daos.NewTestBuilder(t).Db(db).Courses(2).Tags([]string{"PHP", "Go", "Java", "TypeScript", "JavaScript"}).Build()

		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp := tagsUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		require.Nil(t, tagsResp[0].Courses)
		require.Equal(t, 2, tagsResp[0].CourseCount)

		// ----------------------------
		// Courses
		// ----------------------------

		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?expand=true", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		require.Len(t, tagsResp[0].Courses, 2)
	})

	t.Run("200 (orderBy)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		daos.NewTestBuilder(t).
			Db(db).
			Courses([]string{"course 1", "course 2"}).
			Tags([]string{"PHP", "Go", "Java", "TypeScript", "JavaScript"}).
			Build()

		// ----------------------------
		// CREATED_AT ASC
		// ----------------------------
		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?orderBy=created_at%20asc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp := tagsUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		require.Equal(t, "PHP", tagsResp[0].Tag)

		// ----------------------------
		// CREATED_AT DESC
		// ----------------------------
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?orderBy=created_at%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		require.Equal(t, "JavaScript", tagsResp[0].Tag)

		// ----------------------------
		// CREATED_AT ASC + COURSES.TITLE DESC
		// ----------------------------
		courseDao := daos.NewCourseDao(db)

		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?expand=true&orderBy=created_at%20asc,"+courseDao.Table()+".title%20desc", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, 5, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 5)
		require.Equal(t, "PHP", tagsResp[0].Tag)
		require.Len(t, tagsResp[0].Courses, 2)
		require.Equal(t, "course 2", tagsResp[0].Courses[0].Title)
	})

	t.Run("200 (filter)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		daos.NewTestBuilder(t).
			Db(db).
			Courses(1).
			Tags([]string{"slightly", "light", "lighter", "highlight", "ghoul", "lightning", "delight"}).
			Build()

		// ----------------------------
		// Filter by `none`
		// ----------------------------
		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?filter=none", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp := tagsUnmarshalHelper(t, body)
		require.Equal(t, 0, int(paginationResp.TotalItems))
		require.Zero(t, tagsResp)

		// ----------------------------
		// Filter by `li`
		// ----------------------------
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?filter=li", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, 6, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 6)
		require.Equal(t, "light", tagsResp[0].Tag)
		require.Equal(t, "lighter", tagsResp[1].Tag)
		require.Equal(t, "lightning", tagsResp[2].Tag)
		require.Equal(t, "delight", tagsResp[3].Tag)
		require.Equal(t, "highlight", tagsResp[4].Tag)
		require.Equal(t, "slightly", tagsResp[5].Tag)

		// ----------------------------
		// Filter by `gh`
		// ----------------------------
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?filter=gh", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, 7, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 7)
		require.Equal(t, "ghoul", tagsResp[0].Tag)
		require.Equal(t, "delight", tagsResp[1].Tag)
		require.Equal(t, "highlight", tagsResp[2].Tag)
		require.Equal(t, "light", tagsResp[3].Tag)
		require.Equal(t, "lighter", tagsResp[4].Tag)
		require.Equal(t, "lightning", tagsResp[5].Tag)
		require.Equal(t, "slightly", tagsResp[6].Tag)

		// ----------------------------
		// Filter by `slight`
		// ----------------------------
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?filter=slight", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, 1, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 1)
		require.Equal(t, "slightly", tagsResp[0].Tag)

		// ----------------------------
		// Case insensitive
		// ----------------------------
		daos.NewTestBuilder(t).
			Db(db).
			Courses(1).
			Tags([]string{"Slight"}).
			Build()

		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?filter=slight", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, 2, int(paginationResp.TotalItems))
		require.Len(t, tagsResp, 2)
		require.Equal(t, "Slight", tagsResp[0].Tag)
		require.Equal(t, "slightly", tagsResp[1].Tag)

	})

	t.Run("200 (pagination)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		for _, tag := range test_tags {
			require.Nil(t, daos.NewTagDao(db).Create(&models.Tag{Tag: tag}, nil))
		}

		// ----------------------------
		// Get the first page (11 tags)
		// ----------------------------
		params := url.Values{
			"orderBy":                    {"tag asc"},
			pagination.PageQueryParam:    {"1"},
			pagination.PerPageQueryParam: {"11"},
		}

		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp := tagsUnmarshalHelper(t, body)
		require.Equal(t, len(test_tags), int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 11)

		// Check the last tag in the paginated response
		require.Equal(t, "Perl", tagsResp[len(paginationResp.Items)-1].Tag)

		// ----------------------------
		// Get the second page (9 tags)
		// ----------------------------
		params = url.Values{
			"orderBy":                    {"tag asc"},
			pagination.PageQueryParam:    {"2"},
			pagination.PerPageQueryParam: {"11"},
		}
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/?"+params.Encode(), nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		paginationResp, tagsResp = tagsUnmarshalHelper(t, body)
		require.Equal(t, len(test_tags), int(paginationResp.TotalItems))
		require.Len(t, paginationResp.Items, 9)

		// Check the last tag in the paginated response
		require.Equal(t, "TypeScript", tagsResp[len(paginationResp.Items)-1].Tag)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		// Drop the courses table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewTagDao(db).Table())
		require.Nil(t, err)

		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTags_GetTag(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(3).Tags([]string{"Go", "PHP"}).Build()

		// By ID
		status, body, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/"+testData[1].Tags[1].TagId, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var tagResp tagResponse
		err = json.Unmarshal(body, &tagResp)
		require.Nil(t, err)
		require.Equal(t, testData[1].Tags[1].TagId, tagResp.ID)
		require.Zero(t, tagResp.Courses)
		require.Equal(t, 3, tagResp.CourseCount)

		// By Name
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/Go/?byName=true", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		err = json.Unmarshal(body, &tagResp)
		require.Nil(t, err)
		require.Equal(t, testData[0].Tags[0].TagId, tagResp.ID)
		require.Zero(t, tagResp.Courses)
		require.Equal(t, 3, tagResp.CourseCount)

		// ----------------------------
		// Courses
		// ----------------------------
		status, body, err = tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/"+testData[1].Tags[1].TagId+"/?expand=true", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		err = json.Unmarshal(body, &tagResp)
		require.Nil(t, err)
		require.Equal(t, 3, tagResp.CourseCount)
		require.Len(t, tagResp.Courses, 3)

	})

	t.Run("404 (not found)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewTagDao(db).Table())
		require.Nil(t, err)

		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodGet, "/api/tags/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTags_CreateTag(t *testing.T) {
	t.Run("201 (created)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(`{"tag": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		var tagResp tagResponse
		err = json.Unmarshal(body, &tagResp)
		require.Nil(t, err)
		require.NotNil(t, tagResp.ID)
		require.Equal(t, "test", tagResp.Tag)
	})

	t.Run("400 (bind error)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(`{`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "error parsing data")
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		// ----------------------------
		// Missing tag
		// ----------------------------
		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(`{"tag": ""}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "a tag is required")
	})

	t.Run("400 (existing tag)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(`{"tag": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, _, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		// Create the tag again
		status, body, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Contains(t, string(body), "tag already exists")
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		// Drop the courses table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewTagDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/tags/", strings.NewReader(`{"tag": "test"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		status, body, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
		require.Contains(t, string(body), "error creating tag")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTags_UpdateTag(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Tags([]string{"Go", "PHP"}).Build()

		tagDao := daos.NewTagDao(db)

		goTag, err := tagDao.Get(testData[0].Tags[0].TagId, false, nil, nil)
		require.Nil(t, err)

		// ----------------------------
		// Update the tag name to 'go'
		// ----------------------------
		goTag.Tag = "go"

		// Marshal the tag for the request
		data, err := json.Marshal(tagResponseHelper([]*models.Tag{goTag})[0])
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/tags/"+goTag.ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var tagResp1 tagResponse
		err = json.Unmarshal(body, &tagResp1)
		require.Nil(t, err)
		require.Equal(t, goTag.ID, tagResp1.ID)
		require.Equal(t, "go", tagResp1.Tag)
	})

	t.Run("400 (invalid data)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		req := httptest.NewRequest(http.MethodPut, "/api/tags/test", strings.NewReader(`bob`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("404 (not found)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		req := httptest.NewRequest(http.MethodPut, "/api/tags/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("400 (duplicate)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses(1).Tags([]string{"Go", "PHP"}).Build()

		tagDao := daos.NewTagDao(db)

		tag, err := tagDao.Get(testData[0].Tags[0].TagId, false, nil, nil)
		require.Nil(t, err)

		tag.Tag = "PHP"

		// Marshal the tag for the request
		data, err := json.Marshal(tagResponseHelper([]*models.Tag{tag})[0])
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/tags/"+tag.ID, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")

		status, body, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)

		require.Contains(t, string(body), "error updating tag - constraint failed: UNIQUE constraint failed:"+fmt.Sprintf(" %s.tag", tagDao.Table()))
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewTagDao(db).Table())
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/tags/test", strings.NewReader(`{"id": "1234567"}`))
		req.Header.Set("Content-Type", "application/json")

		status, _, err := tagsRequestHelper(db, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTags_DeleteTag(t *testing.T) {
	t.Run("204 (deleted)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		tags := []*models.Tag{}
		for _, taggy := range test_tags {
			tag := &models.Tag{Tag: taggy}
			require.Nil(t, daos.NewTagDao(db).Create(tag, nil))
			tags = append(tags, tag)
		}

		tagsDao := daos.NewTagDao(db)

		// ----------------------------
		// Delete tag 3
		// ----------------------------
		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodDelete, "/api/tags/"+tags[2].ID, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)

		result, err := tagsDao.List(&database.DatabaseParams{Where: squirrel.Eq{daos.NewTagDao(db).Table() + ".id": tags[2].ID}}, nil)
		require.Nil(t, err)
		require.Zero(t, len(result))

		// // ----------------------------
		// // Cascades
		// // ----------------------------

		// // Scan
		// _, err = scanDao.Get(testData[2].ID)
		// require.ErrorIs(t, err, sql.ErrNoRows)

		// // Assets
		// count, err := assetsDao.Count(&database.DatabaseParams{Where: squirrel.Eq{daos.Table()Assets() + ".course_id": testData[2].ID}})
		// require.Nil(t, err)
		// require.Zero(t, count)

		// // Attachments
		// count, err = attachmentsDao.Count(&database.DatabaseParams{Where: squirrel.Eq{daos.Table()Attachments() + ".course_id": testData[2].ID}})
		// require.Nil(t, err)
		// require.Zero(t, count)
	})

	t.Run("204 (not found)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodDelete, "/api/tags/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, status)
	})

	t.Run("500 (internal error)", func(t *testing.T) {
		_, db, _, _ := setup(t)

		// Drop the table
		_, err := db.Exec("DROP TABLE IF EXISTS " + daos.NewTagDao(db).Table())
		require.Nil(t, err)

		status, _, err := tagsRequestHelper(db, httptest.NewRequest(http.MethodDelete, "/api/tags/test", nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, status)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func tagsRequestHelper(db database.Database, req *http.Request) (int, []byte, error) {
	f := fiber.New()
	bindTagsApi(f.Group("/api"), db)

	resp, err := f.Test(req)
	if err != nil {
		return -1, nil, err
	}

	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func tagsUnmarshalHelper(t *testing.T, body []byte) (pagination.PaginationResult, []tagResponse) {
	var respData pagination.PaginationResult
	err := json.Unmarshal(body, &respData)
	require.Nil(t, err)

	var tagsResponse []tagResponse
	for _, item := range respData.Items {
		var tag tagResponse
		require.Nil(t, json.Unmarshal(item, &tag))
		tagsResponse = append(tagsResponse, tag)
	}

	return respData, tagsResponse
}
