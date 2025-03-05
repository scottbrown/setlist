package setlist

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseNicknameMapping(t *testing.T) {
	tt := []struct {
		name     string
		mapping  string
		expected map[string]string
	}{
		{
			"knowngood",
			"123456789012=foo,123456789013=bar",
			map[string]string{
				"123456789012": "foo",
				"123456789013": "bar",
			},
		},
		{
			"empty mapping",
			"",
			map[string]string{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, _ := ParseNicknameMapping(tc.mapping)

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("unexpected output: got %v, want %v", actual, tc.expected)
			}
		})
	}
}

// Tests for error cases in ParseNicknameMapping
func TestParseNicknameMappingErrors(t *testing.T) {
	tests := []struct {
		name        string
		mapping     string
		wantErr     bool
		errContains string
	}{
		{
			"empty mapping",
			"",
			false,
			"",
		},
		{
			"valid mapping",
			"123456789012=prod,123456789013=dev",
			false,
			"",
		},
		{
			"missing equals sign",
			"12345prod",
			true,
			"invalid nickname mapping format",
		},
		{
			"empty account ID",
			"=prod",
			true,
			"empty account ID",
		},
		{
			"empty nickname",
			"123456789012=",
			true,
			"empty nickname",
		},
		{
			"too many equals signs",
			"12345=prod=test",
			true,
			"invalid nickname mapping format",
		},
		{
			"mixed valid and invalid",
			"123456789012=prod,invalid",
			true,
			"invalid nickname mapping format",
		},
		{
			"too short account id",
			"123=prod,1234567789012=stg",
			true,
			"invalid account ID in mapping entry",
		},
		{
			"too long account id",
			"1234567890123=prod,1234567789012=stg",
			true,
			"invalid account ID in mapping entry",
		},
		{
			"alphanumeric account id",
			"123a45789012=prod,1234567789012=stg",
			true,
			"invalid account ID in mapping entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNicknameMapping(tt.mapping)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNicknameMapping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error message %q should contain %q", err.Error(), tt.errContains)
				}
			}

			if !tt.wantErr {
				// Verify the mapping was correctly parsed
				if tt.mapping == "" && len(result) != 0 {
					t.Errorf("Expected empty map for empty input, got %v", result)
				}

				if tt.mapping != "" {
					expected := make(map[string]string)
					for _, pair := range strings.Split(tt.mapping, ",") {
						if strings.Contains(pair, "=") {
							parts := strings.Split(pair, "=")
							expected[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
						}
					}

					if !reflect.DeepEqual(result, expected) {
						t.Errorf("Expected %v, got %v", expected, result)
					}
				}
			}
		})
	}
}
