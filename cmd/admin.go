package cmd

import (
	"github.com/spf13/cobra"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// adminCmd represents the admin command
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin management commands",
	Long:  "Commands for managing admin users and system administration.",
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	rootCmd.AddCommand(adminCmd)
}
