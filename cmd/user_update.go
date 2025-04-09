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
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// TODO Support updating the role and display name

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
		err = dao.Get(ctx, user, options)
		if err != nil {
			fmt.Println()

			if err == sql.ErrNoRows {
				errorMessage("User '%s' not found", username)
				os.Exit(1)
			}

			errorMessage("Failed to lookup user: %s", err)
			os.Exit(1)
		}

		// Get password
		var password string
		for {
			password = questionPassword("Password")
			if password != "" {
				break
			}

			errorMessage("Password cannot be empty")
		}

		// Confirm password
		for {
			pwd := questionPassword("Confirm Password")
			if pwd == password {
				break
			}

			errorMessage("Passwords do not match")
		}

		fmt.Println()

		user.PasswordHash = auth.GeneratePassword(password)

		err = dao.UpdateUser(ctx, user)
		if err != nil {
			errorMessage("Failed to update password: %s", err)
			os.Exit(1)
		}

		successMessage("Password updated for '%s'", username)
	},
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	userCmd.AddCommand(updateCmd)
}
