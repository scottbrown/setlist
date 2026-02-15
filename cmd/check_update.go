package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/scottbrown/setlist"

	"github.com/spf13/cobra"
)

var checkUpdateCmd = &cobra.Command{
	Use:   "check-update",
	Short: "Checking if a newer version is available",
	Long:  "Checking GitHub for a newer release of this tool",
	RunE:  handleCheckUpdateCommand,
}

func init() {
	rootCmd.AddCommand(checkUpdateCmd)
}

func handleCheckUpdateCommand(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	return handleCheckUpdate(ctx)
}

func handleCheckUpdate(ctx context.Context) error {
	client := &http.Client{Timeout: 5 * time.Second}
	return checkForUpdate(ctx, client)
}

func checkForUpdate(ctx context.Context, client setlist.HTTPDoer) error {
	info, err := setlist.CheckForUpdates(ctx, client)
	if err != nil {
		return fmt.Errorf("error checking for updates: %w", err)
	}

	if info == nil {
		fmt.Printf("You're running the latest version (v%s)\n", setlist.VERSION)
	} else {
		fmt.Printf("A new version is available!\n")
		fmt.Printf("Current version: v%s\n", info.CurrentVersion)
		fmt.Printf("Latest version:  %s (released %s)\n",
			info.LatestVersion,
			info.ReleaseDate.Format("2006-01-02"))
		fmt.Printf("Download URL:    %s\n", info.ReleaseURL)
	}

	return nil
}
