package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var adminResetPasswordCmd = &cobra.Command{
	Use:   "reset-password <username>",
	Short: "Reset password for an admin user",
	Long:  "Reset password for an admin user using recovery tokens. This command communicates with the running application.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]

		fmt.Println()
		fmt.Println("🔐 Admin Password Reset")
		fmt.Println("=======================")
		fmt.Println()

		// Get configuration
		dataDir := viper.GetString("data-dir")
		httpAddr := viper.GetString("http")

		// Verify user exists and is admin
		if err := verifyAdminUser(username, dataDir); err != nil {
			errorMessage("%s", err)
			os.Exit(1)
		}

		// Get new password
		var password string
		for {
			password = questionPassword("New Password")
			if password != "" {
				break
			}
			errorMessage("Password cannot be empty")
		}

		// Confirm password
		for {
			confirmPassword := questionPassword("Confirm Password")
			if confirmPassword == password {
				break
			}
			errorMessage("Passwords do not match")
		}

		fmt.Println()

		// Generate recovery token
		recoveryToken, err := auth.GenerateRecoveryToken(username, password, dataDir)
		if err != nil {
			errorMessage("Failed to generate recovery token: %s", err)
			os.Exit(1)
		}

		// Make HTTP request to running application
		if err := resetPasswordViaAPI(recoveryToken.Token, httpAddr); err != nil {
			// Clean up token file on error
			auth.DeleteRecoveryToken(dataDir)
			errorMessage("Failed to reset password: %s", err)
			os.Exit(1)
		}

		// Clean up token file on success
		auth.DeleteRecoveryToken(dataDir)

		successMessage("✅ Password reset successfully for admin user '%s'", username)
	},
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// verifyAdminUser checks if the user exists and is an admin
func verifyAdminUser(username, dataDir string) error {
	ctx := context.Background()
	appFs := appfs.New(afero.NewOsFs())

	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: dataDir,
		AppFs:   appFs,
		Testing: false,
	})

	if err != nil {
		return fmt.Errorf("failed to create database manager: %w", err)
	}

	dao := dao.New(dbManager.DataDb)
	options := &database.Options{
		Where: squirrel.Eq{models.USER_TABLE_USERNAME: username},
	}

	user, err := dao.GetUser(ctx, options)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user '%s' not found", username)
		}
		return fmt.Errorf("failed to lookup user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user '%s' not found", username)
	}

	if user.Role != types.UserRoleAdmin {
		return fmt.Errorf("user '%s' is not an admin user", username)
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// resetPasswordViaAPI makes an HTTP request to the running application
func resetPasswordViaAPI(token, httpAddr string) error {
	// Prepare request
	requestBody := map[string]string{
		"token": token,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	url := fmt.Sprintf("http://%s/api/admin/recovery", httpAddr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("application is not running or not accessible at %s: %w", httpAddr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("recovery request failed with status %d", resp.StatusCode)
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	adminCmd.AddCommand(adminResetPasswordCmd)
}
