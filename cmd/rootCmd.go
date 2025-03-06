package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/scottbrown/setlist"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     AppName,
	Short:   AppDescShort,
	Long:    AppDescLong,
	RunE:    handleRoot,
	Version: setlist.VERSION,
}

// handleRoot executes the main logic of the command-line application.
// It retrieves AWS accounts and permission sets, builds the configuration
// structure, and outputs it to the specified destination.
func handleRoot(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	// Handle the check-update flag if specified
	if checkUpdate {
		return handleCheckUpdate(ctx)
	}

	if permissions {
		return handleListPermissions()
	}

	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return err
	}

	ssoClient := ssoadmin.NewFromConfig(cfg)
	orgClient := organizations.NewFromConfig(cfg)

	instance, err := setlist.SsoInstance(ctx, ssoClient)
	if err != nil {
		return fmt.Errorf("failed to retrieve SSO instance: %w", err)
	}

	accounts, err := setlist.ListAccounts(ctx, orgClient)
	if err != nil {
		return fmt.Errorf("failed to list AWS accounts: %w", err)
	}

	nicknameMapping, err := setlist.ParseNicknameMapping(mapping)
	if err != nil {
		return fmt.Errorf("invalid mapping format: %w", err)
	}

	if listAccounts {
		return displayAccounts(accounts)
	}

	configFile, err := buildConfigFile(ctx, ssoClient, instance, accounts, nicknameMapping)
	if err != nil {
		return err
	}

	return outputConfig(configFile)
}

// loadAWSConfig loads AWS configuration based on provided flags
func loadAWSConfig(ctx context.Context) (aws.Config, error) {
	if profile != "" {
		// Create a config with the specified profile
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(ssoRegion),
			config.WithSharedConfigProfile(profile))
		if err != nil {
			return aws.Config{}, fmt.Errorf("failed to load AWS configuration with profile %s: %w", profile, err)
		}
		return cfg, nil
	}

	// Create a config with default credentials
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(ssoRegion))
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS configuration: %w", err)
	}
	return cfg, nil
}

// buildConfigFile constructs the configuration file with all profiles
func buildConfigFile(
	ctx context.Context,
	ssoClient *ssoadmin.Client,
	instance ssotypes.InstanceMetadata,
	accounts []orgtypes.Account,
	nicknameMapping map[string]string,
) (setlist.ConfigFile, error) {
	configFile := setlist.ConfigFile{
		SessionName:     ssoSession,
		IdentityStoreId: *instance.IdentityStoreId,
		FriendlyName:    ssoFriendlyName,
		Region:          ssoRegion,
		NicknameMapping: nicknameMapping,
	}

	profiles, err := buildProfiles(ctx, ssoClient, instance, accounts)
	if err != nil {
		return configFile, err
	}

	configFile.Profiles = profiles
	return configFile, nil
}

// buildProfiles builds all the profile configurations
func buildProfiles(
	ctx context.Context,
	ssoClient *ssoadmin.Client,
	instance ssotypes.InstanceMetadata,
	accounts []orgtypes.Account,
) ([]setlist.Profile, error) {
	profiles := []setlist.Profile{}

	includedAccounts := buildIncludedAccounts(accounts)
	excludedAccounts := buildExcludedAccounts()

	for _, account := range accounts {
		if account.Id == nil {
			fmt.Fprintf(os.Stderr, "Warning: Found account with nil ID, skipping\n")
			continue
		}

		if !includedAccounts.Contains(*account.Id) {
			continue
		}

		if excludedAccounts.Contains(*account.Id) {
			continue
		}

		permissionSets, err := setlist.PermissionSets(ctx, ssoClient, *instance.InstanceArn, *account.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to list permission sets for account %s: %w", *account.Id, err)
		}

		accountProfiles, err := buildAccountProfiles(account, permissionSets)
		if err != nil {
			return nil, err
		}

		profiles = append(profiles, accountProfiles...)
	}

	return profiles, nil
}

// buildAccountProfiles builds profiles for a specific account
func buildAccountProfiles(
	account orgtypes.Account,
	permissionSets []ssotypes.PermissionSet,
) ([]setlist.Profile, error) {
	profiles := []setlist.Profile{}

	for _, p := range permissionSets {
		if p.Name == nil || p.Description == nil || p.SessionDuration == nil {
			fmt.Fprintf(os.Stderr, "Warning: Found incomplete permission set data for account %s, skipping\n", *account.Id)
			continue
		}

		profile := setlist.Profile{
			Description:     *p.Description,
			SessionDuration: *p.SessionDuration,
			SessionName:     ssoSession,
			AccountId:       *account.Id,
			RoleName:        *p.Name,
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// outputConfig writes the configuration to the specified destination
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

type AccountsFilter []string

func buildIncludedAccounts(accounts []orgtypes.Account) AccountsFilter {
	if includeAccounts == "" {
		var x AccountsFilter
		for _, i := range accounts {
			x = append(x, *i.Id)
		}
		return x
	}

	return strings.Split(includeAccounts, ",")
}

func buildExcludedAccounts() AccountsFilter {
	if includeAccounts == "" {
		return AccountsFilter{}
	}

	return strings.Split(excludeAccounts, ",")
}

func (a AccountsFilter) Contains(id string) bool {
	for _, i := range a {
		if i == id {
			return true
		}
	}

	return false
}
