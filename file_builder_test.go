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
