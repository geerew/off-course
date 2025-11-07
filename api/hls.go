package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/media/hls"
	"github.com/gofiber/fiber/v2"
	"github.com/houseme/mobiledetect/ua"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// hlsAPI handles HLS transcoding requests
type hlsAPI struct {
	r *Router
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RegisterHLSRoutes registers HLS routes
func (r *Router) initHlsRoutes() {

	hlsApi := hlsAPI{r: r}

	g := r.apiGroup("hls")

	// Master playlist
	g.Get("/:asset_id/master.m3u8", hlsApi.GetMaster)

	// Video streams
	g.Get("/:asset_id/video/:index/:quality/index.m3u8", hlsApi.GetVideoIndex)
	g.Get("/:asset_id/video/:index/:quality/segment-:num.ts", hlsApi.GetVideoSegment)

	// Audio streams
	g.Get("/:asset_id/audio/:index/index.m3u8", hlsApi.GetAudioIndex)
	g.Get("/:asset_id/audio/:index/segment-:num.ts", hlsApi.GetAudioSegment)

	// Qualities endpoint
	g.Get("/:asset_id/qualities", hlsApi.GetQualities)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetMaster returns the master playlist (single stream based on device type)
func (api *hlsAPI) GetMaster(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	if assetID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "asset_id is required",
		})
	}

	// Get asset with metadata
	asset, err := api.getAssetWithMetadata(c.Context(), assetID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Check if transcoder is available
	if api.r.app.Transcoder == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Transcoder not available",
		})
	}

	ua := ua.New(c.Get("User-Agent"))

	// Get simple master playlist (single stream based on device type)
	master, err := api.r.app.Transcoder.GetMasterPlaylistSingle(c.Context(), asset.Path, assetID, ua.Mobile())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate master playlist",
		})
	}

	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.Status(http.StatusOK).SendString(master)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoIndex returns the video index playlist
func (api *hlsAPI) GetVideoIndex(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	indexStr := c.Params("index")
	qualityStr := c.Params("quality")

	index, err := strconv.ParseUint(indexStr, 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid video index",
		})
	}

	quality, err := hls.QualityFromString(qualityStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quality",
		})
	}

	// Get asset
	asset, err := api.getAssetWithMetadata(c.Context(), assetID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get video index
	indexPlaylist, err := api.r.app.Transcoder.GetVideoIndex(c.Context(), asset.Path, uint32(index), quality, assetID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate video index",
		})
	}

	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.Status(http.StatusOK).SendString(indexPlaylist)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoSegment returns a video segment
func (api *hlsAPI) GetVideoSegment(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	indexStr := c.Params("index")
	qualityStr := c.Params("quality")
	segmentStr := c.Params("num")

	index, err := strconv.ParseUint(indexStr, 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid video index",
		})
	}

	quality, err := hls.QualityFromString(qualityStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quality",
		})
	}

	segment, err := strconv.ParseInt(segmentStr, 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid segment number",
		})
	}

	// Get asset
	asset, err := api.getAssetWithMetadata(c.Context(), assetID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get video segment
	segmentPath, err := api.r.app.Transcoder.GetVideoSegment(c.Context(), asset.Path, uint32(index), quality, int32(segment), assetID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate video segment",
		})
	}

	// Serve the segment file
	return c.Status(http.StatusOK).SendFile(segmentPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioIndex returns the audio index playlist
func (api *hlsAPI) GetAudioIndex(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	indexStr := c.Params("index")

	index, err := strconv.ParseUint(indexStr, 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audio index",
		})
	}

	// Get asset
	asset, err := api.getAssetWithMetadata(c.Context(), assetID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get audio index
	indexPlaylist, err := api.r.app.Transcoder.GetAudioIndex(c.Context(), asset.Path, uint32(index), assetID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate audio index",
		})
	}

	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.Status(http.StatusOK).SendString(indexPlaylist)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioSegment returns an audio segment
func (api *hlsAPI) GetAudioSegment(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	indexStr := c.Params("index")
	segmentStr := c.Params("num")

	index, err := strconv.ParseUint(indexStr, 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audio index",
		})
	}

	segment, err := strconv.ParseInt(segmentStr, 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid segment number",
		})
	}

	// Get asset
	asset, err := api.getAssetWithMetadata(c.Context(), assetID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get audio segment
	segmentPath, err := api.r.app.Transcoder.GetAudioSegment(c.Context(), asset.Path, uint32(index), int32(segment), assetID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate audio segment",
		})
	}

	// Serve the segment file
	return c.Status(http.StatusOK).SendFile(segmentPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetQualities returns the available qualities for a video
func (api *hlsAPI) GetQualities(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	if assetID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "asset_id is required",
		})
	}

	// Get asset with metadata
	asset, err := api.getAssetWithMetadata(c.Context(), assetID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Check if transcoder is available
	if api.r.app.Transcoder == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Transcoder not available",
		})
	}

	// Get available qualities
	qualities, err := api.r.app.Transcoder.GetQualities(c.Context(), asset.Path, assetID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get available qualities",
		})
	}

	// Convert qualities to strings for JSON response
	qualityStrings := make([]string, len(qualities))
	for i, quality := range qualities {
		qualityStrings[i] = string(quality)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"qualities": qualityStrings,
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getAssetWithMetadata retrieves an asset with its metadata by asset ID
func (api *hlsAPI) getAssetWithMetadata(ctx context.Context, assetID string) (*models.Asset, error) {
	return api.r.appDao.GetAsset(ctx, database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
}
