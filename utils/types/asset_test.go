package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_NewAsset(t *testing.T) {
	// Valid
	tests := []struct {
		ext      string
		expected AssetType
	}{
		// Video
		{"avi", AssetVideo},
		{"mkv", AssetVideo},
		{"flac", AssetVideo},
		{"mp4", AssetVideo},
		{"m4a", AssetVideo},
		{"mp3", AssetVideo},
		{"ogv", AssetVideo},
		{"ogm", AssetVideo},
		{"ogg", AssetVideo},
		{"oga", AssetVideo},
		{"opus", AssetVideo},
		{"webm", AssetVideo},
		{"wav", AssetVideo},
		// document
		{"pdf", AssetPDF},
		// markdown
		{"md", AssetMarkdown},
		// text
		{"txt", AssetText},
	}

	for _, tt := range tests {
		a, err := NewAsset(tt.ext)
		require.NoError(t, err)
		require.Equal(t, tt.expected, a)
	}

	// Invalid
	_, err := NewAsset("test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid asset extension")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_MustAsset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			ext      string
			expected AssetType
		}{
			{"mp4", AssetVideo},
			{"pdf", AssetPDF},
			{"md", AssetMarkdown},
			{"txt", AssetText},
		}

		for _, tt := range tests {
			t.Run(tt.ext, func(t *testing.T) {
				result := MustAsset(tt.ext)
				require.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("panic on invalid", func(t *testing.T) {
		require.Panics(t, func() {
			MustAsset("invalid")
		}, "MustAsset should panic on invalid extension")

		require.Panics(t, func() {
			MustAsset("")
		}, "MustAsset should panic on empty extension")
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Is(t *testing.T) {
	// Is video
	a, _ := NewAsset("mp4")
	require.True(t, a.IsVideo())
	require.True(t, a.IsValid())

	// Is PDF
	a, _ = NewAsset("pdf")
	require.True(t, a.IsPDF())
	require.True(t, a.IsValid())

	// Is Markdown
	a, _ = NewAsset("md")
	require.True(t, a.IsMarkdown())
	require.True(t, a.IsValid())

	// Is Text
	a, _ = NewAsset("txt")
	require.True(t, a.IsText())
	require.True(t, a.IsValid())

	// Is invalid
	invalid := AssetType("invalid")
	require.False(t, invalid.IsValid())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_String(t *testing.T) {
	a, _ := NewAsset("mp4")
	require.Equal(t, "video", a.String())

	a, _ = NewAsset("pdf")
	require.Equal(t, "pdf", a.String())

	a, _ = NewAsset("md")
	require.Equal(t, "markdown", a.String())

	a, _ = NewAsset("txt")
	require.Equal(t, "text", a.String())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_MarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"mp4", `"video"`, false},
		{"pdf", `"pdf"`, false},
		{"md", `"markdown"`, false},
		{"txt", `"text"`, false},
	}

	for _, tt := range tests {
		a, err := NewAsset(tt.input)
		require.NoError(t, err)

		res, err := a.MarshalJSON()
		if tt.hasError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(res))
		}
	}

	// Invalid asset type
	invalid := AssetType("invalid")
	_, err := invalid.MarshalJSON()
	require.Error(t, err)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected AssetType
		err      string
	}{
		// Errors
		{"", "", "unexpected end of JSON input"},
		{"xxx", "", "invalid character 'x' looking for beginning of value"},
		// Invalid asset types
		{`""`, "", "invalid asset type"},
		{`"bob"`, "", "invalid asset type"},
		// Success
		{`"video"`, AssetVideo, ""},
		{`"pdf"`, AssetPDF, ""},
		{`"markdown"`, AssetMarkdown, ""},
		{`"text"`, AssetText, ""},
	}

	for _, tt := range tests {
		var a AssetType
		err := a.UnmarshalJSON([]byte(tt.input))

		if tt.err == "" {
			require.NoError(t, err)
			require.Equal(t, tt.expected, a)
		} else {
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.err)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Value(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"mp4", "video", false},
		{"pdf", "pdf", false},
		{"md", "markdown", false},
		{"txt", "text", false},
	}

	for _, tt := range tests {
		a, err := NewAsset(tt.input)
		require.NoError(t, err)

		res, err := a.Value()
		if tt.hasError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.expected, res)
		}
	}

	// Invalid asset type
	invalid := AssetType("invalid")
	_, err := invalid.Value()
	require.Error(t, err)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Scan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			value    any
			expected AssetType
		}{
			{"video", AssetVideo},
			{"pdf", AssetPDF},
			{"markdown", AssetMarkdown},
			{"text", AssetText},
		}

		for _, tt := range tests {
			var a AssetType

			err := a.Scan(tt.value)
			require.NoError(t, err)
			require.Equal(t, tt.expected, a)
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []struct {
			value any
		}{
			{nil},
			{""},
			{"invalid"},
		}

		for _, tt := range tests {
			var a AssetType

			err := a.Scan(tt.value)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid asset type")
		}
	})
}
