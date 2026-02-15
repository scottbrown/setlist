package main

import (
	"fmt"

	"github.com/scottbrown/setlist"

	"github.com/spf13/cobra"
)

var permissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Listing required AWS permissions",
	Long:  "Listing all AWS IAM permissions required by this tool to function correctly",
	RunE:  handlePermissionsCommand,
}

func init() {
	rootCmd.AddCommand(permissionsCmd)
}

func handlePermissionsCommand(cmd *cobra.Command, args []string) error {
	return handleListPermissions()
}

func handleListPermissions() error {
	for _, p := range setlist.ListPermissionsRequired() {
		fmt.Println(p)
	}
	return nil
}
