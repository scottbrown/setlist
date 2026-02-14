package setlist

import (
	"strings"
	"testing"
)

func TestFileBuilderErrors(t *testing.T) {
	tests := []struct {
		name          string
		config        ConfigFile
		errorContains string
	}{
		{
			name: "missing session name",
			config: ConfigFile{
				// Missing SessionName
				IdentityStoreId: "test-id",
				Region:          "us-east-1",
			},
			errorContains: "missing required field: SessionName",
		},
		{
			name: "missing identity store and friendly name",
			config: ConfigFile{
				SessionName: "test-session",
				// Missing both IdentityStoreId and FriendlyName
				Region: "us-east-1",
			},
			errorContains: "missing required field: either IdentityStoreId or FriendlyName must be provided",
		},
		{
			name: "missing region",
			config: ConfigFile{
				SessionName:     "test-session",
				IdentityStoreId: "test-id",
				// Missing Region
			},
			errorContains: "missing required field: Region",
		},
		{
			name: "profile with missing fields",
			config: ConfigFile{
				SessionName:     "test-session",
				IdentityStoreId: "test-id",
				Region:          "us-east-1",
				Profiles: []Profile{
					{
						// Missing SessionName
						AccountId: "123456789012",
						RoleName:  "TestRole",
					},
				},
			},
			errorContains: "profile missing required field: SessionName",
		},
		{
			name: "valid configuration should not error",
			config: ConfigFile{
				SessionName:     "test-session",
				IdentityStoreId: "test-id",
				Region:          "us-east-1",
				Profiles: []Profile{
					{
						SessionName: "test-session",
						AccountId:   "123456789012",
						RoleName:    "TestRole",
					},
				},
			},
			errorContains: "", // No error expected
		},
		{
			name: "friendly name only (no identity store) should be valid",
			config: ConfigFile{
				SessionName:  "test-session",
				FriendlyName: "friendly-name", // No IdentityStoreId
				Region:       "us-east-1",
			},
			errorContains: "", // No error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFileBuilder(tt.config)
			_, err := builder.Build()

			if tt.errorContains == "" {
				// No error expected
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				// Error expected
				if err == nil {
					t.Errorf("Expected error containing %q but got nil", tt.errorContains)
					return
				}

				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			}
		})
	}
}

func TestFileBuilderBuildEmptyProfiles(t *testing.T) {
	config := ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-1234567890",
		Region:          "us-east-1",
		Profiles:        []Profile{},
	}

	builder := NewFileBuilder(config)
	payload, err := builder.Build()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have default and sso-session sections but no profile sections
	sections := payload.SectionStrings()
	for _, s := range sections {
		if strings.HasPrefix(s, "profile ") {
			t.Errorf("Expected no profile sections, found: %s", s)
		}
	}
}

func TestFileBuilderBuildWithNicknameMapping(t *testing.T) {
	config := ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-1234567890",
		Region:          "us-east-1",
		Profiles: []Profile{
			{
				SessionName: "test-session",
				AccountId:   "123456789012",
				RoleName:    "AdminRole",
				Description: "Admin access",
			},
		},
		NicknameMapping: map[string]string{
			"123456789012": "prod",
		},
	}

	builder := NewFileBuilder(config)
	payload, err := builder.Build()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have both the account ID profile and the nickname profile
	accountSection := payload.Section("profile 123456789012-AdminRole")
	if accountSection == nil {
		t.Error("Expected account ID profile section")
	}

	nicknameSection := payload.Section("profile prod-AdminRole")
	if nicknameSection == nil {
		t.Error("Expected nickname profile section")
	}
}

func TestFileBuilderBuildProfileMissingAccountId(t *testing.T) {
	config := ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-1234567890",
		Region:          "us-east-1",
		Profiles: []Profile{
			{
				SessionName: "test-session",
				AccountId:   "",
				RoleName:    "TestRole",
			},
		},
	}

	builder := NewFileBuilder(config)
	_, err := builder.Build()
	if err == nil {
		t.Error("Expected error for profile missing AccountId, got nil")
	}
}

func TestFileBuilderBuildProfileMissingRoleName(t *testing.T) {
	config := ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-1234567890",
		Region:          "us-east-1",
		Profiles: []Profile{
			{
				SessionName: "test-session",
				AccountId:   "123456789012",
				RoleName:    "",
			},
		},
	}

	builder := NewFileBuilder(config)
	_, err := builder.Build()
	if err == nil {
		t.Error("Expected error for profile missing RoleName, got nil")
	}
}

func TestFileBuilderBuildOutputContent(t *testing.T) {
	config := ConfigFile{
		SessionName:     "my-sso",
		IdentityStoreId: "d-1234567890",
		Region:          "ca-central-1",
		Profiles: []Profile{
			{
				SessionName:     "my-sso",
				AccountId:       "123456789012",
				RoleName:        "ReadOnly",
				Description:     "Read only access",
				SessionDuration: "PT1H",
			},
		},
	}

	builder := NewFileBuilder(config)
	payload, err := builder.Build()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify the INI content
	var buf strings.Builder
	if _, err := payload.WriteTo(&buf); err != nil {
		t.Fatalf("Failed to write INI: %v", err)
	}
	output := buf.String()

	checks := []string{
		"sso_session",
		"my-sso",
		"https://d-1234567890.awsapps.com/start",
		"ca-central-1",
		"123456789012",
		"ReadOnly",
		"[profile 123456789012-ReadOnly]",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Expected output to contain %q, got:\n%s", check, output)
		}
	}
}
