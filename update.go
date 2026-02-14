package setlist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// GithubAPI is the URL to check for the latest release
	GithubAPI = "https://api.github.com/repos/scottbrown/setlist/releases/latest"
)

// ReleaseInfo represents the GitHub API response for a release
type ReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// CheckForUpdates compares the current version with the latest release
// and returns information about a newer version if available
func CheckForUpdates(ctx context.Context) (*UpdateInfo, error) {
	// Create an HTTP client with reasonable timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GithubAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header as good practice
	req.Header.Set("User-Agent", "Setlist-UpdateCheck/"+VERSION)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check for updates: unexpected status code %d", resp.StatusCode)
	}

	// Parse the response
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse update information: %w", err)
	}

	// Clean up the tag name (remove 'v' prefix if present)
	tagVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(VERSION, "v")

	// Compare versions
	isNewer, err := compareVersions(currentVersion, tagVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to compare versions: %w", err)
	}
	if !isNewer {
		return nil, nil // No update available
	}

	return &UpdateInfo{
		CurrentVersion: VERSION,
		LatestVersion:  release.TagName,
		ReleaseURL:     release.HTMLURL,
		ReleaseDate:    release.PublishedAt,
	}, nil
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	ReleaseDate    time.Time
}

// compareVersions compares two semantic version strings
// Returns true if latestVersion is newer than currentVersion
func compareVersions(currentVersion, latestVersion string) (bool, error) {
	// Split versions by dots
	currentParts := strings.Split(currentVersion, ".")
	latestParts := strings.Split(latestVersion, ".")

	// Compare major, minor, patch
	for i := 0; i < len(currentParts) && i < len(latestParts); i++ {
		current, err := strconv.Atoi(currentParts[i])
		if err != nil {
			return false, fmt.Errorf("invalid version segment %q: %w", currentParts[i], err)
		}
		latest, err := strconv.Atoi(latestParts[i])
		if err != nil {
			return false, fmt.Errorf("invalid version segment %q: %w", latestParts[i], err)
		}

		if latest > current {
			return true, nil
		}
		if current > latest {
			return false, nil
		}
	}

	// If we get here and latestParts has more elements, it's newer
	return len(latestParts) > len(currentParts), nil
}
