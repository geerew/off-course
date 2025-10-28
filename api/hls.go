package api

import (
	"net/http"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media/hls"
	"github.com/gofiber/fiber/v2"
	"github.com/houseme/mobiledetect/ua"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// HLSHandler handles HLS transcoding requests
type HLSHandler struct {
	dao        *dao.DAO
	transcoder *hls.Transcoder
	logger     *logger.Logger
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewHLSHandler creates a new HLS handler
func NewHLSHandler(dao *dao.DAO, transcoder *hls.Transcoder, logger *logger.Logger) *HLSHandler {
	return &HLSHandler{
		dao:        dao,
		transcoder: transcoder,
		logger:     logger,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RegisterHLSRoutes registers HLS routes
func (r *Router) initHlsRoutes() {
	hlsHandler := NewHLSHandler(r.dao, r.config.Transcoder, r.logger.WithHLS())

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
func (api *HLSHandler) GetMaster(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	if assetID == "" {
		api.logger.Error().Msg("GetMaster - asset_id is empty")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "asset_id is required",
		})
	}

	// Get asset with metadata
	asset, err := api.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		api.logger.Error().Err(err).Str("asset_id", assetID).Msg("Failed to get asset")
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Check if transcoder is available
	if api.transcoder == nil {
		api.logger.Error().Msg("Transcoder not initialized")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Transcoder not available",
		})
	}

	ua := ua.New(c.Get("User-Agent"))

	// Get simple master playlist (single stream based on device type)
	master, err := api.transcoder.GetMasterPlaylistSingle(c.Context(), asset.Path, assetID, ua.Mobile())
	if err != nil {
		api.logger.Error().Err(err).Str("asset_id", assetID).Msg("Failed to get master playlist")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate master playlist",
		})
	}

	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.Status(http.StatusOK).SendString(master)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoIndex returns the video index playlist
func (api *HLSHandler) GetVideoIndex(c *fiber.Ctx) error {
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
	asset, err := api.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get video index
	indexPlaylist, err := api.transcoder.GetVideoIndex(c.Context(), asset.Path, uint32(index), quality, assetID)
	if err != nil {
		api.logger.Error().Err(err).Str("asset_id", assetID).Msg("Failed to get video index")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate video index",
		})
	}

	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.Status(http.StatusOK).SendString(indexPlaylist)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetVideoSegment returns a video segment
func (api *HLSHandler) GetVideoSegment(c *fiber.Ctx) error {
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
	asset, err := api.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get video segment
	api.logger.Info().
		Str("path", asset.Path).
		Uint64("index", index).
		Str("quality", qualityStr).
		Int64("segment", segment).
		Msg("Requesting video segment from transcoder")

	segmentPath, err := api.transcoder.GetVideoSegment(c.Context(), asset.Path, uint32(index), quality, int32(segment), assetID)
	if err != nil {
		api.logger.Error().Err(err).Str("asset_id", assetID).Msg("Failed to get video segment")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate video segment",
		})
	}

	api.logger.Info().Str("segment_path", segmentPath).Msg("Video segment ready")

	// Serve the segment file
	return c.Status(http.StatusOK).SendFile(segmentPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioIndex returns the audio index playlist
func (api *HLSHandler) GetAudioIndex(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	indexStr := c.Params("index")

	index, err := strconv.ParseUint(indexStr, 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audio index",
		})
	}

	// Get asset
	asset, err := api.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get audio index
	indexPlaylist, err := api.transcoder.GetAudioIndex(c.Context(), asset.Path, uint32(index), assetID)
	if err != nil {
		api.logger.Error().Err(err).Str("asset_id", assetID).Msg("Failed to get audio index")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate audio index",
		})
	}

	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.Status(http.StatusOK).SendString(indexPlaylist)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAudioSegment returns an audio segment
func (api *HLSHandler) GetAudioSegment(c *fiber.Ctx) error {
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
	asset, err := api.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get audio segment
	api.logger.Info().
		Str("path", asset.Path).
		Uint64("index", index).
		Int64("segment", segment).
		Msg("Requesting audio segment from transcoder")

	segmentPath, err := api.transcoder.GetAudioSegment(c.Context(), asset.Path, uint32(index), int32(segment), assetID)
	if err != nil {
		api.logger.Error().Err(err).Str("asset_id", assetID).Msg("Failed to get audio segment")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate audio segment",
		})
	}

	api.logger.Info().Str("segment_path", segmentPath).Msg("Audio segment ready")

	// Serve the segment file
	return c.Status(http.StatusOK).SendFile(segmentPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetQualities returns the available qualities for a video
func (api *HLSHandler) GetQualities(c *fiber.Ctx) error {
	assetID := c.Params("asset_id")
	if assetID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "asset_id is required",
		})
	}

	// Get asset with metadata
	asset, err := api.dao.GetAsset(c.Context(), database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID}).
		WithAssetMetadata())
	if err != nil {
		api.logger.Error().Err(err).Str("asset_id", assetID).Msg("Failed to get asset")
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Check if transcoder is available
	if api.transcoder == nil {
		api.logger.Error().Msg("Transcoder not initialized")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Transcoder not available",
		})
	}

	// Get available qualities
	qualities, err := api.transcoder.GetQualities(c.Context(), asset.Path, assetID)
	if err != nil {
		api.logger.Error().Err(err).Str("asset_id", assetID).Msg("Failed to get qualities")
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
