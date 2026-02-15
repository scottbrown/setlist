package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

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
