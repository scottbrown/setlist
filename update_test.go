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

// mockHTTPClient is a mock implementation of the HTTPDoer interface for testing
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, errors.New("DoFunc not implemented")
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		isNewer        bool
		expectError    bool
	}{
		{"equal versions", "1.2.3", "1.2.3", false, false},
		{"newer major version", "1.2.3", "2.0.0", true, false},
		{"older major version", "2.0.0", "1.2.3", false, false},
		{"newer minor version", "1.2.3", "1.3.0", true, false},
		{"older minor version", "1.3.0", "1.2.0", false, false},
		{"newer patch version", "1.2.3", "1.2.4", true, false},
		{"older patch version", "1.2.3", "1.2.2", false, false},
		{"different lengths", "1.2", "1.2.1", true, false},
		{"different lengths reverse", "1.2.1", "1.2", false, false},
		{"non-numeric current", "1.beta.3", "1.2.3", false, true},
		{"non-numeric latest", "1.2.3", "1.2.beta", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := compareVersions(tt.currentVersion, tt.latestVersion)
			if (err != nil) != tt.expectError {
				t.Errorf("compareVersions(%q, %q) error = %v, expectError %v",
					tt.currentVersion, tt.latestVersion, err, tt.expectError)
				return
			}
			if !tt.expectError && result != tt.isNewer {
				t.Errorf("compareVersions(%q, %q) = %v, want %v",
					tt.currentVersion, tt.latestVersion, result, tt.isNewer)
			}
		})
	}
}

func TestCheckForUpdates(t *testing.T) {
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
			name:           "HTTP error status code",
			currentVersion: "1.0.0",
			mockResponse:   nil,
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
			expectUpdate:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			VERSION = tt.currentVersion

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

			updateInfo, err := CheckForUpdates(context.Background(), mockClient)

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v (error: %v)", tt.expectError, err != nil, err)
			}

			if (updateInfo != nil) != tt.expectUpdate {
				t.Errorf("Expected update: %v, got: %v", tt.expectUpdate, updateInfo != nil)
			}

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

func TestCheckForUpdatesJSONError(t *testing.T) {
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
			}, nil
		},
	}

	_, err := CheckForUpdates(context.Background(), mockClient)
	if err == nil {
		t.Errorf("Expected error for invalid JSON, got nil")
	}
}

func TestCheckForUpdatesHTTPError(t *testing.T) {
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	_, err := CheckForUpdates(context.Background(), mockClient)
	if err == nil {
		t.Errorf("Expected error for HTTP failure, got nil")
	}
}

func TestCheckForUpdatesDevBuild(t *testing.T) {
	originalVersion := VERSION
	defer func() { VERSION = originalVersion }()

	VERSION = "dev"

	mockClient := &mockHTTPClient{}

	_, err := CheckForUpdates(context.Background(), mockClient)
	if err == nil {
		t.Error("Expected error for dev build, got nil")
	}

	expected := "cannot check for updates: running a dev build"
	if err.Error() != expected {
		t.Errorf("Expected error %q, got %q", expected, err.Error())
	}
}

func TestCheckForUpdatesWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, req.Context().Err()
		},
	}

	_, err := CheckForUpdates(ctx, mockClient)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}
