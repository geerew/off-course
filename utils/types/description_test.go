package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDescription_NewDescription(t *testing.T) {
	tests := []struct {
		ext      string
		expected DescriptionType
	}{
		// Supported
		{"md", DescriptionMarkdown},
		{"txt", DescriptionText},
		// Unsupported
		{"", DescriptionUnsupported},
		{"unknown", DescriptionUnsupported},
		{"test", DescriptionUnsupported},
	}

	for _, tt := range tests {
		d := NewDescription(tt.ext)
		require.Equal(t, tt.expected, d.s)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDescription_Set(t *testing.T) {
	d := NewDescription("md")
	require.Equal(t, DescriptionMarkdown, d.s)

	// Set to Text
	d.SetText()
	require.Equal(t, DescriptionText, d.s)

	// Set to Markdown
	d.SetMarkdown()
	require.Equal(t, DescriptionMarkdown, d.s)

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDescription_Is(t *testing.T) {
	// Is Markdown
	d := NewDescription("md")
	require.True(t, d.IsMarkdown())
	require.True(t, d.IsSupported())

	// Is Text
	d = NewDescription("txt")
	require.True(t, d.IsText())
	require.True(t, d.IsSupported())

	// Is Unsupported
	d = NewDescription("unknown")
	require.True(t, d.IsUnsupported())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDescription_String(t *testing.T) {
	d := NewDescription("md")
	require.Equal(t, "markdown", d.String())

	d = NewDescription("txt")
	require.Equal(t, "text", d.String())

	d = NewDescription("unknown")
	require.Equal(t, "unsupported", d.String())

	var zero Description
	require.Equal(t, "", zero.String())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDescription_MarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Supported
		{"md", `"markdown"`},
		{"txt", `"text"`},
		// Unsupported
		{"unknown", `"unsupported"`},
		{"", `"unsupported"`},
	}

	for _, tt := range tests {
		d := NewDescription(tt.input)
		require.NotNil(t, d)

		res, err := d.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, tt.expected, string(res))
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDescription_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected DescriptionType
		err      string
	}{
		// Errors
		{"", "", "unexpected end of JSON input"},
		{"xxx", "", "invalid character 'x' looking for beginning of value"},
		// Unsupported
		{`""`, DescriptionUnsupported, ""},
		{`"bob"`, DescriptionUnsupported, ""},
		// Supported
		{`"markdown"`, DescriptionMarkdown, ""},
		{`"text"`, DescriptionText, ""},
	}

	for _, tt := range tests {
		d := Description{}
		err := d.UnmarshalJSON([]byte(tt.input))

		if tt.err == "" {
			require.Equal(t, tt.expected, d.s)
		} else {
			require.EqualError(t, err, tt.err)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDescription_Value(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Supported
		{"md", "markdown"},
		{"txt", "text"},
		// Unsupported
		{"unknown", "unsupported"},
		{"", "unsupported"},
	}

	for _, tt := range tests {
		d := NewDescription(tt.input)
		require.NotNil(t, d)

		res, err := d.Value()
		require.NoError(t, err)
		require.Equal(t, tt.expected, res)
	}

	// Nil
	d := Description{}
	res, err := d.Value()
	require.NoError(t, err)
	require.Empty(t, res)
	require.True(t, d.IsUnsupported())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDescription_Scan(t *testing.T) {
	tests := []struct {
		value    any
		expected string
	}{
		// Supported
		{"markdown", "markdown"},
		{"text", "text"},
		// Unsupported
		{"", "unsupported"},
		{"unknown", "unsupported"},
		{nil, "unsupported"},
		{"", "unsupported"},
		{"invalid", "unsupported"},
	}

	for _, tt := range tests {
		d := Description{}

		err := d.Scan(tt.value)
		require.NoError(t, err)
		require.Contains(t, d.s, tt.expected)
	}
}
