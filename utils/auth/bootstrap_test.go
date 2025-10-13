package auth

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestGenerateBootstrapToken(t *testing.T) {
	// Create in-memory filesystem
	appFs := afero.NewMemMapFs()
	tempDir := "/test-data"

	// Generate token
	token, err := GenerateBootstrapToken(tempDir, appFs)
	require.NoError(t, err)
	require.NotNil(t, token)

	// Check token properties
	assert.NotEmpty(t, token.Token)
	assert.True(t, len(token.Token) >= 32)
	assert.True(t, time.Now().Before(token.ExpiresAt))
	assert.True(t, time.Now().After(token.CreatedAt))
	assert.True(t, token.ExpiresAt.After(token.CreatedAt))

	// Check file was created
	tokenPath := filepath.Join(tempDir, ".bootstrap-token")
	exists, err := afero.Exists(appFs, tokenPath)
	require.NoError(t, err)
	assert.True(t, exists)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestValidateBootstrapToken(t *testing.T) {
	// Create in-memory filesystem
	appFs := afero.NewMemMapFs()
	tempDir := "/test-data"

	// Generate token
	originalToken, err := GenerateBootstrapToken(tempDir, appFs)
	require.NoError(t, err)

	// Test valid token
	validatedToken, err := ValidateBootstrapToken(originalToken.Token, tempDir, appFs)
	require.NoError(t, err)
	assert.Equal(t, originalToken.Token, validatedToken.Token)
	assert.True(t, originalToken.ExpiresAt.Equal(validatedToken.ExpiresAt))
	assert.True(t, originalToken.CreatedAt.Equal(validatedToken.CreatedAt))

	// Test invalid token
	_, err = ValidateBootstrapToken("invalid-token", tempDir, appFs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")

	// Test non-existent file
	_, err = ValidateBootstrapToken("any-token", "/non-existent-dir", appFs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap token file not found")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestValidateBootstrapToken_Expired(t *testing.T) {
	// Create in-memory filesystem
	appFs := afero.NewMemMapFs()
	tempDir := "/test-data"

	// Create expired token manually
	expiredToken := &BootstrapToken{
		Token:     "test-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	// Write expired token to file
	tokenPath := filepath.Join(tempDir, ".bootstrap-token")
	tokenData, err := json.Marshal(expiredToken)
	require.NoError(t, err)
	err = afero.WriteFile(appFs, tokenPath, tokenData, 0600)
	require.NoError(t, err)

	// Test expired token
	_, err = ValidateBootstrapToken("test-token", tempDir, appFs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token expired")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDeleteBootstrapToken(t *testing.T) {
	// Create in-memory filesystem
	appFs := afero.NewMemMapFs()
	tempDir := "/test-data"

	// Generate token
	_, err := GenerateBootstrapToken(tempDir, appFs)
	require.NoError(t, err)

	// Check file exists
	tokenPath := filepath.Join(tempDir, ".bootstrap-token")
	exists, err := afero.Exists(appFs, tokenPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete token
	err = DeleteBootstrapToken(tempDir, appFs)
	assert.NoError(t, err)

	// Check file is gone
	exists, err = afero.Exists(appFs, tokenPath)
	require.NoError(t, err)
	assert.False(t, exists)

	// Test deleting non-existent file (should not error)
	err = DeleteBootstrapToken(tempDir, appFs)
	assert.NoError(t, err)
}
