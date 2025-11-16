package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCardExtension_IsValid(t *testing.T) {
	t.Run("valid extensions", func(t *testing.T) {
		valid := []CardExtension{
			CardExtensionJPG,
			CardExtensionJPEG,
			CardExtensionPNG,
			CardExtensionWebP,
			CardExtensionTIFF,
		}

		for _, ext := range valid {
			require.True(t, ext.IsValid(), "extension %s should be valid", ext)
		}
	})

	t.Run("invalid extensions", func(t *testing.T) {
		invalid := []CardExtension{
			CardExtension("gif"),
			CardExtension("bmp"),
			CardExtension("svg"),
			CardExtension(""),
			CardExtension("mp4"),
		}

		for _, ext := range invalid {
			require.False(t, ext.IsValid(), "extension %s should be invalid", ext)
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestCardExtension_String(t *testing.T) {
	tests := []struct {
		ext      CardExtension
		expected string
	}{
		{CardExtensionJPG, "jpg"},
		{CardExtensionJPEG, "jpeg"},
		{CardExtensionPNG, "png"},
		{CardExtensionWebP, "webp"},
		{CardExtensionTIFF, "tiff"},
		{CardExtension("unknown"), "unknown"},
	}

	for _, tt := range tests {
		require.Equal(t, tt.expected, tt.ext.String())
	}
}
