package utils

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DecodeString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		res, err := DecodeString("")
		require.NoError(t, err)
		require.Equal(t, "", res)
	})

	t.Run("decode error", func(t *testing.T) {
		res, err := DecodeString("`")
		require.EqualError(t, err, "failed to decode path")
		require.Empty(t, res)
	})

	t.Run("unescape error", func(t *testing.T) {
		res, err := DecodeString("dGVzdCUyMDElMiUyNiUyMHRlc3QlMjAy")
		require.EqualError(t, err, "failed to unescape path")
		require.Empty(t, res)
	})

	t.Run("success", func(t *testing.T) {
		res, err := DecodeString("JTJGdGVzdCUyRmRhdGE=")
		require.NoError(t, err)
		require.Equal(t, "/test/data", res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_EncodeString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		res := EncodeString("")
		require.Equal(t, "", res)
	})

	t.Run("success", func(t *testing.T) {
		res := EncodeString("/test/data")
		require.Equal(t, "JTJGdGVzdCUyRmRhdGE=", res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NormalizeWindowsDrive(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Test cases for Windows drive paths
		{"C:", "C:\\"},
		{"C:\\", "C:\\"},
		{"C:folder", "C:\\folder"},
		{"C:\\folder", "C:\\folder"},
	}

	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific tests on non-Windows systems")
	}

	for _, test := range tests {
		got := NormalizeWindowsDrive(test.input)
		require.Equal(t, test.expected, got)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetEnvOr(t *testing.T) {
	// Ensure a clean environment variable first
	_ = os.Unsetenv("UTILS_TEST_ENV")
	require.Equal(t, "default", GetEnvOr("UTILS_TEST_ENV", "default"))

	t.Setenv("UTILS_TEST_ENV", "value")
	require.Equal(t, "value", GetEnvOr("UTILS_TEST_ENV", "default"))
}
