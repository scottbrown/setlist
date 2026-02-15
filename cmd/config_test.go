package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringVar(&ssoSession, FlagSSOSession, "", "")
	cmd.Flags().StringVar(&ssoRegion, FlagSSORegion, "", "")
	cmd.Flags().StringVar(&profile, FlagProfile, "", "")
	cmd.Flags().StringVar(&mapping, FlagMapping, "", "")
	cmd.Flags().StringVar(&filename, FlagOutput, DEFAULT_FILENAME, "")
	cmd.Flags().BoolVar(&stdout, FlagStdout, false, "")
	cmd.Flags().StringVar(&ssoFriendlyName, FlagSSOFriendlyName, "", "")
	cmd.Flags().BoolVar(&verbose, FlagVerbose, false, "")
	cmd.Flags().StringVar(&logFormat, FlagLogFormat, "plain", "")
	cmd.Flags().StringVar(&includeAccounts, FlagIncludeAccounts, "", "")
	cmd.Flags().StringVar(&excludeAccounts, FlagExcludeAccounts, "", "")
	cmd.Flags().StringVar(&includePermissionSets, FlagIncludePermissionSets, "", "")
	cmd.Flags().StringVar(&excludePermissionSets, FlagExcludePermissionSets, "", "")
	cmd.Flags().StringVar(&configFile, FlagConfig, "", "")
	return cmd
}

func resetGlobals() {
	ssoSession = ""
	ssoRegion = ""
	profile = ""
	mapping = ""
	filename = DEFAULT_FILENAME
	stdout = false
	ssoFriendlyName = ""
	verbose = false
	logFormat = "plain"
	includeAccounts = ""
	excludeAccounts = ""
	includePermissionSets = ""
	excludePermissionSets = ""
	configFile = ""
}

func TestLoadConfigFile_ValidConfig(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `sso-session: myorg
sso-region: us-east-1
profile: admin
mapping: "123456789012=prod"
output: ~/.aws/config
stdout: false
sso-friendly-name: my-company
verbose: true
log-format: json
include-accounts: "111111111111"
exclude-accounts: "222222222222"
include-permission-sets: AdminAccess
exclude-permission-sets: ReadOnly
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCommand()
	configFile = cfgPath
	cmd.Flags().Set(FlagConfig, cfgPath)

	if err := loadConfigFile(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ssoSession != "myorg" {
		t.Errorf("ssoSession = %q, want %q", ssoSession, "myorg")
	}
	if ssoRegion != "us-east-1" {
		t.Errorf("ssoRegion = %q, want %q", ssoRegion, "us-east-1")
	}
	if profile != "admin" {
		t.Errorf("profile = %q, want %q", profile, "admin")
	}
	if mapping != "123456789012=prod" {
		t.Errorf("mapping = %q, want %q", mapping, "123456789012=prod")
	}
	if filename != "~/.aws/config" {
		t.Errorf("filename = %q, want %q", filename, "~/.aws/config")
	}
	if stdout != false {
		t.Errorf("stdout = %v, want false", stdout)
	}
	if ssoFriendlyName != "my-company" {
		t.Errorf("ssoFriendlyName = %q, want %q", ssoFriendlyName, "my-company")
	}
	if verbose != true {
		t.Errorf("verbose = %v, want true", verbose)
	}
	if logFormat != "json" {
		t.Errorf("logFormat = %q, want %q", logFormat, "json")
	}
	if includeAccounts != "111111111111" {
		t.Errorf("includeAccounts = %q, want %q", includeAccounts, "111111111111")
	}
	if excludeAccounts != "222222222222" {
		t.Errorf("excludeAccounts = %q, want %q", excludeAccounts, "222222222222")
	}
	if includePermissionSets != "AdminAccess" {
		t.Errorf("includePermissionSets = %q, want %q", includePermissionSets, "AdminAccess")
	}
	if excludePermissionSets != "ReadOnly" {
		t.Errorf("excludePermissionSets = %q, want %q", excludePermissionSets, "ReadOnly")
	}
}

func TestLoadConfigFile_FlagOverridesConfig(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `sso-session: from-config
sso-region: eu-west-1
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCommand()
	configFile = cfgPath
	cmd.Flags().Set(FlagConfig, cfgPath)
	cmd.Flags().Set(FlagSSOSession, "from-flag")

	if err := loadConfigFile(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ssoSession != "from-flag" {
		t.Errorf("ssoSession = %q, want %q (flag should override config)", ssoSession, "from-flag")
	}
	if ssoRegion != "eu-west-1" {
		t.Errorf("ssoRegion = %q, want %q (config should apply when flag not set)", ssoRegion, "eu-west-1")
	}
}

func TestLoadConfigFile_MissingFileNoFlag(t *testing.T) {
	resetGlobals()

	cmd := newTestCommand()

	origHome := os.Getenv("HOME")
	dir := t.TempDir()
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	if err := loadConfigFile(cmd); err != nil {
		t.Fatalf("missing config without --config should be silent, got: %v", err)
	}
}

func TestLoadConfigFile_MissingFileWithFlag(t *testing.T) {
	resetGlobals()

	cmd := newTestCommand()
	configFile = "/nonexistent/path/config.yaml"
	cmd.Flags().Set(FlagConfig, configFile)

	err := loadConfigFile(cmd)
	if err == nil {
		t.Fatal("expected error when --config points to nonexistent file")
	}
	if !strings.Contains(err.Error(), "unable to read config file") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "unable to read config file")
	}
}

func TestLoadConfigFile_InvalidYAML(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("{{invalid yaml:::"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCommand()
	configFile = cfgPath
	cmd.Flags().Set(FlagConfig, cfgPath)

	err := loadConfigFile(cmd)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "unable to parse config file") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "unable to parse config file")
	}
}

func TestLoadConfigFile_PartialConfig(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `sso-session: partial-org
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCommand()
	configFile = cfgPath
	cmd.Flags().Set(FlagConfig, cfgPath)

	if err := loadConfigFile(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ssoSession != "partial-org" {
		t.Errorf("ssoSession = %q, want %q", ssoSession, "partial-org")
	}
	if ssoRegion != "" {
		t.Errorf("ssoRegion = %q, want empty (not set in config)", ssoRegion)
	}
	if logFormat != "plain" {
		t.Errorf("logFormat = %q, want %q (default should remain)", logFormat, "plain")
	}
}

func TestLoadConfigFile_BoolNilVsFalse(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// Config with stdout explicitly set to false
	content := `stdout: false
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCommand()
	configFile = cfgPath
	cmd.Flags().Set(FlagConfig, cfgPath)

	if err := loadConfigFile(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout != false {
		t.Errorf("stdout = %v, want false", stdout)
	}

	// Now test with stdout not in config at all — verbose should stay at default
	resetGlobals()
	content = `sso-session: test
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd = newTestCommand()
	verbose = false // default
	configFile = cfgPath
	cmd.Flags().Set(FlagConfig, cfgPath)

	if err := loadConfigFile(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if verbose != false {
		t.Errorf("verbose = %v, want false (should remain at default when not in config)", verbose)
	}
}

func TestLoadConfigFile_BoolTrue(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `verbose: true
stdout: true
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCommand()
	configFile = cfgPath
	cmd.Flags().Set(FlagConfig, cfgPath)

	if err := loadConfigFile(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if verbose != true {
		t.Errorf("verbose = %v, want true", verbose)
	}
	if stdout != true {
		t.Errorf("stdout = %v, want true", stdout)
	}
}

func TestLoadConfigFile_CustomPath(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	customPath := filepath.Join(dir, "custom", "my-config.yaml")
	if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
		t.Fatal(err)
	}
	content := `sso-session: custom-org
`
	if err := os.WriteFile(customPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCommand()
	configFile = customPath
	cmd.Flags().Set(FlagConfig, customPath)

	if err := loadConfigFile(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ssoSession != "custom-org" {
		t.Errorf("ssoSession = %q, want %q", ssoSession, "custom-org")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path, err := defaultConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".setlist.yaml")
	if path != expected {
		t.Errorf("defaultConfigPath() = %q, want %q", path, expected)
	}
}

func TestConfigFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup(FlagConfig)
	if flag == nil {
		t.Errorf("Expected --%s flag to be registered", FlagConfig)
	}
	if flag.Shorthand != "c" {
		t.Errorf("Expected --%s shorthand to be 'c', got %q", FlagConfig, flag.Shorthand)
	}
}
