package cardcache

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/media"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestNewCardCache(t *testing.T) {
	t.Run("creates cache directory", func(t *testing.T) {
		appFs := appfs.New(afero.NewMemMapFs())
		testLogger := logger.NilLogger()

		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			t.Skip("FFmpeg not available for testing")
		}

		cache, err := NewCardCache(&CardCacheConfig{
			CachePath: "/test",
			AppFs:     appFs,
			Logger:    testLogger,
			FFmpeg:    ffmpeg,
		})

		require.NoError(t, err)
		require.NotNil(t, cache)

		// Verify cache path is set
		cachePath := cache.GetCardPath("test")
		cachePath = filepath.Dir(cachePath) // Get the cache directory
		require.Contains(t, cachePath, "cards")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestEnsureFallbackCard(t *testing.T) {
	t.Run("generates fallback card", func(t *testing.T) {
		appFs := appfs.New(afero.NewOsFs())
		testLogger := logger.NilLogger()

		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			t.Skip("FFmpeg not available for testing")
		}

		// Create temporary directory
		tmpDir := t.TempDir()

		cache, err := NewCardCache(&CardCacheConfig{
			CachePath: tmpDir,
			AppFs:     appFs,
			Logger:    testLogger,
			FFmpeg:    ffmpeg,
		})
		require.NoError(t, err)

		fallbackPath := cache.GetFallbackPath()

		err = cache.EnsureFallbackCard(fallbackPath)
		require.NoError(t, err)

		// Verify file exists
		exists, err := cache.CardExists(fallbackPath)
		require.NoError(t, err)
		require.True(t, exists)

		// Verify file is WebP
		require.True(t, filepath.Ext(fallbackPath) == ".webp")
	})

	t.Run("is idempotent", func(t *testing.T) {
		appFs := appfs.New(afero.NewOsFs())
		testLogger := logger.NilLogger()

		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			t.Skip("FFmpeg not available for testing")
		}

		tmpDir := t.TempDir()

		cache, err := NewCardCache(&CardCacheConfig{
			CachePath: tmpDir,
			AppFs:     appFs,
			Logger:    testLogger,
			FFmpeg:    ffmpeg,
		})
		require.NoError(t, err)

		fallbackPath := cache.GetFallbackPath()

		// Call multiple times - should succeed each time
		err = cache.EnsureFallbackCard(fallbackPath)
		require.NoError(t, err)

		err = cache.EnsureFallbackCard(fallbackPath)
		require.NoError(t, err)

		// Verify file still exists
		exists, err := cache.CardExists(fallbackPath)
		require.NoError(t, err)
		require.True(t, exists)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestGenerateOptimizedCard(t *testing.T) {
	t.Run("generates optimized card from JPEG", func(t *testing.T) {
		appFs := appfs.New(afero.NewOsFs())
		testLogger := logger.NilLogger()

		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			t.Skip("FFmpeg not available for testing")
		}

		tmpDir := t.TempDir()

		cache, err := NewCardCache(&CardCacheConfig{
			CachePath: tmpDir,
			AppFs:     appFs,
			Logger:    testLogger,
			FFmpeg:    ffmpeg,
		})
		require.NoError(t, err)

		// Create a test image file using FFmpeg to ensure it's valid
		testImagePath := filepath.Join(tmpDir, "test.jpg")
		// Use FFmpeg to create a valid test JPEG
		ffmpegPath := ffmpeg.GetFFmpegPath()
		createImageCmd := exec.Command(ffmpegPath,
			"-f", "lavfi",
			"-i", "color=c=red:s=100x100:d=1",
			"-frames:v", "1",
			"-y",
			testImagePath,
		)
		err = createImageCmd.Run()
		if err != nil {
			t.Skipf("Failed to create test image: %v", err)
		}

		outputPath := filepath.Join(tmpDir, "optimized.webp")
		ctx := context.Background()

		err = cache.GenerateOptimizedCard(ctx, testImagePath, outputPath)
		require.NoError(t, err)

		// Verify file exists
		exists, err := cache.CardExists(outputPath)
		require.NoError(t, err)
		require.True(t, exists)

		// Verify file is WebP
		require.True(t, filepath.Ext(outputPath) == ".webp")
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		appFs := appfs.New(afero.NewOsFs())
		testLogger := logger.NilLogger()

		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			t.Skip("FFmpeg not available for testing")
		}

		tmpDir := t.TempDir()

		cache, err := NewCardCache(&CardCacheConfig{
			CachePath: tmpDir,
			AppFs:     appFs,
			Logger:    testLogger,
			FFmpeg:    ffmpeg,
		})
		require.NoError(t, err)

		// Create a test image file using FFmpeg to ensure it's valid
		testImagePath := filepath.Join(tmpDir, "test.jpg")
		ffmpegPath := ffmpeg.GetFFmpegPath()
		createImageCmd := exec.Command(ffmpegPath,
			"-f", "lavfi",
			"-i", "color=c=red:s=100x100:d=1",
			"-frames:v", "1",
			"-y",
			testImagePath,
		)
		err = createImageCmd.Run()
		if err != nil {
			t.Skipf("Failed to create test image: %v", err)
		}

		outputPath := filepath.Join(tmpDir, "optimized.webp")
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err = cache.GenerateOptimizedCard(ctx, testImagePath, outputPath)
		require.Error(t, err)
		require.Equal(t, context.Canceled, err)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDeleteCard(t *testing.T) {
	t.Run("deletes existing card", func(t *testing.T) {
		appFs := appfs.New(afero.NewOsFs())
		testLogger := logger.NilLogger()

		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			t.Skip("FFmpeg not available for testing")
		}

		tmpDir := t.TempDir()

		cache, err := NewCardCache(&CardCacheConfig{
			CachePath: tmpDir,
			AppFs:     appFs,
			Logger:    testLogger,
			FFmpeg:    ffmpeg,
		})
		require.NoError(t, err)

		// Create a test file
		testCardPath := filepath.Join(tmpDir, "test.webp")
		err = os.WriteFile(testCardPath, []byte("test"), 0644)
		require.NoError(t, err)

		// Delete it
		err = cache.DeleteCard(testCardPath)
		require.NoError(t, err)

		// Verify it's gone
		exists, err := cache.CardExists(testCardPath)
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("handles non-existent card gracefully", func(t *testing.T) {
		appFs := appfs.New(afero.NewOsFs())
		testLogger := logger.NilLogger()

		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			t.Skip("FFmpeg not available for testing")
		}

		tmpDir := t.TempDir()

		cache, err := NewCardCache(&CardCacheConfig{
			CachePath: tmpDir,
			AppFs:     appFs,
			Logger:    testLogger,
			FFmpeg:    ffmpeg,
		})
		require.NoError(t, err)

		// Try to delete non-existent file
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.webp")
		err = cache.DeleteCard(nonExistentPath)
		require.NoError(t, err) // Should not error
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestGetCardPath(t *testing.T) {
	t.Run("returns correct path format", func(t *testing.T) {
		appFs := appfs.New(afero.NewMemMapFs())
		testLogger := logger.NilLogger()

		ffmpeg, err := media.NewFFmpeg()
		if err != nil {
			t.Skip("FFmpeg not available for testing")
		}

		cache, err := NewCardCache(&CardCacheConfig{
			CachePath: "/test",
			AppFs:     appFs,
			Logger:    testLogger,
			FFmpeg:    ffmpeg,
		})
		require.NoError(t, err)

		courseID := "test-course-id"
		cardPath := cache.GetCardPath(courseID)

		require.Contains(t, cardPath, courseID)
		require.True(t, filepath.Ext(cardPath) == ".webp")
	})
}
