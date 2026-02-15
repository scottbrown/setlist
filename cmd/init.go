package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func validateRequiredFlags(cmd *cobra.Command) error {
	// Validate ssoSession
	if ssoSession == "" {
		return fmt.Errorf("required flag --%s not set", FlagSSOSession)
	}

	// Validate ssoRegion
	if ssoRegion == "" {
		return fmt.Errorf("required flag --%s not set", FlagSSORegion)
	}

	// Make sure the ssoRegion follows the basic pattern for AWS regions
	if !strings.HasPrefix(ssoRegion, "us-") &&
		!strings.HasPrefix(ssoRegion, "eu-") &&
		!strings.HasPrefix(ssoRegion, "ap-") &&
		!strings.HasPrefix(ssoRegion, "sa-") &&
		!strings.HasPrefix(ssoRegion, "ca-") &&
		!strings.HasPrefix(ssoRegion, "me-") &&
		!strings.HasPrefix(ssoRegion, "af-") {
		return fmt.Errorf("invalid region format: %s", ssoRegion)
	}

	// If writing to a file, check path validity
	if !stdout {
		// Check if directory exists
		dir := filepath.Dir(filename)
		if dir != "." {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return fmt.Errorf("output directory does not exist: %s", dir)
			}
		}

		// Try to check if the file is writable by creating it
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) //#nosec: G302
		if err != nil {
			return fmt.Errorf("cannot write to output file %s: %w", filename, err)
		}
		f.Close() //#nosec: G104

		// If it's just a test, remove the file if it was newly created
		if _, err := os.Stat(filename); err == nil {
			fi, err := os.Stat(filename)
			if err == nil && fi.Size() == 0 {
				os.Remove(filename) //#nosec: G104
			}
		}
	}

	return nil
}

func validateRegionOnly() error {
	if ssoRegion == "" {
		return fmt.Errorf("required flag --%s not set", FlagSSORegion)
	}

	if !strings.HasPrefix(ssoRegion, "us-") &&
		!strings.HasPrefix(ssoRegion, "eu-") &&
		!strings.HasPrefix(ssoRegion, "ap-") &&
		!strings.HasPrefix(ssoRegion, "sa-") &&
		!strings.HasPrefix(ssoRegion, "ca-") &&
		!strings.HasPrefix(ssoRegion, "me-") &&
		!strings.HasPrefix(ssoRegion, "af-") {
		return fmt.Errorf("invalid region format: %s", ssoRegion)
	}

	return nil
}

func configureLogging() error {
	level := slog.LevelWarn
	if verbose {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch logFormat {
	case "plain":
		handler = slog.NewTextHandler(os.Stderr, opts)
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		return fmt.Errorf("invalid log format %q: must be \"plain\" or \"json\"", logFormat)
	}

	slog.SetDefault(slog.New(handler))
	return nil
}

// init initializes command-line flags for the root command.
func init() {
	rootCmd.PersistentFlags().StringVarP(&ssoSession, FlagSSOSession, "s", "", "Nickname to give the SSO Session (e.g. org name) (required)")
	rootCmd.PersistentFlags().StringVarP(&profile, FlagProfile, "p", "", "Profile")
	rootCmd.PersistentFlags().StringVarP(&ssoRegion, FlagSSORegion, "r", "", "AWS region where AWS SSO resides (required)")
	rootCmd.PersistentFlags().StringVarP(&mapping, FlagMapping, "m", "", "Comma-delimited Account Nickname Mapping (id=nickname)")
	rootCmd.PersistentFlags().StringVarP(&filename, FlagOutput, "o", DEFAULT_FILENAME, "Where the AWS config file will be written")
	rootCmd.PersistentFlags().BoolVar(&stdout, FlagStdout, false, "Specify this flag to write the config file to stdout instead of a file")
	rootCmd.PersistentFlags().BoolVar(&permissions, FlagPermissions, false, "Specify this flag to print the required AWS permissions and then exit")
	rootCmd.PersistentFlags().StringVar(&ssoFriendlyName, FlagSSOFriendlyName, "", "Use this instead of the identity store ID for the start URL")
	rootCmd.PersistentFlags().BoolVar(&checkUpdate, FlagCheckUpdate, false, "Check if a newer version of the tool is available")
	rootCmd.PersistentFlags().BoolVar(&listAccounts, FlagListAccounts, false, "List all available AWS accounts")
	rootCmd.PersistentFlags().BoolVar(&listPermissionSets, FlagListPermissionSets, false, "List all available permission sets in the SSO instance")
	rootCmd.PersistentFlags().StringVar(&includeAccounts, FlagIncludeAccounts, "", "Comma-delimited list of account IDs to include (mutually exclusive with --exclude-accounts)")
	rootCmd.PersistentFlags().StringVar(&excludeAccounts, FlagExcludeAccounts, "", "Comma-delimited list of account IDs to exclude (mutually exclusive with --include-accounts)")
	rootCmd.PersistentFlags().StringVar(&includePermissionSets, FlagIncludePermissionSets, "", "Comma-delimited list of permission set names to include (mutually exclusive with --exclude-permission-sets)")
	rootCmd.PersistentFlags().StringVar(&excludePermissionSets, FlagExcludePermissionSets, "", "Comma-delimited list of permission set names to exclude (mutually exclusive with --include-permission-sets)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, FlagVerbose, "v", false, "Enable verbose logging output")
	rootCmd.PersistentFlags().StringVar(&logFormat, FlagLogFormat, "plain", "Log output format: \"plain\" or \"json\"")

	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := configureLogging(); err != nil {
			return err
		}

		if permissions || checkUpdate {
			return nil
		}

		if listAccounts || listPermissionSets {
			return validateRegionOnly()
		}

		return validateRequiredFlags(cmd)
	}
}
