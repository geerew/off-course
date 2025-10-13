package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestGenerateRecoveryToken(t *testing.T) {
	tempDir := t.TempDir()
	username := "testuser"
	password := "testpass123"

	token, err := GenerateRecoveryToken(username, password, tempDir)
	if err != nil {
		t.Fatalf("Failed to generate recovery token: %v", err)
	}

	// Check token properties
	if token.Username != username {
		t.Errorf("Expected username %s, got %s", username, token.Username)
	}

	if token.Token == "" {
		t.Error("Token should not be empty")
	}

	if token.PasswordHash == "" {
		t.Error("Password hash should not be empty")
	}

	// Check token file exists
	tokenPath := filepath.Join(tempDir, ".recovery-token")
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Error("Token file should exist")
	}

	// Check expiration is in the future
	if time.Now().After(token.ExpiresAt) {
		t.Error("Token should not be expired")
	}

	// Check expiration is within 5 minutes
	expectedExpiry := time.Now().Add(5 * time.Minute)
	if token.ExpiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Error("Token should expire within 5 minutes")
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestValidateRecoveryToken(t *testing.T) {
	tempDir := t.TempDir()
	username := "testuser"
	password := "testpass123"

	// Generate token
	originalToken, err := GenerateRecoveryToken(username, password, tempDir)
	if err != nil {
		t.Fatalf("Failed to generate recovery token: %v", err)
	}

	// Validate token
	validatedToken, err := ValidateRecoveryToken(originalToken.Token, tempDir)
	if err != nil {
		t.Fatalf("Failed to validate recovery token: %v", err)
	}

	// Check token data matches
	if validatedToken.Username != originalToken.Username {
		t.Errorf("Expected username %s, got %s", originalToken.Username, validatedToken.Username)
	}

	if validatedToken.Token != originalToken.Token {
		t.Errorf("Expected token %s, got %s", originalToken.Token, validatedToken.Token)
	}

	if validatedToken.PasswordHash != originalToken.PasswordHash {
		t.Errorf("Password hash mismatch")
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestValidateRecoveryToken_InvalidToken(t *testing.T) {
	tempDir := t.TempDir()
	username := "testuser"
	password := "testpass123"

	// Generate token
	_, err := GenerateRecoveryToken(username, password, tempDir)
	if err != nil {
		t.Fatalf("Failed to generate recovery token: %v", err)
	}

	// Try to validate with wrong token
	_, err = ValidateRecoveryToken("invalid-token", tempDir)
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestValidateRecoveryToken_NoFile(t *testing.T) {
	tempDir := t.TempDir()

	// Try to validate token when no file exists
	_, err := ValidateRecoveryToken("any-token", tempDir)
	if err == nil {
		t.Error("Expected error when token file does not exist")
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDeleteRecoveryToken(t *testing.T) {
	tempDir := t.TempDir()
	username := "testuser"
	password := "testpass123"

	// Generate token
	_, err := GenerateRecoveryToken(username, password, tempDir)
	if err != nil {
		t.Fatalf("Failed to generate recovery token: %v", err)
	}

	// Check token file exists
	tokenPath := filepath.Join(tempDir, ".recovery-token")
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Error("Token file should exist")
	}

	// Delete token
	err = DeleteRecoveryToken(tempDir)
	if err != nil {
		t.Fatalf("Failed to delete recovery token: %v", err)
	}

	// Check token file is gone
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Error("Token file should be deleted")
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDeleteRecoveryToken_NoFile(t *testing.T) {
	tempDir := t.TempDir()

	// Try to delete token when no file exists
	err := DeleteRecoveryToken(tempDir)
	if err != nil {
		t.Errorf("Expected no error when deleting non-existent token file: %v", err)
	}
}
