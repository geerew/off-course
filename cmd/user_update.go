package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a user password",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()

		ctx := context.Background()
		appFs := appfs.New(afero.NewOsFs(), nil)

		dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
			DataDir:  "./oc_data",
			AppFs:    appFs,
			InMemory: false,
		})

		if err != nil {
			errorMessage("Failed to create database manager: %s", err)
			os.Exit(1)
		}

		// Get username
		var username string
		for {
			username = questionPlain("Username")
			if username != "" {
				break
			}

			errorMessage("Username cannot be empty")
		}

		dao := dao.New(dbManager.DataDb)
		options := &database.Options{
			Where: squirrel.Eq{models.USER_TABLE_USERNAME: username},
		}

		user := &models.User{}
		err = dao.GetUser(ctx, user, options)
		if err != nil {
			fmt.Println()

			if err == sql.ErrNoRows {
				errorMessage("User '%s' not found", username)
				os.Exit(1)
			}

			errorMessage("Failed to lookup user: %s", err)
			os.Exit(1)
		}

		// Display name
		user.DisplayName = questionPlainWithDefault("Display Name", user.DisplayName)

		// Role
		var role string
		for {
			role = questionPlainWithDefault("Role", user.Role.String())
			if role == types.UserRoleAdmin.String() || role == types.UserRoleUser.String() {
				break
			}
			errorMessage("Role must be either 'admin' or 'user'")
		}

		// Get password
		password := questionPassword("Password (leave empty to skip)")

		// Confirm password
		if password != "" {
			for {
				pwd := questionPassword("Confirm Password")
				if pwd == password {
					break
				}

				errorMessage("Passwords do not match")
			}

			user.PasswordHash = auth.GeneratePassword(password)
		}

		fmt.Println()

		err = dao.UpdateUser(ctx, user)
		if err != nil {
			errorMessage("Failed to update: %s", err)
			os.Exit(1)
		}

		successMessage("Updated '%s'", username)
	},
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	userCmd.AddCommand(updateCmd)
}
