package setlist

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

func TestParseAccountIdList(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []AWSAccountId
		wantErr     bool
		errContains string
	}{
		{
			name:     "single valid ID",
			input:    "123456789012",
			expected: []AWSAccountId{"123456789012"},
		},
		{
			name:     "multiple valid IDs",
			input:    "123456789012,234567890123,345678901234",
			expected: []AWSAccountId{"123456789012", "234567890123", "345678901234"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "trailing comma",
			input:    "123456789012,",
			expected: []AWSAccountId{"123456789012"},
		},
		{
			name:        "too short ID",
			input:       "12345",
			wantErr:     true,
			errContains: "invalid account ID",
		},
		{
			name:        "non-numeric ID",
			input:       "12345678901a",
			wantErr:     true,
			errContains: "invalid account ID",
		},
		{
			name:        "too long ID",
			input:       "1234567890123",
			wantErr:     true,
			errContains: "invalid account ID",
		},
		{
			name:     "whitespace around IDs",
			input:    " 123456789012 , 234567890123 ",
			expected: []AWSAccountId{"123456789012", "234567890123"},
		},
		{
			name:        "one valid one invalid",
			input:       "123456789012,bad",
			wantErr:     true,
			errContains: "invalid account ID at position 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAccountIdList(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAccountIdList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error message %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d IDs, got %d", len(tt.expected), len(result))
				return
			}

			for i, id := range result {
				if id != tt.expected[i] {
					t.Errorf("ID at position %d: expected %q, got %q", i, tt.expected[i], id)
				}
			}
		})
	}
}

func TestFilterAccounts(t *testing.T) {
	accounts := []orgtypes.Account{
		{Id: aws.String("111111111111"), Name: aws.String("Account One")},
		{Id: aws.String("222222222222"), Name: aws.String("Account Two")},
		{Id: aws.String("333333333333"), Name: aws.String("Account Three")},
	}

	tests := []struct {
		name        string
		accounts    []orgtypes.Account
		include     []AWSAccountId
		exclude     []AWSAccountId
		expectedIDs []string
		wantErr     bool
	}{
		{
			name:        "no filters returns all accounts",
			accounts:    accounts,
			include:     nil,
			exclude:     nil,
			expectedIDs: []string{"111111111111", "222222222222", "333333333333"},
		},
		{
			name:        "include filter",
			accounts:    accounts,
			include:     []AWSAccountId{"111111111111", "333333333333"},
			exclude:     nil,
			expectedIDs: []string{"111111111111", "333333333333"},
		},
		{
			name:        "exclude filter",
			accounts:    accounts,
			include:     nil,
			exclude:     []AWSAccountId{"222222222222"},
			expectedIDs: []string{"111111111111", "333333333333"},
		},
		{
			name:     "both set returns error",
			accounts: accounts,
			include:  []AWSAccountId{"111111111111"},
			exclude:  []AWSAccountId{"222222222222"},
			wantErr:  true,
		},
		{
			name:        "include with no matching accounts",
			accounts:    accounts,
			include:     []AWSAccountId{"999999999999"},
			exclude:     nil,
			expectedIDs: nil,
		},
		{
			name:        "exclude with no matching accounts",
			accounts:    accounts,
			include:     nil,
			exclude:     []AWSAccountId{"999999999999"},
			expectedIDs: []string{"111111111111", "222222222222", "333333333333"},
		},
		{
			name:        "empty account list",
			accounts:    []orgtypes.Account{},
			include:     []AWSAccountId{"111111111111"},
			exclude:     nil,
			expectedIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterAccounts(tt.accounts, tt.include, tt.exclude)

			if (err != nil) != tt.wantErr {
				t.Errorf("FilterAccounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(result) != len(tt.expectedIDs) {
				t.Errorf("Expected %d accounts, got %d", len(tt.expectedIDs), len(result))
				return
			}

			for i, a := range result {
				if *a.Id != tt.expectedIDs[i] {
					t.Errorf("Account at position %d: expected ID %q, got %q", i, tt.expectedIDs[i], *a.Id)
				}
			}
		})
	}
}

func TestParsePermissionSetList(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []string
		wantErr     bool
		errContains string
	}{
		{
			name:     "single name",
			input:    "AdminAccess",
			expected: []string{"AdminAccess"},
		},
		{
			name:     "multiple names",
			input:    "AdminAccess,ReadOnly,PowerUser",
			expected: []string{"AdminAccess", "ReadOnly", "PowerUser"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "trailing comma",
			input:    "AdminAccess,",
			expected: []string{"AdminAccess"},
		},
		{
			name:     "whitespace around names",
			input:    " AdminAccess , ReadOnly ",
			expected: []string{"AdminAccess", "ReadOnly"},
		},
		{
			name:        "name with internal whitespace",
			input:       "Admin Access",
			wantErr:     true,
			errContains: "invalid permission set name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePermissionSetList(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePermissionSetList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error message %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d names, got %d", len(tt.expected), len(result))
				return
			}

			for i, name := range result {
				if name != tt.expected[i] {
					t.Errorf("Name at position %d: expected %q, got %q", i, tt.expected[i], name)
				}
			}
		})
	}
}

func TestFilterPermissionSets(t *testing.T) {
	permissionSets := []ssotypes.PermissionSet{
		{Name: aws.String("AdminAccess"), Description: aws.String("Admin"), SessionDuration: aws.String("PT1H")},
		{Name: aws.String("ReadOnly"), Description: aws.String("Read"), SessionDuration: aws.String("PT2H")},
		{Name: aws.String("PowerUser"), Description: aws.String("Power"), SessionDuration: aws.String("PT1H")},
	}

	tests := []struct {
		name          string
		permSets      []ssotypes.PermissionSet
		include       []string
		exclude       []string
		expectedNames []string
		wantErr       bool
	}{
		{
			name:          "no filters returns all",
			permSets:      permissionSets,
			include:       nil,
			exclude:       nil,
			expectedNames: []string{"AdminAccess", "ReadOnly", "PowerUser"},
		},
		{
			name:          "include filter",
			permSets:      permissionSets,
			include:       []string{"AdminAccess", "PowerUser"},
			exclude:       nil,
			expectedNames: []string{"AdminAccess", "PowerUser"},
		},
		{
			name:          "exclude filter",
			permSets:      permissionSets,
			include:       nil,
			exclude:       []string{"ReadOnly"},
			expectedNames: []string{"AdminAccess", "PowerUser"},
		},
		{
			name:     "both set returns error",
			permSets: permissionSets,
			include:  []string{"AdminAccess"},
			exclude:  []string{"ReadOnly"},
			wantErr:  true,
		},
		{
			name:          "include with no matches",
			permSets:      permissionSets,
			include:       []string{"NonExistent"},
			exclude:       nil,
			expectedNames: nil,
		},
		{
			name:          "exclude with no matches",
			permSets:      permissionSets,
			include:       nil,
			exclude:       []string{"NonExistent"},
			expectedNames: []string{"AdminAccess", "ReadOnly", "PowerUser"},
		},
		{
			name:          "empty permission sets list",
			permSets:      []ssotypes.PermissionSet{},
			include:       []string{"AdminAccess"},
			exclude:       nil,
			expectedNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterPermissionSets(tt.permSets, tt.include, tt.exclude)

			if (err != nil) != tt.wantErr {
				t.Errorf("FilterPermissionSets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(result) != len(tt.expectedNames) {
				t.Errorf("Expected %d permission sets, got %d", len(tt.expectedNames), len(result))
				return
			}

			for i, p := range result {
				if *p.Name != tt.expectedNames[i] {
					t.Errorf("Permission set at position %d: expected name %q, got %q", i, tt.expectedNames[i], *p.Name)
				}
			}
		})
	}
}
