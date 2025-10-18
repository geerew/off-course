package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media"
	"github.com/geerew/off-course/utils/media/hls"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	// Cache FFmpeg instance to avoid expensive re-initialization
	cachedFFmpeg *media.FFmpeg
	ffmpegOnce   sync.Once

	// Cache hardware acceleration detection to avoid expensive FFmpeg calls
	cachedHwAccel *hls.HwAccelConfig
	hwAccelOnce   sync.Once
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getCachedFFmpeg returns a cached FFmpeg instance, initializing it only once
func getCachedFFmpeg(t *testing.T) *media.FFmpeg {
	ffmpegOnce.Do(func() {
		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			// Skip tests if FFmpeg is not available
			t.Skip("FFmpeg not available for testing")
		}
		cachedFFmpeg = ffmpeg
	})
	return cachedFFmpeg
}

// getCachedHwAccel returns a cached hardware acceleration config, initializing it only once
func getCachedHwAccel(t *testing.T) *hls.HwAccelConfig {
	hwAccelOnce.Do(func() {
		ffmpeg := getCachedFFmpeg(t)
		cachedHwAccel = hls.DetectHardwareAcceleration(ffmpeg.GetFFmpegPath())
	})
	return cachedHwAccel
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T, id string, role types.UserRole) (*Router, context.Context) {
	t.Helper()

	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.TestWriteFn(&logs, &logsMux),
	})
	require.NoError(t, err, "Failed to initialize logger")

	appFs := appfs.New(afero.NewMemMapFs(), logger)

	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: "./oc_data",
		AppFs:   appFs,
		Testing: true,
	})

	require.NoError(t, err)
	require.NotNil(t, dbManager)

	// Get cached FFmpeg instance
	ffmpeg := getCachedFFmpeg(t)

	courseScan := coursescan.New(&coursescan.CourseScanConfig{
		Db:     dbManager.DataDb,
		AppFs:  appFs,
		Logger: logger,
		FFmpeg: ffmpeg,
	})

	// Router
	config := &RouterConfig{
		DbManager:     dbManager,
		AppFs:         appFs,
		CourseScan:    courseScan,
		FFmpeg:        ffmpeg,
		Logger:        logger,
		SignupEnabled: true,
		Testing:       true, // Skip expensive operations in tests
	}

	router := devRouter(config, id, role)

	// create the user
	user := models.User{
		Base: models.Base{
			ID: id,
		},
		Username:     id,
		Role:         role,
		PasswordHash: "password",
		DisplayName:  "Test User",
	}
	require.NoError(t, router.dao.CreateUser(context.Background(), &user))

	// Initialize bootstrap status after creating user
	router.InitBootstrap()

	ctx := context.Background()
	principal := types.Principal{
		UserID: id,
		Role:   role,
	}
	ctx = context.WithValue(ctx, types.PrincipalContextKey, principal)

	return router, ctx
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func requestHelper(t *testing.T, router *Router, req *http.Request) (int, []byte, error) {
	t.Helper()

	resp, err := router.App.Test(req)
	if err != nil {
		return -1, nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func unmarshalHelper[T any](t *testing.T, body []byte) (pagination.PaginationResult, []T) {
	t.Helper()

	var respData pagination.PaginationResult
	err := json.Unmarshal(body, &respData)
	require.NoError(t, err)

	var resp []T
	for _, item := range respData.Items {
		var r T
		require.Nil(t, json.Unmarshal(item, &r))
		resp = append(resp, r)
	}

	return respData, resp
}
