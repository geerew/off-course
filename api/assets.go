package api

import (
	"database/sql"
	"fmt"
	"log/slog"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assets struct {
	logger           *slog.Logger
	appFs            *appFs.AppFs
	assetDao         *daos.AssetDao
	assetProgressDao *daos.AssetProgressDao
	attachmentDao    *daos.AttachmentDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetResponse struct {
	ID        string      `json:"id"`
	CourseID  string      `json:"courseId"`
	Title     string      `json:"title"`
	Prefix    int         `json:"prefix"`
	Chapter   string      `json:"chapter"`
	Path      string      `json:"path"`
	Type      types.Asset `json:"assetType"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`

	// Progress
	VideoPos    int       `json:"videoPos"`
	Completed   bool      `json:"completed"`
	CompletedAt time.Time `json:"completedAt"`

	// Attachments
	Attachments []*attachmentResponse `json:"attachments,omitempty"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const bufferSize = 1024 * 8                 // 8KB per chunk, adjust as needed
const maxInitialChunkSize = 1024 * 1024 * 5 // 5MB, adjust as needed

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *assets) getAssets(c *fiber.Ctx) error {
	expand := c.QueryBool("expand", false)
	orderBy := c.Query("orderBy", "created_at desc")

	dbParams := &database.DatabaseParams{
		OrderBy:    strings.Split(orderBy, ","),
		Pagination: pagination.NewFromApi(c),
	}

	if expand {
		dbParams.IncludeRelations = []string{api.attachmentDao.Table()}
	}

	assets, err := api.assetDao.List(dbParams, nil)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "error looking up assets", err)
	}

	pResult, err := dbParams.Pagination.BuildResult(assetResponseHelper(assets))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *assets) getAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	expand := c.QueryBool("expand", false)

	dbParams := &database.DatabaseParams{}
	if expand {
		dbParams.IncludeRelations = []string{api.attachmentDao.Table()}
	}

	// TODO: support attachments orderby
	asset, err := api.assetDao.Get(id, dbParams, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	return c.Status(fiber.StatusOK).JSON(assetResponseHelper([]*models.Asset{asset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *assets) updateAsset(c *fiber.Ctx) error {
	id := c.Params("id")

	// Parse the request body to get the updated fields
	reqAsset := &assetResponse{}
	if err := c.BodyParser(reqAsset); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	// Create an asset progress
	ap := &models.AssetProgress{
		AssetID:   id,
		CourseID:  reqAsset.CourseID,
		VideoPos:  reqAsset.VideoPos,
		Completed: reqAsset.Completed,
	}

	// Update the asset progress
	if err := api.assetProgressDao.Update(ap, nil); err != nil {
		if err == sql.ErrNoRows || strings.HasPrefix(err.Error(), "constraint failed: FOREIGN KEY constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid course ID", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error updating asset", err)
	}

	// Get the updated asset
	asset, err := api.assetDao.Get(id, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	return c.Status(fiber.StatusOK).JSON(assetResponseHelper([]*models.Asset{asset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *assets) serveAsset(c *fiber.Ctx) error {
	id := c.Params("id")

	asset, err := api.assetDao.Get(id, nil, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	// Check for invalid path
	if exists, err := afero.Exists(api.appFs.Fs, asset.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not exist", nil)
	}

	if asset.Type.IsVideo() {
		return handleVideo(c, api.appFs, asset)
	} else if asset.Type.IsHTML() {
		return handleHtml(c, api.appFs, asset)
	}

	// TODO: Handle PDF
	return c.Status(fiber.StatusOK).SendString("done")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func assetResponseHelper(assets []*models.Asset) []*assetResponse {
	responses := []*assetResponse{}
	for _, asset := range assets {
		responses = append(responses, &assetResponse{
			ID:        asset.ID,
			CourseID:  asset.CourseID,
			Title:     asset.Title,
			Prefix:    int(asset.Prefix.Int16),
			Chapter:   asset.Chapter,
			Path:      asset.Path,
			Type:      asset.Type,
			CreatedAt: asset.CreatedAt,
			UpdatedAt: asset.UpdatedAt,

			// Progress
			VideoPos:    asset.VideoPos,
			Completed:   asset.Completed,
			CompletedAt: asset.CompletedAt,

			// Association
			Attachments: attachmentResponseHelper(asset.Attachments),
		})

	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// handleVideo handles the video streaming logic
func handleVideo(c *fiber.Ctx, appFs *appFs.AppFs, asset *models.Asset) error {
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
func handleHtml(c *fiber.Ctx, appFs *appFs.AppFs, asset *models.Asset) error {
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
