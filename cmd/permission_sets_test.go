package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

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
