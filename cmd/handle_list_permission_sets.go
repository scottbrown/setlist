package main

import (
	"context"
	"fmt"

	"github.com/scottbrown/setlist"

	ssotypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

func handleListPermissionSets(ctx context.Context, ssoClient setlist.SSOAdminClient, instance ssotypes.InstanceMetadata) error {
	permSets, err := setlist.AllPermissionSets(ctx, ssoClient, *instance.InstanceArn)
	if err != nil {
		return fmt.Errorf("failed to list permission sets: %w", err)
	}

	for _, ps := range permSets {
		name := ""
		if ps.Name != nil {
			name = *ps.Name
		}
		description := ""
		if ps.Description != nil {
			description = *ps.Description
		}
		fmt.Printf("%s\t%s\n", name, description)
	}

	return nil
}
