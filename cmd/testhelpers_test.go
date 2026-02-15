package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
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

type mockOrganizationsClient struct {
	ListAccountsFunc func(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error)
}

func (m *mockOrganizationsClient) ListAccounts(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
	return m.ListAccountsFunc(ctx, params, optFns...)
}

