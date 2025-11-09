package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_NewScanStatusWaiting(t *testing.T) {
	require.Equal(t, ScanStatusWaiting, NewScanStatusWaiting())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_NewScanStatusProcessing(t *testing.T) {
	require.Equal(t, ScanStatusProcessing, NewScanStatusProcessing())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_IsWaiting(t *testing.T) {
	require.True(t, NewScanStatusWaiting().IsWaiting())
	require.False(t, NewScanStatusProcessing().IsWaiting())
	require.False(t, ScanStatusType("").IsWaiting())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_IsProcessing(t *testing.T) {
	require.False(t, NewScanStatusWaiting().IsProcessing())
	require.True(t, NewScanStatusProcessing().IsProcessing())
	require.False(t, ScanStatusType("").IsProcessing())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_IsValid(t *testing.T) {
	require.True(t, NewScanStatusWaiting().IsValid())
	require.True(t, NewScanStatusProcessing().IsValid())
	require.False(t, ScanStatusType("").IsValid())
	require.False(t, ScanStatusType("invalid").IsValid())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_MarshalJSON(t *testing.T) {
	waiting := NewScanStatusWaiting()
	res, err := waiting.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `"waiting"`, string(res))

	processing := NewScanStatusProcessing()
	res, err = processing.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `"processing"`, string(res))

	empty := ScanStatusType("")
	_, err = empty.MarshalJSON()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid scan status")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected ScanStatusType
		err      string
	}{
		// Errors
		{"", "", "unexpected end of JSON input"},
		{"xxx", "", "invalid character 'x' looking for beginning of value"},
		// Invalid scan statuses
		{`""`, "", "invalid scan status"},
		{`"bob"`, "", "invalid scan status"},
		// Success
		{`"waiting"`, ScanStatusWaiting, ""},
		{`"processing"`, ScanStatusProcessing, ""},
	}

	for _, tt := range tests {
		var s ScanStatusType
		err := s.UnmarshalJSON([]byte(tt.input))

		if tt.err == "" {
			require.NoError(t, err)
			require.Equal(t, tt.expected, s)
		} else {
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.err)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_Value(t *testing.T) {
	waiting := NewScanStatusWaiting()
	res, err := waiting.Value()
	require.NoError(t, err)
	require.Equal(t, "waiting", res)

	processing := NewScanStatusProcessing()
	res, err = processing.Value()
	require.NoError(t, err)
	require.Equal(t, "processing", res)

	empty := ScanStatusType("")
	_, err = empty.Value()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid scan status")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_Scan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			value    any
			expected ScanStatusType
		}{
			{"waiting", ScanStatusWaiting},
			{"processing", ScanStatusProcessing},
		}

		for _, tt := range tests {
			var s ScanStatusType

			err := s.Scan(tt.value)
			require.NoError(t, err)
			require.Equal(t, tt.expected, s)
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
			var s ScanStatusType

			err := s.Scan(tt.value)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid scan status")
		}
	})
}
