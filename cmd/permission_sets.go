package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/scottbrown/setlist"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/spf13/cobra"
)

var permissionSetsCmd = &cobra.Command{
	Use:   "permission-sets",
	Short: "Listing all available permission sets",
	Long:  "Listing all permission sets found in the SSO instance",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validateRegionOnly()
	},
	RunE: handlePermissionSetsCommand,
}

func init() {
	rootCmd.AddCommand(permissionSetsCmd)
}

func handlePermissionSetsCommand(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	slog.Info("Loading AWS configuration", "region", ssoRegion)
	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return err
	}

	ssoClient := ssoadmin.NewFromConfig(cfg)

	return handleListPermissionSetsFlow(ctx, ssoClient)
}

func handleListPermissionSetsFlow(ctx context.Context, ssoClient setlist.SSOAdminClient) error {
	slog.Info("Retrieving SSO instance")
	instance, err := setlist.SsoInstance(ctx, ssoClient)
	if err != nil {
		return fmt.Errorf("failed to retrieve SSO instance: %w", err)
	}

	return handleListPermissionSets(ctx, ssoClient, instance)
}

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
