package main

import (
	"context"
	"fmt"
	"os"

	core "github.com/scottbrown/setlist"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     AppName,
	Short:   AppDescShort,
	Long:    AppDescLong,
	RunE:    handleRoot,
	Version: core.VERSION,
}

// handleRoot executes the main logic of the command-line application.
func handleRoot(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	var cfg aws.Config
	var err error

	if permissions {
		for _, p := range core.ListPermissionsRequired() {
			fmt.Println(p)
		}
		return nil
	}

	// check if a profile is specified
	if profile != "" {
		// create a config with the specified profile
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(ssoRegion), config.WithSharedConfigProfile(profile))
		if err != nil {
			return fmt.Errorf("failed to load AWS configuration with profile %s: %w", profile, err)
		}
	} else {
		// create a config with static credentials
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(ssoRegion))
		if err != nil {
			return fmt.Errorf("failed to load AWS configuration: %w", err)
		}
	}

	ssoClient := ssoadmin.NewFromConfig(cfg)
	orgClient := organizations.NewFromConfig(cfg)

	instance, err := core.SsoInstance(ctx, ssoClient)
	if err != nil {
		return fmt.Errorf("failed to retrieve SSO instance: %w", err)
	}

	accounts, err := core.ListAccounts(ctx, orgClient)
	if err != nil {
		return fmt.Errorf("failed to list AWS accounts: %w", err)
	}

	nicknameMapping, err := core.ParseNicknameMapping(mapping)
	if err != nil {
		return fmt.Errorf("invalid mapping format: %w", err)
	}

	configFile := core.ConfigFile{
		SessionName:     ssoSession,
		IdentityStoreId: *instance.IdentityStoreId,
		FriendlyName:    ssoFriendlyName,
		Region:          ssoRegion,
		NicknameMapping: nicknameMapping,
	}

	profiles := []core.Profile{}
	for _, account := range accounts {
		if account.Id == nil {
			fmt.Fprintf(os.Stderr, "Warning: Found account with nil ID, skipping\n")
			continue
		}

		permissionSets, err := core.PermissionSets(ctx, ssoClient, *instance.InstanceArn, *account.Id)
		if err != nil {
			return fmt.Errorf("failed to list permission sets for account %s: %w", *account.Id, err)
		}

		for _, p := range permissionSets {
			if p.Name == nil || p.Description == nil || p.SessionDuration == nil {
				fmt.Fprintf(os.Stderr, "Warning: Found incomplete permission set data for account %s, skipping\n", *account.Id)
			}

			profile := core.Profile{
				Description:     *p.Description,
				SessionDuration: *p.SessionDuration,
				SessionName:     ssoSession,
				AccountId:       *account.Id,
				RoleName:        *p.Name,
			}

			profiles = append(profiles, profile)
		}
	}

	configFile.Profiles = profiles

	builder := core.NewFileBuilder(configFile)
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
