package setlist

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgTypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

type mockOrganizationsClient struct {
	ListAccountsFunc func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error)
}

func (m *mockOrganizationsClient) ListAccounts(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
	return m.ListAccountsFunc(ctx, params, optFns...)
}

func TestListAccounts(t *testing.T) {
	tests := []struct {
		name          string
		client        *mockOrganizationsClient
		expectedCount int
		expectError   bool
	}{
		{
			name: "single page of results",
			client: &mockOrganizationsClient{
				ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
					return &organizations.ListAccountsOutput{
						Accounts: []orgTypes.Account{
							{Id: aws.String("111111111111"), Name: aws.String("Account 1")},
							{Id: aws.String("222222222222"), Name: aws.String("Account 2")},
						},
					}, nil
				},
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "multiple pages of results",
			client: &mockOrganizationsClient{
				ListAccountsFunc: func() func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
					callCount := 0
					return func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
						callCount++
						if callCount == 1 {
							return &organizations.ListAccountsOutput{
								Accounts: []orgTypes.Account{
									{Id: aws.String("111111111111"), Name: aws.String("Account 1")},
								},
								NextToken: aws.String("page2"),
							}, nil
						}
						return &organizations.ListAccountsOutput{
							Accounts: []orgTypes.Account{
								{Id: aws.String("222222222222"), Name: aws.String("Account 2")},
							},
						}, nil
					}
				}(),
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "API error on first call",
			client: &mockOrganizationsClient{
				ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
					return nil, fmt.Errorf("access denied")
				},
			},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name: "API error on second page",
			client: &mockOrganizationsClient{
				ListAccountsFunc: func() func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
					callCount := 0
					return func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
						callCount++
						if callCount == 1 {
							return &organizations.ListAccountsOutput{
								Accounts: []orgTypes.Account{
									{Id: aws.String("111111111111"), Name: aws.String("Account 1")},
								},
								NextToken: aws.String("page2"),
							}, nil
						}
						return nil, fmt.Errorf("throttled")
					}
				}(),
			},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name: "empty results",
			client: &mockOrganizationsClient{
				ListAccountsFunc: func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
					return &organizations.ListAccountsOutput{
						Accounts: []orgTypes.Account{},
					}, nil
				},
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, err := ListAccounts(context.Background(), tt.client)

			if (err != nil) != tt.expectError {
				t.Errorf("ListAccounts() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && len(accounts) != tt.expectedCount {
				t.Errorf("Expected %d accounts, got %d", tt.expectedCount, len(accounts))
			}
		})
	}
}

// Mock SSO Admin client that waits before responding
type delaySSOAdminClient struct {
	mockSSOAdminClient
	delay time.Duration
}

func (m *delaySSOAdminClient) ListInstances(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error) {
	select {
	case <-time.After(m.delay):
		return &ssoadmin.ListInstancesOutput{
			Instances: []types.InstanceMetadata{
				{
					InstanceArn:     aws.String("test-arn"),
					IdentityStoreId: aws.String("test-store"),
				},
			},
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *delaySSOAdminClient) ListPermissionSetsProvisionedToAccount(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
	select {
	case <-time.After(m.delay):
		return &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
			PermissionSets: []string{"test-permission-set"},
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *delaySSOAdminClient) DescribePermissionSet(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error) {
	select {
	case <-time.After(m.delay):
		return &ssoadmin.DescribePermissionSetOutput{
			PermissionSet: &types.PermissionSet{
				Name:             aws.String("TestPermSet"),
				Description:      aws.String("Test Permission Set"),
				SessionDuration:  aws.String("PT1H"),
				PermissionSetArn: aws.String("test-permission-set-arn"),
			},
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Mock Organizations client that waits before responding
type delayOrganizationsClient struct {
	delay time.Duration
}

func (m *delayOrganizationsClient) ListAccounts(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
	select {
	case <-time.After(m.delay):
		return &organizations.ListAccountsOutput{
			Accounts: []orgTypes.Account{
				{
					Id:     aws.String("123456789012"),
					Name:   aws.String("Test Account"),
					Status: orgTypes.AccountStatusActive,
				},
			},
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestContextCancellation(t *testing.T) {
	// Define delay that's longer than our context timeout
	delay := 200 * time.Millisecond

	tests := []struct {
		name             string
		contextTimeout   time.Duration
		expectedErrorMsg string
		testFunc         func(context.Context, testing.TB) error
	}{
		{
			name:             "SsoInstance respects cancellation",
			contextTimeout:   50 * time.Millisecond, // Less than the delay
			expectedErrorMsg: "context deadline exceeded",
			testFunc: func(ctx context.Context, t testing.TB) error {
				client := &delaySSOAdminClient{delay: delay}
				_, err := SsoInstance(ctx, client)
				return err
			},
		},
		{
			name:             "ListAccounts respects cancellation",
			contextTimeout:   50 * time.Millisecond,
			expectedErrorMsg: "context deadline exceeded",
			testFunc: func(ctx context.Context, t testing.TB) error {
				client := &delayOrganizationsClient{delay: delay}
				_, err := ListAccounts(ctx, client)
				return err
			},
		},
		{
			name:             "PermissionSets list respects cancellation",
			contextTimeout:   50 * time.Millisecond,
			expectedErrorMsg: "context deadline exceeded",
			testFunc: func(ctx context.Context, t testing.TB) error {
				client := &delaySSOAdminClient{delay: delay}
				_, err := PermissionSets(ctx, client, "test-instance-arn", "123456789012")
				return err
			},
		},
		{
			name:             "Context without timeout completes successfully",
			contextTimeout:   500 * time.Millisecond, // Longer than the delay
			expectedErrorMsg: "",                     // No error expected
			testFunc: func(ctx context.Context, t testing.TB) error {
				client := &delaySSOAdminClient{delay: delay}
				_, err := SsoInstance(ctx, client)
				return err
			},
		},
		{
			name:             "Already cancelled context fails immediately",
			contextTimeout:   0, // Not used - we'll manually cancel
			expectedErrorMsg: "context canceled",
			testFunc: func(ctx context.Context, t testing.TB) error {
				// Create and immediately cancel a context
				cancelCtx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				client := &delaySSOAdminClient{delay: delay}
				_, err := SsoInstance(cancelCtx, client)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			var cancel context.CancelFunc

			if tt.contextTimeout > 0 {
				ctx, cancel = context.WithTimeout(context.Background(), tt.contextTimeout)
				defer cancel()
			} else {
				ctx = context.Background()
			}

			err := tt.testFunc(ctx, t)

			if tt.expectedErrorMsg == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing %q but got nil", tt.expectedErrorMsg)
					return
				}

				if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
					t.Errorf("Expected context error, got: %v", err)
				}

				if err.Error() != tt.expectedErrorMsg && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
					t.Errorf("Expected error message %q, got %q", tt.expectedErrorMsg, err.Error())
				}
			}
		})
	}
}
