package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type SetlistConfig struct {
	SSOSession            string `yaml:"sso-session"`
	SSORegion             string `yaml:"sso-region"`
	Profile               string `yaml:"profile"`
	Mapping               string `yaml:"mapping"`
	Output                string `yaml:"output"`
	Stdout                *bool  `yaml:"stdout"`
	SSOFriendlyName       string `yaml:"sso-friendly-name"`
	Verbose               *bool  `yaml:"verbose"`
	LogFormat             string `yaml:"log-format"`
	IncludeAccounts       string `yaml:"include-accounts"`
	ExcludeAccounts       string `yaml:"exclude-accounts"`
	IncludePermissionSets string `yaml:"include-permission-sets"`
	ExcludePermissionSets string `yaml:"exclude-permission-sets"`
}

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	return filepath.Join(home, ".setlist.yaml"), nil
}

func loadConfigFile(cmd *cobra.Command) error {
	configExplicit := cmd.Flags().Changed(FlagConfig)

	path := configFile
	if path == "" {
		defaultPath, err := defaultConfigPath()
		if err != nil {
			return err
		}
		path = defaultPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && !configExplicit {
			return nil
		}
		return fmt.Errorf("unable to read config file %s: %w", path, err)
	}

	var cfg SetlistConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("unable to parse config file %s: %w", path, err)
	}

	applyConfig(cmd, &cfg)
	return nil
}

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func applyConfig(cmd *cobra.Command, cfg *SetlistConfig) {
	if !cmd.Flags().Changed(FlagSSOSession) && cfg.SSOSession != "" {
		ssoSession = cfg.SSOSession
	}
	if !cmd.Flags().Changed(FlagSSORegion) && cfg.SSORegion != "" {
		ssoRegion = cfg.SSORegion
	}
	if !cmd.Flags().Changed(FlagProfile) && cfg.Profile != "" {
		profile = cfg.Profile
	}
	if flagExists(cmd, FlagMapping) && !cmd.Flags().Changed(FlagMapping) && cfg.Mapping != "" {
		mapping = cfg.Mapping
	}
	if flagExists(cmd, FlagOutput) && !cmd.Flags().Changed(FlagOutput) && cfg.Output != "" {
		filename = cfg.Output
	}
	if flagExists(cmd, FlagStdout) && !cmd.Flags().Changed(FlagStdout) && cfg.Stdout != nil {
		stdout = *cfg.Stdout
	}
	if flagExists(cmd, FlagSSOFriendlyName) && !cmd.Flags().Changed(FlagSSOFriendlyName) && cfg.SSOFriendlyName != "" {
		ssoFriendlyName = cfg.SSOFriendlyName
	}
	if !cmd.Flags().Changed(FlagVerbose) && cfg.Verbose != nil {
		verbose = *cfg.Verbose
	}
	if !cmd.Flags().Changed(FlagLogFormat) && cfg.LogFormat != "" {
		logFormat = cfg.LogFormat
	}
	if flagExists(cmd, FlagIncludeAccounts) && !cmd.Flags().Changed(FlagIncludeAccounts) && cfg.IncludeAccounts != "" {
		includeAccounts = cfg.IncludeAccounts
	}
	if flagExists(cmd, FlagExcludeAccounts) && !cmd.Flags().Changed(FlagExcludeAccounts) && cfg.ExcludeAccounts != "" {
		excludeAccounts = cfg.ExcludeAccounts
	}
	if flagExists(cmd, FlagIncludePermissionSets) && !cmd.Flags().Changed(FlagIncludePermissionSets) && cfg.IncludePermissionSets != "" {
		includePermissionSets = cfg.IncludePermissionSets
	}
	if flagExists(cmd, FlagExcludePermissionSets) && !cmd.Flags().Changed(FlagExcludePermissionSets) && cfg.ExcludePermissionSets != "" {
		excludePermissionSets = cfg.ExcludePermissionSets
	}
}
