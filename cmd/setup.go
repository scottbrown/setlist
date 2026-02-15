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
	if ssoSession == "" {
		return fmt.Errorf("required flag --%s not set", FlagSSOSession)
	}

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

	if !stdout {
		dir := filepath.Dir(filename)
		if dir != "." {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return fmt.Errorf("output directory does not exist: %s", dir)
			}
		}

		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) //#nosec: G302
		if err != nil {
			return fmt.Errorf("cannot write to output file %s: %w", filename, err)
		}
		f.Close() //#nosec: G104

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

func init() {
	rootCmd.PersistentFlags().StringVarP(&ssoSession, FlagSSOSession, "s", "", "Nickname to give the SSO Session (e.g. org name) (required)")
	rootCmd.PersistentFlags().StringVarP(&profile, FlagProfile, "p", "", "Profile")
	rootCmd.PersistentFlags().StringVarP(&ssoRegion, FlagSSORegion, "r", "", "AWS region where AWS SSO resides (required)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, FlagVerbose, "v", false, "Enable verbose logging output")
	rootCmd.PersistentFlags().StringVar(&logFormat, FlagLogFormat, "plain", "Log output format: \"plain\" or \"json\"")
	rootCmd.PersistentFlags().StringVarP(&configFile, FlagConfig, "c", "", "Path to config file (default: ~/.setlist.yaml)")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := configureLogging(); err != nil {
			return err
		}

		if err := loadConfigFile(cmd); err != nil {
			return err
		}

		return configureLogging()
	}
}
