package api

import (
	"net/http"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/media/hls"
	"github.com/gofiber/fiber/v2"
	"github.com/houseme/mobiledetect/ua"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// HLSHandler handles HLS transcoding requests
type HLSHandler struct {
	dao        *dao.DAO
	transcoder *hls.Transcoder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewHLSHandler creates a new HLS handler
func NewHLSHandler(dao *dao.DAO, transcoder *hls.Transcoder) *HLSHandler {
	return &HLSHandler{
		dao:        dao,
		transcoder: transcoder,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RegisterHLSRoutes registers HLS routes
func (r *Router) initHlsRoutes() {
	// Initialize transcoder with DAO
	transcoder, err := hls.NewTranscoder(r.dao)
	if err != nil {
		panic("Failed to create HLS transcoder: " + err.Error())
	}
	hlsHandler := NewHLSHandler(r.dao, transcoder)

	// Master playlist
	r.api.Get("/hls/:asset_id/master.m3u8", hlsHandler.GetMaster)

	// Video streams
	r.api.Get("/hls/:asset_id/video/:index/:quality/index.m3u8", hlsHandler.GetVideoIndex)
	r.api.Get("/hls/:asset_id/video/:index/:quality/segment-:num.ts", hlsHandler.GetVideoSegment)

	// Audio streams
	r.api.Get("/hls/:asset_id/audio/:index/index.m3u8", hlsHandler.GetAudioIndex)
	r.api.Get("/hls/:asset_id/audio/:index/segment-:num.ts", hlsHandler.GetAudioSegment)

	// Qualities endpoint
	r.api.Get("/hls/:asset_id/qualities", hlsHandler.GetQualities)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetMaster returns the master playlist (single stream based on device type)
func (h *HLSHandler) GetMaster(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	if assetID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "asset_id is required",
		})
	}

	// Get asset with metadata
	asset, err := h.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		utils.Errf("Failed to get asset %s: %v\n", assetID, err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Check if transcoder is available
	if h.transcoder == nil {
		utils.Errf("Transcoder not initialized\n")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Transcoder not available",
		})
	}

	ua := ua.New(c.Get("User-Agent"))

	// Get simple master playlist (single stream based on device type)
	master, err := h.transcoder.GetMasterPlaylistSingle(c.Context(), asset.Path, assetID, ua.Mobile())
	if err != nil {
		utils.Errf("Failed to get master playlist for asset %s: %v\n", assetID, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate master playlist",
		})
	}

	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.SendString(master)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoIndex returns the video index playlist
func (h *HLSHandler) GetVideoIndex(c *fiber.Ctx) error {
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
	asset, err := h.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get video index
	indexPlaylist, err := h.transcoder.GetVideoIndex(c.Context(), asset.Path, uint32(index), quality, assetID)
	if err != nil {
		utils.Errf("Failed to get video index for asset %s: %v\n", assetID, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate video index",
		})
	}

	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.SendString(indexPlaylist)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoSegment returns a video segment
func (h *HLSHandler) GetVideoSegment(c *fiber.Ctx) error {
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
	asset, err := h.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get video segment
	utils.Infof("API (hls): Requesting video segment from transcoder: path=%s, index=%d, quality=%s, segment=%d\n",
		asset.Path, index, qualityStr, segment)

	segmentPath, err := h.transcoder.GetVideoSegment(c.Context(), asset.Path, uint32(index), quality, int32(segment), assetID)
	if err != nil {
		utils.Errf("Failed to get video segment for asset %s: %v\n", assetID, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate video segment",
		})
	}

	utils.Infof("API (hls): Video segment ready: %s\n", segmentPath)

	// Serve the segment file
	return c.SendFile(segmentPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioIndex returns the audio index playlist
func (h *HLSHandler) GetAudioIndex(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	indexStr := c.Params("index")

	index, err := strconv.ParseUint(indexStr, 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audio index",
		})
	}

	// Get asset
	asset, err := h.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get audio index
	indexPlaylist, err := h.transcoder.GetAudioIndex(c.Context(), asset.Path, uint32(index), assetID)
	if err != nil {
		utils.Errf("Failed to get audio index for asset %s: %v\n", assetID, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate audio index",
		})
	}

	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.SendString(indexPlaylist)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioSegment returns an audio segment
func (h *HLSHandler) GetAudioSegment(c *fiber.Ctx) error {
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
	asset, err := h.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get audio segment
	utils.Infof("API (hls): Requesting audio segment from transcoder: path=%s, index=%d, segment=%d\n",
		asset.Path, index, segment)

	segmentPath, err := h.transcoder.GetAudioSegment(c.Context(), asset.Path, uint32(index), int32(segment), assetID)
	if err != nil {
		utils.Errf("Failed to get audio segment for asset %s: %v\n", assetID, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate audio segment",
		})
	}

	utils.Infof("API (hls): Audio segment ready: %s\n", segmentPath)

	// Serve the segment file
	return c.SendFile(segmentPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetQualities returns the available qualities for a video
func (h *HLSHandler) GetQualities(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	if assetID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "asset_id is required",
		})
	}

	// Get asset with metadata
	asset, err := h.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		utils.Errf("Failed to get asset %s: %v\n", assetID, err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Check if transcoder is available
	if h.transcoder == nil {
		utils.Errf("Transcoder not initialized\n")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Transcoder not available",
		})
	}

	// Get available qualities
	qualities, err := h.transcoder.GetQualities(c.Context(), asset.Path, assetID)
	if err != nil {
		utils.Errf("Failed to get qualities for asset %s: %v\n", assetID, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get available qualities",
		})
	}

	// Convert qualities to strings for JSON response
	qualityStrings := make([]string, len(qualities))
	for i, quality := range qualities {
		qualityStrings[i] = string(quality)
	}

	return c.JSON(fiber.Map{
		"qualities": qualityStrings,
	})
}
