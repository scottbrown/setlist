package awsconfigcreator

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

// get the SSO instance ARN (there's only one allowed)
func SsoInstance(ctx context.Context, client *ssoadmin.Client) (types.InstanceMetadata, error) {
	resp, err := client.ListInstances(ctx, nil)
	if err != nil {
		return types.InstanceMetadata{}, err
	}

	if len(resp.Instances) == 0 {
		return types.InstanceMetadata{}, errors.New("SSO is not enabled.  No SSO instances exist.")
	}

	instance := resp.Instances[0]

	return instance, nil
}

func PermissionSets(ctx context.Context, client *ssoadmin.Client, instanceArn string, accountId string) ([]types.PermissionSet, error) {
	permissionSets := []types.PermissionSet{}

	params := &ssoadmin.ListPermissionSetsProvisionedToAccountInput{
		InstanceArn: aws.String(instanceArn),
		AccountId:   aws.String(accountId),
	}
	resp, err := client.ListPermissionSetsProvisionedToAccount(ctx, params)

	if err != nil {
		return permissionSets, err
	}

	// loop through permissions sets
	for _, permissionSetArn := range resp.PermissionSets {
		// get permission set name
		params := &ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(permissionSetArn),
		}
		resp, err := client.DescribePermissionSet(ctx, params)
		if err != nil {
			return permissionSets, err
		}

		permissionSets = append(permissionSets, *resp.PermissionSet)
	}

	return permissionSets, nil
}
