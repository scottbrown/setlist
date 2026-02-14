package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/scottbrown/setlist"
)

// handleCheckUpdate checks for updates and displays the results
func handleCheckUpdate(ctx context.Context) error {
	client := &http.Client{Timeout: 5 * time.Second}
	return checkForUpdate(ctx, client)
}

// checkForUpdate performs the update check using the provided HTTP client
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
