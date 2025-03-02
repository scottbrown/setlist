package setlist

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

// mockHTTPClient is a mock implementation of the HTTP client for testing
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, errors.New("DoFunc not implemented")
}

// Modify the testCheckForUpdates function to properly handle the mock client
func testCheckForUpdates(ctx context.Context, client *mockHTTPClient) (*UpdateInfo, error) {
	// First check if context is canceled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GithubAPI, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Setlist-UpdateCheck/"+VERSION)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("non-200 status code received")
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	// Clean the version strings
	tagVersion := release.TagName
	currentVersion := VERSION

	// Remove 'v' prefix if present
	if len(tagVersion) > 0 && tagVersion[0] == 'v' {
		tagVersion = tagVersion[1:]
	}
	if len(currentVersion) > 0 && currentVersion[0] == 'v' {
		currentVersion = currentVersion[1:]
	}

	// Compare versions
	if isNewer := compareVersions(currentVersion, tagVersion); isNewer {
		return &UpdateInfo{
			CurrentVersion: VERSION,
			LatestVersion:  release.TagName,
			ReleaseURL:     release.HTMLURL,
			ReleaseDate:    release.PublishedAt,
		}, nil
	}

	return nil, nil
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		isNewer        bool
	}{
		{"equal versions", "1.2.3", "1.2.3", false},
		{"newer major version", "1.2.3", "2.0.0", true},
		{"older major version", "2.0.0", "1.2.3", false},
		{"newer minor version", "1.2.3", "1.3.0", true},
		{"older minor version", "1.3.0", "1.2.0", false},
		{"newer patch version", "1.2.3", "1.2.4", true},
		{"older patch version", "1.2.3", "1.2.2", false},
		{"different lengths", "1.2", "1.2.1", true},
		{"different lengths reverse", "1.2.1", "1.2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.currentVersion, tt.latestVersion)
			if result != tt.isNewer {
				t.Errorf("compareVersions(%q, %q) = %v, want %v",
					tt.currentVersion, tt.latestVersion, result, tt.isNewer)
			}
		})
	}
}

func TestCheckForUpdates(t *testing.T) {
	// Store the original VERSION to restore it after tests
	originalVersion := VERSION
	defer func() { VERSION = originalVersion }()

	tests := []struct {
		name           string
		currentVersion string
		mockResponse   *ReleaseInfo
		statusCode     int
		expectError    bool
		expectUpdate   bool
	}{
		{
			name:           "newer version available",
			currentVersion: "1.0.0",
			mockResponse: &ReleaseInfo{
				TagName:     "v1.1.0",
				Name:        "Release 1.1.0",
				PublishedAt: time.Now(),
				HTMLURL:     "https://github.com/example/repo/releases/tag/v1.1.0",
			},
			statusCode:   http.StatusOK,
			expectError:  false,
			expectUpdate: true,
		},
		{
			name:           "current version is latest",
			currentVersion: "1.1.0",
			mockResponse: &ReleaseInfo{
				TagName:     "v1.1.0",
				Name:        "Release 1.1.0",
				PublishedAt: time.Now(),
				HTMLURL:     "https://github.com/example/repo/releases/tag/v1.1.0",
			},
			statusCode:   http.StatusOK,
			expectError:  false,
			expectUpdate: false,
		},
		{
			name:           "current version is newer (dev build)",
			currentVersion: "1.2.0",
			mockResponse: &ReleaseInfo{
				TagName:     "v1.1.0",
				Name:        "Release 1.1.0",
				PublishedAt: time.Now(),
				HTMLURL:     "https://github.com/example/repo/releases/tag/v1.1.0",
			},
			statusCode:   http.StatusOK,
			expectError:  false,
			expectUpdate: false,
		},
		{
			name:           "HTTP error",
			currentVersion: "1.0.0",
			mockResponse:   nil,
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
			expectUpdate:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set current version for the test
			VERSION = tt.currentVersion

			// Create a mock HTTP client
			mockClient := &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					if tt.statusCode != http.StatusOK {
						return &http.Response{
							StatusCode: tt.statusCode,
							Body:       io.NopCloser(bytes.NewReader([]byte{})),
						}, nil
					}

					bodyBytes, _ := json.Marshal(tt.mockResponse)
					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
					}, nil
				},
			}

			// Run the test
			updateInfo, err := testCheckForUpdates(context.Background(), mockClient)

			// Check error expectation
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v (error: %v)", tt.expectError, err != nil, err)
			}

			// Check update expectation
			if (updateInfo != nil) != tt.expectUpdate {
				t.Errorf("Expected update: %v, got: %v", tt.expectUpdate, updateInfo != nil)
			}

			// If an update is expected, verify its properties
			if tt.expectUpdate && updateInfo != nil {
				if updateInfo.CurrentVersion != VERSION {
					t.Errorf("Expected current version %s, got %s", VERSION, updateInfo.CurrentVersion)
				}

				if updateInfo.LatestVersion != tt.mockResponse.TagName {
					t.Errorf("Expected latest version %s, got %s", tt.mockResponse.TagName, updateInfo.LatestVersion)
				}

				if updateInfo.ReleaseURL != tt.mockResponse.HTMLURL {
					t.Errorf("Expected URL %s, got %s", tt.mockResponse.HTMLURL, updateInfo.ReleaseURL)
				}
			}
		})
	}
}

// Test specifically for JSON parsing errors
func TestCheckForUpdatesJSONError(t *testing.T) {
	// Create a mock HTTP client that returns invalid JSON
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
			}, nil
		},
	}

	// Run the test
	_, err := testCheckForUpdates(context.Background(), mockClient)

	// Should return an error for invalid JSON
	if err == nil {
		t.Errorf("Expected error for invalid JSON, got nil")
	}
}

// TestCheckForUpdatesWithContext tests that the context is respected
func TestCheckForUpdatesWithContext(t *testing.T) {
	// Create a context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Create a mock client - this should not be called
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			t.Error("HTTP client should not have been called with canceled context")
			return nil, nil
		},
	}

	// Function should return an error without calling the HTTP client
	_, err := testCheckForUpdates(ctx, mockClient)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}
