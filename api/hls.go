package api

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/media/hls"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type hlsAPI struct {
	logger     *slog.Logger
	dao        *dao.DAO
	transcoder *hls.Transcoder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initHlsRoutes initializes the HLS routes
func (r *Router) initHlsRoutes() {
	// Initialize HLS transcoder
	transcoder := r.createHlsTranscoder()
	if transcoder == nil {
		// Skip HLS routes if transcoder is not available
		return
	}

	hlsAPI := hlsAPI{
		logger:     r.config.Logger,
		dao:        r.dao,
		transcoder: transcoder,
	}

	hlsGroup := r.api.Group("/hls")

	// Master playlist
	hlsGroup.Get("/:assetId/master.m3u8", hlsAPI.getMasterPlaylist)

	// Video playlist
	hlsGroup.Get("/:assetId/:quality/index.m3u8", hlsAPI.getVideoPlaylist)

	// Video segment
	hlsGroup.Get("/:assetId/:quality/segment-:segment.ts", hlsAPI.getVideoSegment)

	// Asset info
	hlsGroup.Get("/:assetId/info", hlsAPI.getAssetInfo)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// createHlsTranscoder creates and configures the HLS transcoder
func (r *Router) createHlsTranscoder() *hls.Transcoder {
	// Check if FFmpeg is available
	if r.config.FFmpeg == nil {
		utils.Errf("FFmpeg not configured for HLS transcoding\n")
		return nil
	}

	// Create HLS output directory
	hlsDir := filepath.Join(r.config.DataDir, "hls")
	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		utils.Errf("Failed to create HLS directory: %v\n", err)
	}

	// Detect hardware acceleration
	var hwAccel *hls.HwAccelConfig
	if r.config.Testing {
		// Skip expensive hardware detection in tests
		hwAccel = &hls.HwAccelConfig{Type: hls.HwAccelNone, Available: false}
	} else if r.config.FFmpeg != nil {
		hwAccel = hls.DetectHardwareAcceleration(r.config.FFmpeg.GetFFmpegPath())
	} else {
		hwAccel = &hls.HwAccelConfig{Type: hls.HwAccelNone, Available: false}
	}

	// Create transcoder config
	config := &hls.TranscoderConfig{
		OutputDir:       hlsDir,
		HwAccel:         hwAccel,
		FFmpegPath:      r.config.FFmpeg.GetFFmpegPath(),
		FFProbePath:     r.config.FFmpeg.GetFFProbePath(),
		CleanupInterval: 30 * time.Minute,
		InactiveTimeout: 2 * time.Hour,
	}

	transcoder := hls.NewTranscoder(config)

	// Start cleanup goroutine
	ctx := context.Background()
	transcoder.StartCleanup(ctx)

	return transcoder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getMasterPlaylist returns the master M3U8 playlist for an asset
func (h *hlsAPI) getMasterPlaylist(c *fiber.Ctx) error {
	assetID := c.Params("assetId")
	if assetID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Asset ID is required",
		})
	}

	// Get asset from database
	asset, err := h.getAsset(c.Context(), assetID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Check if asset has keyframes
	keyframes, err := h.getAssetKeyframes(c.Context(), assetID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get keyframes",
		})
	}

	if len(keyframes) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Asset has no keyframes for HLS streaming",
		})
	}

	// Get or create file stream
	fileStream := h.transcoder.GetFileStream(assetID, asset.Path, keyframes)

	// Generate master playlist
	playlistPath, err := fileStream.GenerateMasterPlaylist()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate master playlist",
		})
	}

	// Set content type and serve file
	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.SendFile(playlistPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getVideoPlaylist returns the video M3U8 playlist for an asset and quality
func (h *hlsAPI) getVideoPlaylist(c *fiber.Ctx) error {
	assetID := c.Params("assetId")
	qualityStr := c.Params("quality")

	if assetID == "" || qualityStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Asset ID and quality are required",
		})
	}

	// Parse quality
	quality := hls.Quality(qualityStr)
	if !quality.IsValid() {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid quality",
		})
	}

	// Generate video playlist
	playlistPath, err := h.transcoder.GetVideoPlaylist(assetID, quality)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate video playlist",
		})
	}

	// Set content type and serve file
	c.Set("Content-Type", "application/vnd.apple.mpegurl")
	return c.SendFile(playlistPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getVideoSegment returns a video segment for an asset, quality, and segment number
func (h *hlsAPI) getVideoSegment(c *fiber.Ctx) error {
	assetID := c.Params("assetId")
	qualityStr := c.Params("quality")
	segmentStr := c.Params("segment")

	if assetID == "" || qualityStr == "" || segmentStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Asset ID, quality, and segment are required",
		})
	}

	// Parse quality
	quality := hls.Quality(qualityStr)
	if !quality.IsValid() {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid quality",
		})
	}

	// Parse segment number
	segment, err := strconv.Atoi(segmentStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid segment number",
		})
	}

	// Generate segment
	segmentPath, err := h.transcoder.GetSegment(c.Context(), assetID, quality, segment)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate segment",
		})
	}

	// Set content type and serve file
	c.Set("Content-Type", "video/mp2t")
	return c.SendFile(segmentPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getAssetInfo returns information about an asset's HLS capabilities
func (h *hlsAPI) getAssetInfo(c *fiber.Ctx) error {
	assetID := c.Params("assetId")
	if assetID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Asset ID is required",
		})
	}

	// Get asset from database
	asset, err := h.getAsset(c.Context(), assetID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	// Get keyframes
	keyframes, err := h.getAssetKeyframes(c.Context(), assetID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get keyframes",
		})
	}

	// Get segment count
	segmentCount := len(keyframes)

	// Check if ready
	isReady := h.transcoder.IsReady(assetID)

	// Get ready segments for original quality
	readySegments := h.transcoder.GetReadySegments(assetID, hls.Original)

	// Get stream info
	filePath, totalSegments, ready, err := h.transcoder.GetStreamInfo(assetID)
	if err != nil {
		filePath = asset.Path
		totalSegments = segmentCount
		ready = isReady
	}

	return c.JSON(fiber.Map{
		"asset_id":       assetID,
		"file_path":      filePath,
		"segment_count":  segmentCount,
		"total_segments": totalSegments,
		"ready_segments": readySegments,
		"is_ready":       ready,
		"has_keyframes":  len(keyframes) > 0,
		"keyframe_count": len(keyframes),
		"hls_available":  len(keyframes) > 0,
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getAsset retrieves an asset from the database
func (h *hlsAPI) getAsset(ctx context.Context, assetID string) (*models.Asset, error) {
	dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.ASSET_TABLE_ID: assetID})
	asset, err := h.dao.GetAsset(ctx, dbOpts)
	if err != nil {
		return nil, err
	}

	if asset == nil {
		return nil, fmt.Errorf("asset not found")
	}

	return asset, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getAssetKeyframes retrieves keyframes for an asset
func (h *hlsAPI) getAssetKeyframes(ctx context.Context, assetID string) ([]float64, error) {
	keyframes, err := h.dao.GetAssetKeyframes(ctx, assetID)
	if err != nil {
		return nil, err
	}

	if keyframes == nil {
		return []float64{}, nil
	}

	return keyframes.Keyframes, nil
}
