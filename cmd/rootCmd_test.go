package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scottbrown/setlist"

	"github.com/aws/aws-sdk-go-v2/aws"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

func TestDisplayAccounts(t *testing.T) {
	tests := []struct {
		name     string
		accounts []orgtypes.Account
		expected []string
	}{
		{
			name: "multiple accounts",
			accounts: []orgtypes.Account{
				{Id: aws.String("111111111111"), Name: aws.String("Account One")},
				{Id: aws.String("222222222222"), Name: aws.String("Account Two")},
			},
			expected: []string{"111111111111", "Account One", "222222222222", "Account Two"},
		},
		{
			name: "single account",
			accounts: []orgtypes.Account{
				{Id: aws.String("333333333333"), Name: aws.String("Only Account")},
			},
			expected: []string{"333333333333", "Only Account"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := displayAccounts(tt.accounts)
			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected output to contain %q, got: %s", exp, output)
				}
			}
		})
	}
}

func TestOutputConfigStdout(t *testing.T) {
	origStdout := stdout
	defer func() { stdout = origStdout }()
	stdout = true

	cf := setlist.ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-1234567890",
		Region:          "us-east-1",
		Profiles:        []setlist.Profile{},
	}

	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputConfig(cf)
	w.Close()
	os.Stdout = oldOut

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "sso_session") {
		t.Errorf("Expected output to contain sso_session, got: %s", output)
	}
}

func TestOutputConfigFile(t *testing.T) {
	origStdout := stdout
	origFilename := filename
	defer func() {
		stdout = origStdout
		filename = origFilename
	}()

	stdout = false
	tempDir := t.TempDir()
	filename = filepath.Join(tempDir, "test-config")

	cf := setlist.ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-1234567890",
		Region:          "us-east-1",
		Profiles:        []setlist.Profile{},
	}

	err := outputConfig(cf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "sso_session") {
		t.Errorf("Expected file to contain sso_session, got: %s", string(content))
	}
}

func TestOutputConfigInvalidConfig(t *testing.T) {
	origStdout := stdout
	defer func() { stdout = origStdout }()
	stdout = true

	cf := setlist.ConfigFile{}

	err := outputConfig(cf)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}
