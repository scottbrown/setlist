package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/scottbrown/setlist"
)

func TestCheckForUpdate(t *testing.T) {
	originalVersion := setlist.VERSION
	defer func() { setlist.VERSION = originalVersion }()

	tests := []struct {
		name         string
		version      string
		statusCode   int
		response     *setlist.ReleaseInfo
		httpErr      error
		expectError  bool
		expectOutput []string
	}{
		{
			name:       "newer version available",
			version:    "1.0.0",
			statusCode: http.StatusOK,
			response: &setlist.ReleaseInfo{
				TagName:     "v1.1.0",
				Name:        "Release 1.1.0",
				PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				HTMLURL:     "https://github.com/scottbrown/setlist/releases/tag/v1.1.0",
			},
			expectError:  false,
			expectOutput: []string{"A new version is available", "v1.1.0"},
		},
		{
			name:       "already up to date",
			version:    "1.1.0",
			statusCode: http.StatusOK,
			response: &setlist.ReleaseInfo{
				TagName:     "v1.1.0",
				Name:        "Release 1.1.0",
				PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				HTMLURL:     "https://github.com/scottbrown/setlist/releases/tag/v1.1.0",
			},
			expectError:  false,
			expectOutput: []string{"latest version"},
		},
		{
			name:        "HTTP error",
			version:     "1.0.0",
			httpErr:     errors.New("network error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setlist.VERSION = tt.version

			mockClient := &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					if tt.httpErr != nil {
						return nil, tt.httpErr
					}
					bodyBytes, _ := json.Marshal(tt.response)
					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
					}, nil
				},
			}

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := checkForUpdate(context.Background(), mockClient)
			w.Close()
			os.Stdout = oldStdout

			if (err != nil) != tt.expectError {
				t.Errorf("checkForUpdate() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				var buf bytes.Buffer
				buf.ReadFrom(r)
				output := buf.String()

				for _, exp := range tt.expectOutput {
					if !strings.Contains(output, exp) {
						t.Errorf("Expected output to contain %q, got: %s", exp, output)
					}
				}
			}
		})
	}
}

func TestHandleCheckUpdateCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := handleCheckUpdate(ctx)
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
}
