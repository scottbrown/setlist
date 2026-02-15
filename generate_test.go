package setlist

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

type mockOrgClient struct {
	ListAccountsFunc func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error)
}

func (m *mockOrgClient) ListAccounts(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
	return m.ListAccountsFunc(ctx, params, optFns...)
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		input       GenerateInput
		expectError bool
		errContains string
		checkResult func(t *testing.T, cf ConfigFile)
	}{
		{
			name: "successful generation",
			input: GenerateInput{
				SSOClient: &mockSSOAdminClient{
					ListPermissionSetsProvisionedToAccountFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
						return &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
							PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-123"},
						}, nil
					},
					describePermSetOutput: &ssoadmin.DescribePermissionSetOutput{
						PermissionSet: &types.PermissionSet{
							Name:            aws.String("AdminAccess"),
							Description:     aws.String("Full admin"),
							SessionDuration: aws.String("PT1H"),
						},
					},
				},
				OrgClient: &mockOrgClient{
					ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
						return &organizations.ListAccountsOutput{
							Accounts: []orgtypes.Account{
								{Id: aws.String("123456789012"), Name: aws.String("TestAccount")},
							},
						}, nil
					},
				},
				SessionName: "my-session",
				Region:      "us-east-1",
			},
			checkResult: func(t *testing.T, cf ConfigFile) {
				if cf.SessionName != "my-session" {
					t.Errorf("Expected session name 'my-session', got %q", cf.SessionName)
				}
				if len(cf.Profiles) != 1 {
					t.Errorf("Expected 1 profile, got %d", len(cf.Profiles))
				}
			},
		},
		{
			name: "SSO instance error",
			input: GenerateInput{
				SSOClient: &mockSSOAdminClient{},
				OrgClient: &mockOrgClient{
					ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
						return &organizations.ListAccountsOutput{}, nil
					},
				},
				SessionName: "test",
				Region:      "us-east-1",
			},
			expectError: true,
			errContains: "SSO instance",
		},
		{
			name: "list accounts error",
			input: GenerateInput{
				SSOClient: &mockSSOAdminClient{
					ListPermissionSetsProvisionedToAccountFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
						return &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{}, nil
					},
				},
				OrgClient: &mockOrgClient{
					ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
						return nil, errors.New("org access denied")
					},
				},
				SessionName: "test",
				Region:      "us-east-1",
			},
			expectError: true,
			errContains: "failed to list AWS accounts",
		},
		{
			name: "with nickname mapping",
			input: GenerateInput{
				SSOClient: &mockSSOAdminClient{
					ListPermissionSetsProvisionedToAccountFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
						return &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
							PermissionSets: []string{"arn:1"},
						}, nil
					},
					describePermSetOutput: &ssoadmin.DescribePermissionSetOutput{
						PermissionSet: &types.PermissionSet{
							Name:            aws.String("ReadOnly"),
							Description:     aws.String("Read only"),
							SessionDuration: aws.String("PT2H"),
						},
					},
				},
				OrgClient: &mockOrgClient{
					ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
						return &organizations.ListAccountsOutput{
							Accounts: []orgtypes.Account{
								{Id: aws.String("123456789012"), Name: aws.String("Prod")},
							},
						}, nil
					},
				},
				SessionName:     "my-org",
				Region:          "us-east-1",
				NicknameMapping: "123456789012=prod",
			},
			checkResult: func(t *testing.T, cf ConfigFile) {
				if !cf.HasNickname("123456789012") {
					t.Error("Expected nickname mapping for account 123456789012")
				}
			},
		},
		{
			name: "invalid region",
			input: GenerateInput{
				SSOClient: &mockSSOAdminClient{
					ListPermissionSetsProvisionedToAccountFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
						return &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{}, nil
					},
				},
				OrgClient: &mockOrgClient{
					ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
						return &organizations.ListAccountsOutput{}, nil
					},
				},
				SessionName: "test",
				Region:      "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Provide a valid ListInstances for SSO mock when not testing SSO failure
			ssoClient, ok := tt.input.SSOClient.(*mockSSOAdminClient)
			if ok && ssoClient.ListPermissionSetsProvisionedToAccountFunc != nil {
				// Ensure ListInstances returns a valid instance for non-error tests
				if tt.errContains != "SSO instance" {
					// The mock already has ListInstances returning error by default,
					// but we need it to return a valid instance for Generate to work
				}
			}

			// Wrap the SSOClient to provide a valid ListInstances response for non-error cases
			wrappedInput := tt.input
			if tt.errContains != "SSO instance" {
				wrappedSSO, ok := tt.input.SSOClient.(*mockSSOAdminClient)
				if ok {
					// Create a copy and set a valid ListInstances
					newMock := *wrappedSSO
					newMock.ListPermissionSetsFunc = wrappedSSO.ListPermissionSetsFunc
					newMock.ListPermissionSetsProvisionedToAccountFunc = wrappedSSO.ListPermissionSetsProvisionedToAccountFunc
					newMock.describePermSetOutput = wrappedSSO.describePermSetOutput
					newMock.describePermSetError = wrappedSSO.describePermSetError
					wrappedInput.SSOClient = &generateMockSSO{
						inner: &newMock,
					}
				}
			}

			cf, err := Generate(context.Background(), wrappedInput)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, cf)
			}
		})
	}
}

// generateMockSSO wraps a mockSSOAdminClient but provides a valid ListInstances response
type generateMockSSO struct {
	inner *mockSSOAdminClient
}

func (g *generateMockSSO) ListInstances(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error) {
	return &ssoadmin.ListInstancesOutput{
		Instances: []types.InstanceMetadata{
			{
				InstanceArn:     aws.String("arn:aws:sso:::instance/ssoins-12345678"),
				IdentityStoreId: aws.String("d-1234567890"),
			},
		},
	}, nil
}

func (g *generateMockSSO) ListPermissionSets(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
	return g.inner.ListPermissionSets(ctx, params, optFns...)
}

func (g *generateMockSSO) ListPermissionSetsProvisionedToAccount(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
	return g.inner.ListPermissionSetsProvisionedToAccount(ctx, params, optFns...)
}

func (g *generateMockSSO) DescribePermissionSet(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error) {
	return g.inner.DescribePermissionSet(ctx, params, optFns...)
}
