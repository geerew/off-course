package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_NewScanStatus(t *testing.T) {

	tests := []struct {
		input    ScanStatusType
		expected ScanStatusType
	}{
		{ScanStatusWaiting, ScanStatusWaiting},
		{ScanStatusProcessing, ScanStatusProcessing},
		{"sdf", ScanStatusWaiting},
	}

	for _, tt := range tests {
		s := NewScanStatus(tt.input)
		require.Equal(t, tt.expected, s.s)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_SetWaiting(t *testing.T) {
	s := NewScanStatus(ScanStatusProcessing)
	require.Equal(t, ScanStatusProcessing, s.s)

	s.SetWaiting()
	require.Equal(t, ScanStatusWaiting, s.s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_SetProcessing(t *testing.T) {
	s := NewScanStatus(ScanStatusWaiting)
	require.Equal(t, ScanStatusWaiting, s.s)

	s.SetProcessing()
	require.Equal(t, ScanStatusProcessing, s.s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_IsWaiting(t *testing.T) {
	tests := []struct {
		input    ScanStatusType
		expected bool
	}{
		{ScanStatusWaiting, true},
		{ScanStatusProcessing, false},
	}

	for _, tt := range tests {
		s := NewScanStatus(tt.input)
		require.Equal(t, tt.expected, s.IsWaiting())
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_MarshalJSON(t *testing.T) {
	tests := []struct {
		input    ScanStatusType
		expected string
	}{
		{ScanStatusWaiting, `"waiting"`},
		{ScanStatusProcessing, `"processing"`},
	}

	for _, tt := range tests {
		s := NewScanStatus(tt.input)

		res, err := s.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, tt.expected, string(res))
	}
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
		// Defaults
		{`""`, ScanStatusWaiting, ""},
		{`"bob"`, ScanStatusWaiting, ""},
		// Success
		{`"waiting"`, ScanStatusWaiting, ""},
		{`"processing"`, ScanStatusProcessing, ""},
	}

	for _, tt := range tests {
		ss := ScanStatus{}
		err := ss.UnmarshalJSON([]byte(tt.input))

		if tt.err == "" {
			require.Equal(t, tt.expected, ss.s)
		} else {
			require.EqualError(t, err, tt.err)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_Value(t *testing.T) {
	tests := []struct {
		input    ScanStatusType
		expected string
	}{
		{ScanStatusWaiting, "waiting"},
		{ScanStatusProcessing, "processing"},
	}

	for _, tt := range tests {
		s := NewScanStatus(tt.input)

		res, err := s.Value()
		require.Nil(t, err)
		require.Equal(t, tt.expected, res)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_Scan(t *testing.T) {
	tests := []struct {
		value    any
		expected string
	}{
		{nil, "waiting"},
		{"", "waiting"},
		{"invalid", "waiting"},
		{"waiting", "waiting"},
		{"processing", "processing"},
	}

	for _, tt := range tests {
		ss := ScanStatus{}

		err := ss.Scan(tt.value)
		require.Nil(t, err)
		require.Contains(t, ss.s, tt.expected)
	}
}
