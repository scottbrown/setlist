package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateRequiredFlags(t *testing.T) {
	// Setup a temporary directory for testing file writing
	tempDir, err := os.MkdirTemp("", "setlist-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid writable file path
	validFilePath := filepath.Join(tempDir, "config.txt")

	// Create an invalid directory path
	invalidDir := filepath.Join(tempDir, "nonexistent-dir")

	// Create a test command
	cmd := &cobra.Command{
		Use: "test",
	}

	tests := []struct {
		name        string
		ssoSession  string
		ssoRegion   string
		stdout      bool
		filename    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid_flags_stdout",
			ssoSession:  "mysession",
			ssoRegion:   "us-east-1",
			stdout:      true,
			filename:    "unused",
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "valid_flags_file",
			ssoSession:  "mysession",
			ssoRegion:   "us-east-1",
			stdout:      false,
			filename:    validFilePath,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "missing_sso_session",
			ssoSession:  "",
			ssoRegion:   "us-east-1",
			stdout:      true,
			filename:    "unused",
			wantErr:     true,
			errContains: "required flag --sso-session not set",
		},
		{
			name:        "missing_sso_region",
			ssoSession:  "mysession",
			ssoRegion:   "",
			stdout:      true,
			filename:    "unused",
			wantErr:     true,
			errContains: "required flag --sso-region not set",
		},
		{
			name:        "invalid_region_format",
			ssoSession:  "mysession",
			ssoRegion:   "invalid-region",
			stdout:      true,
			filename:    "unused",
			wantErr:     true,
			errContains: "invalid region format",
		},
		{
			name:        "nonexistent_output_dir",
			ssoSession:  "mysession",
			ssoRegion:   "us-east-1",
			stdout:      false,
			filename:    filepath.Join(invalidDir, "config.txt"),
			wantErr:     true,
			errContains: "output directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the global variables used by validateRequiredFlags
			ssoSession = tt.ssoSession
			ssoRegion = tt.ssoRegion
			stdout = tt.stdout
			filename = tt.filename

			err := validateRequiredFlags(cmd)

			// Check if error is expected
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequiredFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If error is expected, check error message
			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateRequiredFlags() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

// helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
