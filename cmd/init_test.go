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

func TestValidateRegionOnly(t *testing.T) {
	tests := []struct {
		name        string
		ssoRegion   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid region",
			ssoRegion: "us-east-1",
			wantErr:   false,
		},
		{
			name:        "empty region",
			ssoRegion:   "",
			wantErr:     true,
			errContains: "required flag --sso-region not set",
		},
		{
			name:        "invalid region format",
			ssoRegion:   "invalid-region",
			wantErr:     true,
			errContains: "invalid region format",
		},
		{
			name:      "ca region",
			ssoRegion: "ca-central-1",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssoRegion = tt.ssoRegion

			err := validateRegionOnly()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateRegionOnly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateRegionOnly() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestConfigureLogging(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		logFormat string
		wantErr   bool
	}{
		{
			name:      "plain format default",
			verbose:   false,
			logFormat: "plain",
			wantErr:   false,
		},
		{
			name:      "json format",
			verbose:   false,
			logFormat: "json",
			wantErr:   false,
		},
		{
			name:      "plain format verbose",
			verbose:   true,
			logFormat: "plain",
			wantErr:   false,
		},
		{
			name:      "json format verbose",
			verbose:   true,
			logFormat: "json",
			wantErr:   false,
		},
		{
			name:      "invalid format",
			verbose:   false,
			logFormat: "xml",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origVerbose := verbose
			origLogFormat := logFormat
			defer func() {
				verbose = origVerbose
				logFormat = origLogFormat
			}()

			verbose = tt.verbose
			logFormat = tt.logFormat

			err := configureLogging()

			if (err != nil) != tt.wantErr {
				t.Errorf("configureLogging() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerboseFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup(FlagVerbose)
	if flag == nil {
		t.Errorf("Expected --%s flag to be registered", FlagVerbose)
	}
	if flag.Shorthand != "v" {
		t.Errorf("Expected --%s shorthand to be 'v', got %q", FlagVerbose, flag.Shorthand)
	}
}

func TestLogFormatFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup(FlagLogFormat)
	if flag == nil {
		t.Errorf("Expected --%s flag to be registered", FlagLogFormat)
	}
	if flag.DefValue != "plain" {
		t.Errorf("Expected --%s default to be 'plain', got %q", FlagLogFormat, flag.DefValue)
	}
}

func TestListPermissionSetsFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup(FlagListPermissionSets)
	if flag == nil {
		t.Errorf("Expected --%s flag to be registered", FlagListPermissionSets)
	}
}
