package main

import (
	"context"
	"log/slog"

	"github.com/scottbrown/setlist"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generating an AWS config file from SSO configuration",
	Long:  "Parsing AWS Organizations and Permission Sets to build a complete .aws/config file with all permission sets provisioned across all AWS member accounts",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validateRequiredFlags(cmd)
	},
	RunE: handleGenerate,
}

func init() {
	generateCmd.Flags().StringVarP(&filename, FlagOutput, "o", DEFAULT_FILENAME, "Where the AWS config file will be written")
	generateCmd.Flags().BoolVar(&stdout, FlagStdout, false, "Specify this flag to write the config file to stdout instead of a file")
	generateCmd.Flags().StringVarP(&mapping, FlagMapping, "m", "", "Comma-delimited Account Nickname Mapping (id=nickname)")
	generateCmd.Flags().StringVar(&ssoFriendlyName, FlagSSOFriendlyName, "", "Use this instead of the identity store ID for the start URL")
	generateCmd.Flags().StringVar(&includeAccounts, FlagIncludeAccounts, "", "Comma-delimited list of account IDs to include (mutually exclusive with --exclude-accounts)")
	generateCmd.Flags().StringVar(&excludeAccounts, FlagExcludeAccounts, "", "Comma-delimited list of account IDs to exclude (mutually exclusive with --include-accounts)")
	generateCmd.Flags().StringVar(&includePermissionSets, FlagIncludePermissionSets, "", "Comma-delimited list of permission set names to include (mutually exclusive with --exclude-permission-sets)")
	generateCmd.Flags().StringVar(&excludePermissionSets, FlagExcludePermissionSets, "", "Comma-delimited list of permission set names to exclude (mutually exclusive with --include-permission-sets)")

	rootCmd.AddCommand(generateCmd)
}

func handleGenerate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	slog.Info("Loading AWS configuration", "region", ssoRegion)
	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return err
	}

	ssoClient := ssoadmin.NewFromConfig(cfg)
	orgClient := organizations.NewFromConfig(cfg)

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
