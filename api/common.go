package api

import (
	"fmt"
	"mime"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/queryparser"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
var (
	defaultCoursesOrderBy                = []string{models.COURSE_TABLE_CREATED_AT + " desc"}
	defaultCourseAssetsOrderBy           = []string{models.ASSET_TABLE_CHAPTER + " asc", models.ASSET_TABLE_PREFIX + " asc"}
	defaultCourseAssetAttachmentsOrderBy = []string{models.ATTACHMENT_TABLE_TITLE + " asc"}
	defaultTagsOrderBy                   = []string{models.TAG_TABLE_TAG + " asc"}
	defaultUsersOrderBy                  = []string{models.USER_TABLE_CREATED_AT + " desc"}
	defaultLogsOrderBy                   = []string{models.LOG_TABLE_CREATED_AT + " desc"}
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// errorResponse is a helper method to return an error response
func errorResponse(c *fiber.Ctx, status int, message string, err error) error {
	resp := fiber.Map{
		"message": message,
	}

	if err != nil {
		resp["error"] = err.Error()
	}

	return c.Status(status).JSON(resp)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// builderOptions is a struct to hold the options for the optionsBuilder
type builderOptions struct {
	// A default order by clause to use if none is found in the query
	DefaultOrderBy []string

	// A slice of allowed filters to match on in the query
	AllowedFilters []string

	// Whether to paginate the results
	Paginate bool

	// A function to run after the query has been parsed. It will only run if the query is not nil
	AfterParseHook func(*queryparser.QueryResult, *database.Options, string)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// optionsBuilder builds a database.Options based on a `q` query parameter
func optionsBuilder(c *fiber.Ctx, builderOptions builderOptions, userId string) (*database.Options, error) {
	options := &database.Options{}

	orderBy := []string{models.BASE_CREATED_AT + " desc"}
	if len(builderOptions.DefaultOrderBy) > 0 {
		orderBy = builderOptions.DefaultOrderBy
	}
	options.OrderBy = orderBy

	if builderOptions.Paginate {
		options.Pagination = pagination.NewFromApi(c)
	}

	q := c.Query("q", "")
	if q == "" {
		return options, nil
	}

	parsed, err := queryparser.Parse(q, builderOptions.AllowedFilters)
	if err != nil {
		return nil, err
	}

	if parsed == nil {
		return options, nil
	}

	if len(parsed.Sort) > 0 {
		options.OrderBy = parsed.Sort
	}

	if builderOptions.AfterParseHook != nil {
		builderOptions.AfterParseHook(parsed, options, userId)
	}

	return options, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const bufferSize = 1024 * 8                 // 8KB per chunk, adjust as needed
const maxInitialChunkSize = 1024 * 1024 * 5 // 5MB, adjust as needed

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// handleVideo handles the video streaming logic
func handleVideo(c *fiber.Ctx, appFs *appfs.AppFs, asset *models.Asset) error {
	// Open the video
	file, err := appFs.Fs.Open(asset.Path)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error opening file", err)
	}
	defer file.Close()

	// Get the file info
	fileInfo, err := file.Stat()
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting file info", err)
	}

	// Get the range header and return the entire video if there is no range header
	rangeHeader := c.Get("Range", "")
	if rangeHeader == "" {
		return filesystem.SendFile(c, afero.NewHttpFs(appFs.Fs), asset.Path)
	}

	// Parse the "bytes=START-END" format
	bytesPos := strings.Split(rangeHeader, "=")[1]
	rangeStartEnd := strings.Split(bytesPos, "-")
	start, _ := strconv.Atoi(rangeStartEnd[0])
	var end int

	if len(rangeStartEnd) == 2 && rangeStartEnd[0] == "0" && rangeStartEnd[1] == "1" {
		start = 0
		end = 1
	} else {
		end = start + maxInitialChunkSize - 1
		if end >= int(fileInfo.Size()) {
			end = int(fileInfo.Size()) - 1
		}
	}

	if start > end {
		return errorResponse(c, fiber.StatusBadRequest, "Range start cannot be greater than end", fmt.Errorf("range start is greater than end"))
	}

	// Setting required response headers
	c.Set(fiber.HeaderContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
	c.Set(fiber.HeaderContentLength, strconv.Itoa(end-start+1))
	c.Set(fiber.HeaderContentType, mime.TypeByExtension(filepath.Ext(asset.Path)))
	c.Set(fiber.HeaderAcceptRanges, "bytes")

	// Set the status code to 206 Partial Content
	c.Status(fiber.StatusPartialContent)

	file.Seek(int64(start), 0)
	buffer := make([]byte, bufferSize)
	bytesToSend := end - start + 1

	// Respond in chunks
	for bytesToSend > 0 {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			break
		}

		if bytesRead > bytesToSend {
			bytesRead = bytesToSend
		}

		c.Write(buffer[:bytesRead])
		bytesToSend -= bytesRead
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// handleHtml handles serving HTML files
func handleHtml(c *fiber.Ctx, appFs *appfs.AppFs, asset *models.Asset) error {
	// Open the HTML file
	file, err := appFs.Fs.Open(asset.Path)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error opening file", err)
	}
	defer file.Close()

	// Read the content of the HTML file
	content, err := afero.ReadAll(file)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error reading file", err)
	}

	c.Set(fiber.HeaderContentType, "text/html")
	return c.Status(fiber.StatusOK).Send(content)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// protectedRoute protects a route
//
// Example:
//
//	group.Get("/my-route", protectedRoute, myHandler)
func protectedRoute(c *fiber.Ctx) error {
	userRole, ok := c.Locals(types.RoleContextKey).(string)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user", nil)
	}

	if userRole != types.UserRoleAdmin.String() {
		return errorResponse(c, fiber.StatusForbidden, "User is not an admin", nil)
	}

	return c.Next()
}
