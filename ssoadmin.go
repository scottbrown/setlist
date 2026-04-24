package setlist

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

const describePermissionSetConcurrency = 5

// Define interface for the SSO Admin client to make testing easier
type SSOAdminClient interface {
	ListInstances(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error)
	ListPermissionSets(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error)
	ListPermissionSetsProvisionedToAccount(ctx context.Context, params *ssoadmin.ListPermissionSetsProvisionedToAccountInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsProvisionedToAccountOutput, error)
	DescribePermissionSet(ctx context.Context, params *ssoadmin.DescribePermissionSetInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.DescribePermissionSetOutput, error)
}

// SsoInstance retrieves the AWS SSO instance metadata from the AWS account.
// AWS SSO allows only a single instance per organization, so this function
// returns the first (and only) instance found. It validates that required
// fields exist in the response and returns an error if the SSO service is
// not properly configured.
func SsoInstance(ctx context.Context, client SSOAdminClient) (types.InstanceMetadata, error) {
	resp, err := client.ListInstances(ctx, nil)
	if err != nil {
		return types.InstanceMetadata{}, fmt.Errorf("failed to list SSO instances: %w", err)
	}

	if len(resp.Instances) == 0 {
		return types.InstanceMetadata{}, errors.New("SSO is not enabled. No SSO instances exist")
	}

	instance := resp.Instances[0]

	// Validate required fields
	if instance.InstanceArn == nil {
		return types.InstanceMetadata{}, errors.New("received instance with nil InstanceArn")
	}

	if instance.IdentityStoreId == nil {
		return types.InstanceMetadata{}, errors.New("received instance with nil IdentityStoreId")
	}

	return instance, nil
}

// PermissionSets retrieves the list of permission sets provisioned to an
// account.
func PermissionSets(ctx context.Context, client SSOAdminClient, instanceArn string, accountId string) ([]types.PermissionSet, error) {
	// Validate input parameters
	if instanceArn == "" {
		return nil, errors.New("invalid parameter: empty instanceArn")
	}

	if accountId == "" {
		return nil, errors.New("invalid parameter: empty accountId")
	}

	permissionSets := []types.PermissionSet{}

	var permissionSetArns []string
	var token *string
	for {
		params := &ssoadmin.ListPermissionSetsProvisionedToAccountInput{
			InstanceArn: aws.String(instanceArn),
			AccountId:   aws.String(accountId),
			NextToken:   token,
		}
		resp, err := client.ListPermissionSetsProvisionedToAccount(ctx, params)

		if err != nil {
			return permissionSets, fmt.Errorf("failed to list permission sets for account %s: %w", accountId, err)
		}

		for _, i := range resp.PermissionSets {
			permissionSetArns = append(permissionSetArns, i)
		}

		if resp.NextToken == nil {
			break
		}

		token = resp.NextToken
	}

	return describePermissionSetsConcurrently(ctx, client, instanceArn, permissionSetArns)
}

// AllPermissionSets retrieves all permission sets defined in an SSO instance.
// It paginates through the ListPermissionSets API to collect all ARNs, then
// calls DescribePermissionSet for each to get the full details.
func AllPermissionSets(ctx context.Context, client SSOAdminClient, instanceArn string) ([]types.PermissionSet, error) {
	if instanceArn == "" {
		return nil, errors.New("invalid parameter: empty instanceArn")
	}

	var arns []string
	var token *string
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err := client.ListPermissionSets(ctx, &ssoadmin.ListPermissionSetsInput{
			InstanceArn: aws.String(instanceArn),
			NextToken:   token,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list permission sets: %w", err)
		}

		arns = append(arns, resp.PermissionSets...)

		if resp.NextToken == nil {
			break
		}
		token = resp.NextToken
	}

	return describePermissionSetsConcurrently(ctx, client, instanceArn, arns)
}

type describeResult struct {
	index int
	ps    types.PermissionSet
	err   error
}

func describePermissionSetsConcurrently(ctx context.Context, client SSOAdminClient, instanceArn string, arns []string) ([]types.PermissionSet, error) {
	if len(arns) == 0 {
		return []types.PermissionSet{}, nil
	}

	results := make([]describeResult, len(arns))
	sem := make(chan struct{}, describePermissionSetConcurrency)
	var wg sync.WaitGroup

	for i, arn := range arns {
		wg.Add(1)
		go func(idx int, permSetArn string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			resp, err := client.DescribePermissionSet(ctx, &ssoadmin.DescribePermissionSetInput{
				InstanceArn:      aws.String(instanceArn),
				PermissionSetArn: aws.String(permSetArn),
			})
			if err != nil {
				results[idx] = describeResult{index: idx, err: fmt.Errorf("failed to describe permission set %s: %w", permSetArn, err)}
				return
			}
			if resp.PermissionSet == nil {
				results[idx] = describeResult{index: idx, err: fmt.Errorf("nil permission set returned for ARN %s", permSetArn)}
				return
			}
			results[idx] = describeResult{index: idx, ps: *resp.PermissionSet}
		}(i, arn)
	}

	wg.Wait()

	permissionSets := make([]types.PermissionSet, 0, len(arns))
	for _, r := range results {
		if r.err != nil {
			return permissionSets, r.err
		}
		permissionSets = append(permissionSets, r.ps)
	}

	return permissionSets, nil
}
