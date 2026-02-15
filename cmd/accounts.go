package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/scottbrown/setlist"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Listing all available AWS accounts",
	Long:  "Listing all AWS accounts found in the organization, with optional filtering by account ID",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validateRegionOnly()
	},
	RunE: handleAccounts,
}

func init() {
	accountsCmd.Flags().StringVar(&includeAccounts, FlagIncludeAccounts, "", "Comma-delimited list of account IDs to include (mutually exclusive with --exclude-accounts)")
	accountsCmd.Flags().StringVar(&excludeAccounts, FlagExcludeAccounts, "", "Comma-delimited list of account IDs to exclude (mutually exclusive with --include-accounts)")

	rootCmd.AddCommand(accountsCmd)
}

func handleAccounts(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	slog.Info("Loading AWS configuration", "region", ssoRegion)
	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return err
	}

	orgClient := organizations.NewFromConfig(cfg)

	return handleListAccountsFlow(ctx, orgClient)
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

func displayAccounts(accounts []orgtypes.Account) error {
	for _, a := range accounts {
		fmt.Printf("%s\t%s\n", *a.Id, *a.Name)
	}

	return nil
}
