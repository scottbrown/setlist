package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/scottbrown/setlist"

	"github.com/aws/aws-sdk-go-v2/aws"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

type mockSSOAdminClient struct {
	ListInstancesFunc                          func(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error)
	ListPermissionSetsProvisionedToAccountFunc func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error)
	DescribePermissionSetFunc                  func(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error)
}

func (m *mockSSOAdminClient) ListInstances(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error) {
	if m.ListInstancesFunc != nil {
		return m.ListInstancesFunc(ctx, params, optFns...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSSOAdminClient) ListPermissionSetsProvisionedToAccount(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
	if m.ListPermissionSetsProvisionedToAccountFunc != nil {
		return m.ListPermissionSetsProvisionedToAccountFunc(ctx, params, optFns...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSSOAdminClient) DescribePermissionSet(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error) {
	if m.DescribePermissionSetFunc != nil {
		return m.DescribePermissionSetFunc(ctx, params, optFns...)
	}
	return nil, errors.New("not implemented")
}

func TestBuildAccountProfiles(t *testing.T) {
	origSession := ssoSession
	defer func() { ssoSession = origSession }()

	tests := []struct {
		name           string
		ssoSessionVal  string
		account        orgtypes.Account
		permissionSets []ssotypes.PermissionSet
		expectedCount  int
	}{
		{
			name:          "valid account with complete permission sets",
			ssoSessionVal: "test-session",
			account:       orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            aws.String("AdminAccess"),
					Description:     aws.String("Full admin access"),
					SessionDuration: aws.String("PT1H"),
				},
				{
					Name:            aws.String("ReadOnly"),
					Description:     aws.String("Read only access"),
					SessionDuration: aws.String("PT2H"),
				},
			},
			expectedCount: 2,
		},
		{
			name:          "permission set with nil Name is skipped",
			ssoSessionVal: "test-session",
			account:       orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            nil,
					Description:     aws.String("Some description"),
					SessionDuration: aws.String("PT1H"),
				},
			},
			expectedCount: 0,
		},
		{
			name:          "permission set with nil Description is skipped",
			ssoSessionVal: "test-session",
			account:       orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            aws.String("AdminAccess"),
					Description:     nil,
					SessionDuration: aws.String("PT1H"),
				},
			},
			expectedCount: 0,
		},
		{
			name:          "permission set with nil SessionDuration is skipped",
			ssoSessionVal: "test-session",
			account:       orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            aws.String("AdminAccess"),
					Description:     aws.String("Full admin access"),
					SessionDuration: nil,
				},
			},
			expectedCount: 0,
		},
		{
			name:           "empty permission sets list",
			ssoSessionVal:  "test-session",
			account:        orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{},
			expectedCount:  0,
		},
		{
			name:          "mix of valid and incomplete permission sets",
			ssoSessionVal: "test-session",
			account:       orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            aws.String("ValidSet"),
					Description:     aws.String("Valid"),
					SessionDuration: aws.String("PT1H"),
				},
				{
					Name:            nil,
					Description:     aws.String("Missing name"),
					SessionDuration: aws.String("PT1H"),
				},
			},
			expectedCount: 1,
		},
		{
			name:          "empty description string triggers error path",
			ssoSessionVal: "test-session",
			account:       orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            aws.String("AdminAccess"),
					Description:     aws.String(""),
					SessionDuration: aws.String("PT1H"),
				},
			},
			expectedCount: 0,
		},
		{
			name:          "empty session duration string triggers error path",
			ssoSessionVal: "test-session",
			account:       orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            aws.String("AdminAccess"),
					Description:     aws.String("Valid desc"),
					SessionDuration: aws.String(""),
				},
			},
			expectedCount: 0,
		},
		{
			name:          "empty ssoSession triggers session name error path",
			ssoSessionVal: "",
			account:       orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            aws.String("AdminAccess"),
					Description:     aws.String("Valid desc"),
					SessionDuration: aws.String("PT1H"),
				},
			},
			expectedCount: 0,
		},
		{
			name:          "invalid account ID length triggers error path",
			ssoSessionVal: "test-session",
			account:       orgtypes.Account{Id: aws.String("short"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            aws.String("AdminAccess"),
					Description:     aws.String("Valid desc"),
					SessionDuration: aws.String("PT1H"),
				},
			},
			expectedCount: 0,
		},
		{
			name:          "empty role name triggers error path",
			ssoSessionVal: "test-session",
			account:       orgtypes.Account{Id: aws.String("123456789012"), Name: aws.String("Test Account")},
			permissionSets: []ssotypes.PermissionSet{
				{
					Name:            aws.String(""),
					Description:     aws.String("Valid desc"),
					SessionDuration: aws.String("PT1H"),
				},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssoSession = tt.ssoSessionVal
			profiles, err := buildAccountProfiles(tt.account, tt.permissionSets)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(profiles) != tt.expectedCount {
				t.Errorf("Expected %d profiles, got %d", tt.expectedCount, len(profiles))
			}
		})
	}
}

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
			// Capture stdout
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
	// Save and restore globals
	origStdout := stdout
	defer func() { stdout = origStdout }()
	stdout = true

	configFile := setlist.ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-1234567890",
		Region:          "us-east-1",
		Profiles:        []setlist.Profile{},
	}

	// Capture stdout
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputConfig(configFile)
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

	configFile := setlist.ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-1234567890",
		Region:          "us-east-1",
		Profiles:        []setlist.Profile{},
	}

	err := outputConfig(configFile)
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

	// Missing required fields
	configFile := setlist.ConfigFile{}

	err := outputConfig(configFile)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

func TestHandleListPermissions(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := handleListPermissions()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := []string{
		"organizations:ListAccounts",
		"sso:ListInstances",
		"sso:ListPermissionSetsProvisionedToAccount",
		"sso:DescribePermissionSet",
	}

	for _, perm := range expected {
		if !strings.Contains(output, perm) {
			t.Errorf("Expected output to contain %q, got: %s", perm, output)
		}
	}
}

func TestBuildConfigFile(t *testing.T) {
	origSession := ssoSession
	origRegion := ssoRegion
	origFriendlyName := ssoFriendlyName
	defer func() {
		ssoSession = origSession
		ssoRegion = origRegion
		ssoFriendlyName = origFriendlyName
	}()

	tests := []struct {
		name        string
		ssoSession  string
		ssoRegion   string
		instance    ssotypes.InstanceMetadata
		accounts    []orgtypes.Account
		nicknames   map[string]string
		client      *mockSSOAdminClient
		expectError bool
	}{
		{
			name:       "valid config with one account and permission set",
			ssoSession: "my-session",
			ssoRegion:  "us-east-1",
			instance: ssotypes.InstanceMetadata{
				InstanceArn:     aws.String("arn:aws:sso:::instance/ssoins-12345678"),
				IdentityStoreId: aws.String("d-1234567890"),
			},
			accounts: []orgtypes.Account{
				{Id: aws.String("123456789012"), Name: aws.String("TestAccount")},
			},
			nicknames: map[string]string{},
			client: &mockSSOAdminClient{
				ListPermissionSetsProvisionedToAccountFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
					return &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
						PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-12345678"},
					}, nil
				},
				DescribePermissionSetFunc: func(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error) {
					return &ssoadmin.DescribePermissionSetOutput{
						PermissionSet: &ssotypes.PermissionSet{
							Name:            aws.String("AdminRole"),
							Description:     aws.String("Admin access"),
							SessionDuration: aws.String("PT1H"),
						},
					}, nil
				},
			},
			expectError: false,
		},
		{
			name:       "invalid identity store ID",
			ssoSession: "my-session",
			ssoRegion:  "us-east-1",
			instance: ssotypes.InstanceMetadata{
				InstanceArn:     aws.String("arn:aws:sso:::instance/ssoins-12345678"),
				IdentityStoreId: aws.String("invalid"),
			},
			accounts:    []orgtypes.Account{},
			nicknames:   map[string]string{},
			client:      &mockSSOAdminClient{},
			expectError: true,
		},
		{
			name:       "empty region",
			ssoSession: "my-session",
			ssoRegion:  "",
			instance: ssotypes.InstanceMetadata{
				InstanceArn:     aws.String("arn:aws:sso:::instance/ssoins-12345678"),
				IdentityStoreId: aws.String("d-1234567890"),
			},
			accounts:    []orgtypes.Account{},
			nicknames:   map[string]string{},
			client:      &mockSSOAdminClient{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssoSession = tt.ssoSession
			ssoRegion = tt.ssoRegion
			ssoFriendlyName = ""

			configFile, err := buildConfigFile(context.Background(), tt.client, tt.instance, tt.accounts, tt.nicknames, nil, nil)

			if (err != nil) != tt.expectError {
				t.Errorf("buildConfigFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && configFile.SessionName != tt.ssoSession {
				t.Errorf("Expected session name %q, got %q", tt.ssoSession, configFile.SessionName)
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

func TestHandleRootPermissionsFlag(t *testing.T) {
	origPermissions := permissions
	origCheckUpdate := checkUpdate
	origListAccounts := listAccounts
	defer func() {
		permissions = origPermissions
		checkUpdate = origCheckUpdate
		listAccounts = origListAccounts
	}()

	permissions = true
	checkUpdate = false
	listAccounts = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := handleRoot(rootCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "organizations:ListAccounts") {
		t.Errorf("Expected permissions output, got: %s", output)
	}
}

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

			// Capture stdout
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

func TestBuildProfiles(t *testing.T) {
	origSession := ssoSession
	defer func() { ssoSession = origSession }()
	ssoSession = "test-session"

	tests := []struct {
		name          string
		accounts      []orgtypes.Account
		client        *mockSSOAdminClient
		expectedCount int
		expectError   bool
	}{
		{
			name: "multiple accounts with permission sets",
			accounts: []orgtypes.Account{
				{Id: aws.String("111111111111"), Name: aws.String("Account1")},
				{Id: aws.String("222222222222"), Name: aws.String("Account2")},
			},
			client: &mockSSOAdminClient{
				ListPermissionSetsProvisionedToAccountFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
					return &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
						PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-123"},
					}, nil
				},
				DescribePermissionSetFunc: func(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error) {
					return &ssoadmin.DescribePermissionSetOutput{
						PermissionSet: &ssotypes.PermissionSet{
							Name:            aws.String("TestRole"),
							Description:     aws.String("Test role"),
							SessionDuration: aws.String("PT1H"),
						},
					}, nil
				},
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "account with nil ID is skipped",
			accounts: []orgtypes.Account{
				{Id: nil, Name: aws.String("No ID Account")},
				{Id: aws.String("111111111111"), Name: aws.String("Valid Account")},
			},
			client: &mockSSOAdminClient{
				ListPermissionSetsProvisionedToAccountFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
					return &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
						PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-123"},
					}, nil
				},
				DescribePermissionSetFunc: func(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error) {
					return &ssoadmin.DescribePermissionSetOutput{
						PermissionSet: &ssotypes.PermissionSet{
							Name:            aws.String("TestRole"),
							Description:     aws.String("Test role"),
							SessionDuration: aws.String("PT1H"),
						},
					}, nil
				},
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "permission sets API error",
			accounts: []orgtypes.Account{
				{Id: aws.String("111111111111"), Name: aws.String("Account1")},
			},
			client: &mockSSOAdminClient{
				ListPermissionSetsProvisionedToAccountFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
					return nil, errors.New("API access denied")
				},
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := ssotypes.InstanceMetadata{
				InstanceArn:     aws.String("arn:aws:sso:::instance/ssoins-12345678"),
				IdentityStoreId: aws.String("d-1234567890"),
			}

			profiles, err := buildProfiles(context.Background(), tt.client, instance, tt.accounts, nil, nil)

			if (err != nil) != tt.expectError {
				t.Errorf("buildProfiles() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && len(profiles) != tt.expectedCount {
				t.Errorf("Expected %d profiles, got %d", tt.expectedCount, len(profiles))
			}
		})
	}
}
