package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const configTemplate = `# SetList configuration file
# See: https://github.com/scottbrown/setlist

# (Required) Nickname for the SSO session (e.g. your org name)
sso-session: ""

# (Required) AWS region where AWS SSO resides (e.g. ca-central-1)
sso-region: ""

# AWS profile to use for credentials
profile: ""

# Comma-delimited account nickname mapping (id=nickname)
mapping: ""

# Output filename (default: aws.config)
output: ""

# Write output to stdout instead of a file
stdout: false

# Use a friendly name instead of the identity store ID for the start URL
sso-friendly-name: ""

# Enable verbose logging
verbose: false

# Log format: "plain" or "json"
log-format: "plain"

# Comma-delimited list of account IDs to include (mutually exclusive with exclude-accounts)
include-accounts: ""

# Comma-delimited list of account IDs to exclude (mutually exclusive with include-accounts)
exclude-accounts: ""

# Comma-delimited list of permission set names to include
include-permission-sets: ""

# Comma-delimited list of permission set names to exclude
exclude-permission-sets: ""
`

var forceOverwrite bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generating a blank configuration file",
	Long:  "Generating a blank .setlist.yaml configuration file with all available options and descriptive comments",
	RunE:  handleInit,
}

func init() {
	initCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing configuration file")
	rootCmd.AddCommand(initCmd)
}

func handleInit(cmd *cobra.Command, args []string) error {
	path := configFile
	if path == "" {
		defaultPath, err := defaultConfigPath()
		if err != nil {
			return err
		}
		path = defaultPath
	}

	if !forceOverwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("configuration file already exists at %s (use --force to overwrite)", path)
		}
	}

	if err := os.WriteFile(path, []byte(configTemplate), 0644); err != nil { //#nosec: G306
		return fmt.Errorf("unable to write configuration file %s: %w", path, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Wrote configuration file to %s\n", path)
	return nil
}
