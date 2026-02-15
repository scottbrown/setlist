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
	"github.com/aws/aws-sdk-go-v2/service/organizations"
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
	ListPermissionSetsFunc                     func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error)
	ListPermissionSetsProvisionedToAccountFunc func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error)
	DescribePermissionSetFunc                  func(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error)
}

func (m *mockSSOAdminClient) ListInstances(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error) {
	if m.ListInstancesFunc != nil {
		return m.ListInstancesFunc(ctx, params, optFns...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSSOAdminClient) ListPermissionSets(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
	if m.ListPermissionSetsFunc != nil {
		return m.ListPermissionSetsFunc(ctx, params, optFns...)
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

	configFile := setlist.ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-1234567890",
		Region:          "us-east-1",
		Profiles:        []setlist.Profile{},
	}

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

	configFile := setlist.ConfigFile{}

	err := outputConfig(configFile)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

func TestHandleListPermissions(t *testing.T) {
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
		"sso:ListPermissionSets",
		"sso:ListPermissionSetsProvisionedToAccount",
		"sso:DescribePermissionSet",
	}

	for _, perm := range expected {
		if !strings.Contains(output, perm) {
			t.Errorf("Expected output to contain %q, got: %s", perm, output)
		}
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
	origListPermissionSets := listPermissionSets
	defer func() {
		permissions = origPermissions
		checkUpdate = origCheckUpdate
		listAccounts = origListAccounts
		listPermissionSets = origListPermissionSets
	}()

	permissions = true
	checkUpdate = false
	listAccounts = false
	listPermissionSets = false

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

type mockOrganizationsClient struct {
	ListAccountsFunc func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error)
}

func (m *mockOrganizationsClient) ListAccounts(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
	return m.ListAccountsFunc(ctx, params, optFns...)
}

func TestHandleListAccountsFlow(t *testing.T) {
	origIncludeAccounts := includeAccounts
	origExcludeAccounts := excludeAccounts
	defer func() {
		includeAccounts = origIncludeAccounts
		excludeAccounts = origExcludeAccounts
	}()
	includeAccounts = ""
	excludeAccounts = ""

	tests := []struct {
		name        string
		client      *mockOrganizationsClient
		expectError bool
		expected    []string
	}{
		{
			name: "lists accounts successfully",
			client: &mockOrganizationsClient{
				ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
					return &organizations.ListAccountsOutput{
						Accounts: []orgtypes.Account{
							{Id: aws.String("111111111111"), Name: aws.String("Dev")},
							{Id: aws.String("222222222222"), Name: aws.String("Prod")},
						},
					}, nil
				},
			},
			expectError: false,
			expected:    []string{"111111111111", "Dev", "222222222222", "Prod"},
		},
		{
			name: "API error",
			client: &mockOrganizationsClient{
				ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
					return nil, errors.New("access denied")
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := handleListAccountsFlow(context.Background(), tt.client)
			w.Close()
			os.Stdout = oldStdout

			if (err != nil) != tt.expectError {
				t.Errorf("handleListAccountsFlow() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				var buf bytes.Buffer
				buf.ReadFrom(r)
				output := buf.String()

				for _, exp := range tt.expected {
					if !strings.Contains(output, exp) {
						t.Errorf("Expected output to contain %q, got: %s", exp, output)
					}
				}
			}
		})
	}
}

func TestHandleListPermissionSetsFlow(t *testing.T) {
	tests := []struct {
		name        string
		client      *mockSSOAdminClient
		expectError bool
		expected    []string
	}{
		{
			name: "lists permission sets successfully",
			client: &mockSSOAdminClient{
				ListInstancesFunc: func(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error) {
					return &ssoadmin.ListInstancesOutput{
						Instances: []ssotypes.InstanceMetadata{
							{
								InstanceArn:     aws.String("arn:aws:sso:::instance/ssoins-12345678"),
								IdentityStoreId: aws.String("d-1234567890"),
							},
						},
					}, nil
				},
				ListPermissionSetsFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
					return &ssoadmin.ListPermissionSetsOutput{
						PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-123"},
					}, nil
				},
				DescribePermissionSetFunc: func(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error) {
					return &ssoadmin.DescribePermissionSetOutput{
						PermissionSet: &ssotypes.PermissionSet{
							Name:        aws.String("ViewOnly"),
							Description: aws.String("View only access"),
						},
					}, nil
				},
			},
			expectError: false,
			expected:    []string{"ViewOnly", "View only access"},
		},
		{
			name: "SSO instance error",
			client: &mockSSOAdminClient{
				ListInstancesFunc: func(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error) {
					return nil, errors.New("SSO not configured")
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := handleListPermissionSetsFlow(context.Background(), tt.client)
			w.Close()
			os.Stdout = oldStdout

			if (err != nil) != tt.expectError {
				t.Errorf("handleListPermissionSetsFlow() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				var buf bytes.Buffer
				buf.ReadFrom(r)
				output := buf.String()

				for _, exp := range tt.expected {
					if !strings.Contains(output, exp) {
						t.Errorf("Expected output to contain %q, got: %s", exp, output)
					}
				}
			}
		})
	}
}

func TestHandleListPermissionSets(t *testing.T) {
	tests := []struct {
		name        string
		client      *mockSSOAdminClient
		expectError bool
		expected    []string
	}{
		{
			name: "lists permission sets",
			client: &mockSSOAdminClient{
				ListPermissionSetsFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
					return &ssoadmin.ListPermissionSetsOutput{
						PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-123"},
					}, nil
				},
				DescribePermissionSetFunc: func(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error) {
					return &ssoadmin.DescribePermissionSetOutput{
						PermissionSet: &ssotypes.PermissionSet{
							Name:        aws.String("AdminAccess"),
							Description: aws.String("Full admin access"),
						},
					}, nil
				},
			},
			expectError: false,
			expected:    []string{"AdminAccess", "Full admin access"},
		},
		{
			name: "API error",
			client: &mockSSOAdminClient{
				ListPermissionSetsFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
					return nil, errors.New("access denied")
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := ssotypes.InstanceMetadata{
				InstanceArn:     aws.String("arn:aws:sso:::instance/ssoins-12345678"),
				IdentityStoreId: aws.String("d-1234567890"),
			}

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := handleListPermissionSets(context.Background(), tt.client, instance)
			w.Close()
			os.Stdout = oldStdout

			if (err != nil) != tt.expectError {
				t.Errorf("handleListPermissionSets() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				var buf bytes.Buffer
				buf.ReadFrom(r)
				output := buf.String()

				for _, exp := range tt.expected {
					if !strings.Contains(output, exp) {
						t.Errorf("Expected output to contain %q, got: %s", exp, output)
					}
				}
			}
		})
	}
}
