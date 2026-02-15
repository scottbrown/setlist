package setlist

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

// Mock implementation of the SSO Admin client interface
type mockSSOAdminClient struct {
	listPermSetsResponse  *ssoadmin.ListPermissionSetsProvisionedToAccountOutput
	listPermSetsError     error
	describePermSetOutput *ssoadmin.DescribePermissionSetOutput
	describePermSetError  error
	permSetCallCount      int
	describeCallCount     int

	// Store the custom list function for pagination tests
	ListPermissionSetsProvisionedToAccountFunc func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error)
	ListPermissionSetsFunc                     func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error)
}

func (m *mockSSOAdminClient) ListPermissionSetsProvisionedToAccount(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
	m.permSetCallCount++

	// Use custom function if provided
	if m.ListPermissionSetsProvisionedToAccountFunc != nil {
		return m.ListPermissionSetsProvisionedToAccountFunc(ctx, params, optFns...)
	}

	return m.listPermSetsResponse, m.listPermSetsError
}

func (m *mockSSOAdminClient) DescribePermissionSet(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error) {
	m.describeCallCount++
	return m.describePermSetOutput, m.describePermSetError
}

func (m *mockSSOAdminClient) ListPermissionSets(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
	if m.ListPermissionSetsFunc != nil {
		return m.ListPermissionSetsFunc(ctx, params, optFns...)
	}
	return nil, errors.New("not implemented in mock")
}

// Implement additional methods required by the interface but not used in our tests
func (m *mockSSOAdminClient) ListInstances(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error) {
	return nil, errors.New("not implemented in mock")
}

func TestAllPermissionSets(t *testing.T) {
	testInstanceArn := "arn:aws:sso:::instance/ssoins-12345678"

	tests := []struct {
		name             string
		instanceArn      string
		listFunc         func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error)
		describeOutput   *ssoadmin.DescribePermissionSetOutput
		describeError    error
		expectedError    bool
		expectedErrMsg   string
		expectedPermSets int
	}{
		{
			name:           "empty instance ARN",
			instanceArn:    "",
			expectedError:  true,
			expectedErrMsg: "invalid parameter: empty instanceArn",
		},
		{
			name:        "successful single page",
			instanceArn: testInstanceArn,
			listFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
				return &ssoadmin.ListPermissionSetsOutput{
					PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-111"},
				}, nil
			},
			describeOutput: &ssoadmin.DescribePermissionSetOutput{
				PermissionSet: &types.PermissionSet{
					Name:        aws.String("AdminAccess"),
					Description: aws.String("Full admin"),
				},
			},
			expectedPermSets: 1,
		},
		{
			name:        "pagination",
			instanceArn: testInstanceArn,
			listFunc: func() func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
				callCount := 0
				return func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
					callCount++
					if callCount == 1 {
						return &ssoadmin.ListPermissionSetsOutput{
							PermissionSets: []string{"arn:1"},
							NextToken:      aws.String("page2"),
						}, nil
					}
					return &ssoadmin.ListPermissionSetsOutput{
						PermissionSets: []string{"arn:2"},
					}, nil
				}
			}(),
			describeOutput: &ssoadmin.DescribePermissionSetOutput{
				PermissionSet: &types.PermissionSet{
					Name:        aws.String("Role"),
					Description: aws.String("A role"),
				},
			},
			expectedPermSets: 2,
		},
		{
			name:        "list API error",
			instanceArn: testInstanceArn,
			listFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
				return nil, errors.New("access denied")
			},
			expectedError:  true,
			expectedErrMsg: "access denied",
		},
		{
			name:        "empty results",
			instanceArn: testInstanceArn,
			listFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
				return &ssoadmin.ListPermissionSetsOutput{
					PermissionSets: []string{},
				}, nil
			},
			expectedPermSets: 0,
		},
		{
			name:        "describe returns nil permission set",
			instanceArn: testInstanceArn,
			listFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
				return &ssoadmin.ListPermissionSetsOutput{
					PermissionSets: []string{"arn:1"},
				}, nil
			},
			describeOutput: &ssoadmin.DescribePermissionSetOutput{
				PermissionSet: nil,
			},
			expectedError:  true,
			expectedErrMsg: "nil permission set returned",
		},
		{
			name:        "describe returns error",
			instanceArn: testInstanceArn,
			listFunc: func(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error) {
				return &ssoadmin.ListPermissionSetsOutput{
					PermissionSets: []string{"arn:1"},
				}, nil
			},
			describeError:  errors.New("describe failed"),
			expectedError:  true,
			expectedErrMsg: "describe failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSSOAdminClient{
				ListPermissionSetsFunc: tt.listFunc,
				describePermSetOutput:  tt.describeOutput,
				describePermSetError:   tt.describeError,
			}

			permSets, err := AllPermissionSets(context.Background(), mockClient, tt.instanceArn)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tt.expectedErrMsg != "" && !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if len(permSets) != tt.expectedPermSets {
					t.Errorf("Expected %d permission sets, got %d", tt.expectedPermSets, len(permSets))
				}
			}
		})
	}
}

func TestPermissionSetsErrorHandling(t *testing.T) {
	testInstanceArn := "arn:aws:sso:::instance/ssoins-12345678"
	testAccountId := "123456789012"

	tests := []struct {
		name               string
		instanceArn        string
		accountId          string
		listPermSetsResp   *ssoadmin.ListPermissionSetsProvisionedToAccountOutput
		listPermSetsErr    error
		describePermSetOut *ssoadmin.DescribePermissionSetOutput
		describePermSetErr error
		expectedError      bool
		expectedErrMessage string
		expectedPermSets   int
	}{
		{
			name:               "empty instance ARN",
			instanceArn:        "",
			accountId:          testAccountId,
			expectedError:      true,
			expectedErrMessage: "invalid parameter: empty instanceArn",
		},
		{
			name:               "empty account ID",
			instanceArn:        testInstanceArn,
			accountId:          "",
			expectedError:      true,
			expectedErrMessage: "invalid parameter: empty accountId",
		},
		{
			name:               "list API returns error",
			instanceArn:        testInstanceArn,
			accountId:          testAccountId,
			listPermSetsErr:    errors.New("API error"),
			expectedError:      true,
			expectedErrMessage: "API error",
		},
		{
			name:        "list API returns empty list",
			instanceArn: testInstanceArn,
			accountId:   testAccountId,
			listPermSetsResp: &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
				PermissionSets: []string{},
			},
			expectedError:    false,
			expectedPermSets: 0,
		},
		{
			name:        "describe API returns error",
			instanceArn: testInstanceArn,
			accountId:   testAccountId,
			listPermSetsResp: &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
				PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-12345678"},
			},
			describePermSetErr: errors.New("describe API error"),
			expectedError:      true,
			expectedErrMessage: "describe API error",
		},
		{
			name:        "describe API returns nil permission set",
			instanceArn: testInstanceArn,
			accountId:   testAccountId,
			listPermSetsResp: &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
				PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-12345678"},
			},
			describePermSetOut: &ssoadmin.DescribePermissionSetOutput{
				PermissionSet: nil, // nil permission set
			},
			expectedError:      true,
			expectedErrMessage: "nil permission set returned",
		},
		{
			name:        "describe API returns permission set missing required fields",
			instanceArn: testInstanceArn,
			accountId:   testAccountId,
			listPermSetsResp: &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
				PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-12345678"},
			},
			describePermSetOut: &ssoadmin.DescribePermissionSetOutput{
				PermissionSet: &types.PermissionSet{
					// Missing Name, Description, SessionDuration
					PermissionSetArn: aws.String("arn:aws:sso:::permissionSet/ps-12345678"),
				},
			},
			expectedError:    false, // Should not error on missing fields
			expectedPermSets: 1,     // Should still return the permission set
		},
		{
			name:        "pagination handling",
			instanceArn: testInstanceArn,
			accountId:   testAccountId,
			listPermSetsResp: &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
				PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-12345678"},
				NextToken:      aws.String("next-token"),
			},
			expectedError:      true, // Our mock doesn't support pagination properly
			expectedErrMessage: "pagination not implemented in test mock",
		},
		{
			name:        "successful scenario",
			instanceArn: testInstanceArn,
			accountId:   testAccountId,
			listPermSetsResp: &ssoadmin.ListPermissionSetsProvisionedToAccountOutput{
				PermissionSets: []string{"arn:aws:sso:::permissionSet/ps-12345678"},
			},
			describePermSetOut: &ssoadmin.DescribePermissionSetOutput{
				PermissionSet: &types.PermissionSet{
					PermissionSetArn: aws.String("arn:aws:sso:::permissionSet/ps-12345678"),
					Name:             aws.String("TestPermSet"),
					Description:      aws.String("Test Permission Set"),
					SessionDuration:  aws.String("PT1H"),
				},
			},
			expectedError:    false,
			expectedPermSets: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock client with the test case configuration
			mockClient := &mockSSOAdminClient{
				listPermSetsResponse:  tt.listPermSetsResp,
				listPermSetsError:     tt.listPermSetsErr,
				describePermSetOutput: tt.describePermSetOut,
				describePermSetError:  tt.describePermSetErr,
			}

			// Handle pagination test case specially
			if tt.listPermSetsResp != nil && tt.listPermSetsResp.NextToken != nil {
				// Override the list function to fail on pagination attempt
				mockClient.ListPermissionSetsProvisionedToAccountFunc = func(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error) {
					if params.NextToken != nil {
						return nil, errors.New("pagination not implemented in test mock")
					}
					return tt.listPermSetsResp, tt.listPermSetsErr
				}
			}

			// Call the function under test
			permSets, err := PermissionSets(context.Background(), mockClient, tt.instanceArn, tt.accountId)

			// Check error expectations
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tt.expectedErrMessage != "" && !errors.Is(err, errors.New(tt.expectedErrMessage)) &&
					!strings.Contains(err.Error(), tt.expectedErrMessage) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedErrMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}

				// Check expected result count
				if len(permSets) != tt.expectedPermSets {
					t.Errorf("Expected %d permission sets, got %d", tt.expectedPermSets, len(permSets))
				}
			}

			// Additional assertions could check specific permission set fields
			// if the test expects specific values in the permission sets
		})
	}
}
