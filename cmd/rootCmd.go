package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/scottbrown/setlist"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     AppName,
	Short:   AppDescShort,
	Long:    AppDescLong,
	Version: setlist.VERSION,
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
