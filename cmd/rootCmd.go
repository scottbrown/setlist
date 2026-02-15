package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/scottbrown/setlist"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     AppName,
	Short:   AppDescShort,
	Long:    AppDescLong,
	RunE:    handleRoot,
	Version: setlist.VERSION,
}

func handleRoot(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	if checkUpdate {
		return handleCheckUpdate(ctx)
	}

	if permissions {
		return handleListPermissions()
	}

	slog.Info("Loading AWS configuration", "region", ssoRegion)
	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return err
	}

	ssoClient := ssoadmin.NewFromConfig(cfg)
	orgClient := organizations.NewFromConfig(cfg)

	if listAccounts {
		return handleListAccountsFlow(ctx, orgClient)
	}

	if listPermissionSets {
		return handleListPermissionSetsFlow(ctx, ssoClient)
	}

	configFile, err := setlist.Generate(ctx, setlist.GenerateInput{
		SSOClient:             ssoClient,
		OrgClient:             orgClient,
		SessionName:           ssoSession,
		Region:                ssoRegion,
		FriendlyName:          ssoFriendlyName,
		NicknameMapping:       mapping,
		IncludeAccounts:       includeAccounts,
		ExcludeAccounts:       excludeAccounts,
		IncludePermissionSets: includePermissionSets,
		ExcludePermissionSets: excludePermissionSets,
	})
	if err != nil {
		return err
	}

	slog.Info("Writing output")
	return outputConfig(configFile)
}

func loadAWSConfig(ctx context.Context) (aws.Config, error) {
	if profile != "" {
		slog.Info("Loading AWS config with profile", "profile", profile)
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(ssoRegion),
			config.WithSharedConfigProfile(profile))
		if err != nil {
			return aws.Config{}, fmt.Errorf("failed to load AWS configuration with profile %s: %w", profile, err)
		}
		return cfg, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(ssoRegion))
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS configuration: %w", err)
	}
	return cfg, nil
}

func handleListAccountsFlow(ctx context.Context, orgClient setlist.OrganizationsClient) error {
	slog.Info("Listing AWS accounts")
	accounts, err := setlist.ListAccounts(ctx, orgClient)
	if err != nil {
		return fmt.Errorf("failed to list AWS accounts: %w", err)
	}

	includeList, err := setlist.ParseAccountIdList(includeAccounts)
	if err != nil {
		return fmt.Errorf("invalid include-accounts: %w", err)
	}

	excludeList, err := setlist.ParseAccountIdList(excludeAccounts)
	if err != nil {
		return fmt.Errorf("invalid exclude-accounts: %w", err)
	}

	accounts, err = setlist.FilterAccounts(accounts, includeList, excludeList)
	if err != nil {
		return fmt.Errorf("account filter error: %w", err)
	}

	return displayAccounts(accounts)
}

func handleListPermissionSetsFlow(ctx context.Context, ssoClient setlist.SSOAdminClient) error {
	slog.Info("Retrieving SSO instance")
	instance, err := setlist.SsoInstance(ctx, ssoClient)
	if err != nil {
		return fmt.Errorf("failed to retrieve SSO instance: %w", err)
	}

	return handleListPermissionSets(ctx, ssoClient, instance)
}

func outputConfig(configFile setlist.ConfigFile) error {
	builder := setlist.NewFileBuilder(configFile)
	payload, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to build config file: %w", err)
	}

	if stdout {
		if _, err := payload.WriteTo(os.Stdout); err != nil {
			return fmt.Errorf("failed to write config to stdout: %w", err)
		}
	} else {
		if err := payload.SaveTo(filename); err != nil {
			return err
		}
		fmt.Printf("Wrote to %s\n", filename)
	}

	return nil
}

func displayAccounts(accounts []orgtypes.Account) error {
	for _, a := range accounts {
		fmt.Printf("%s\t%s\n", *a.Id, *a.Name)
	}

	return nil
}
