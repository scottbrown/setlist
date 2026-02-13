package setlist

import (
	"testing"
)

func TestStartUrl(t *testing.T) {
	tt := []struct {
		name            string
		identityStoreID string
		ssoFriendlyName string
		expected        string
	}{
		{
			"has friendly name",
			"d-012345",
			"foo",
			"https://foo.awsapps.com/start",
		},
		{
			"missing friendly name",
			"d-012345",
			"",
			"https://d-012345.awsapps.com/start",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
      identityStoreId, err := NewIdentityStoreId(tc.identityStoreID)
      if err != nil {
        t.Errorf("received nil: %v", err)
      }

			c := ConfigFile{
				IdentityStoreId: identityStoreId,
				FriendlyName:    tc.ssoFriendlyName,
			}
			actual := c.StartURL()

			if actual != tc.expected {
				t.Errorf("unexpected output: got %v, want %v", actual, tc.expected)
			}
		})
	}
}

func TestHasFriendlyName(t *testing.T) {
	tests := []struct {
		name  string
		given string
		want  bool
	}{
		{"has_friendly_name", "foo", true},
		{"has_no_friendly_name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ConfigFile{
				FriendlyName: tt.given,
			}
			got := c.hasFriendlyName()
			if got != tt.want {
				t.Fatalf("Expected %v but got %v", tt.want, got)
			}
		})
	}
}

func TestHasNickname(t *testing.T) {
	mapping := make(map[string]string)
	mapping["123"] = "abc"

	tests := []struct {
		name      string
		accountId string
		expected  bool
	}{
		{"mapping exists", "123", true},
		{"mapping missing", "234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ConfigFile{
				NicknameMapping: mapping,
			}

			got := c.HasNickname(tt.accountId)

			if got != tt.expected {
				t.Fatalf("Expected %v but got %v", tt.expected, got)
			}
		})
	}
}

func TestHasNicknameEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		mapping   map[string]string
		accountId string
		expected  bool
	}{
		{
			"nil mapping",
			nil,
			"123",
			false,
		},
		{
			"empty mapping",
			map[string]string{},
			"123",
			false,
		},
		{
			"account exists",
			map[string]string{"123": "test"},
			"123",
			true,
		},
		{
			"account does not exist",
			map[string]string{"123": "test"},
			"456",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ConfigFile{
				NicknameMapping: tt.mapping,
			}

			got := c.HasNickname(tt.accountId)

			if got != tt.expected {
				t.Errorf("HasNickname() = %v, want %v", got, tt.expected)
			}
		})
	}
}
