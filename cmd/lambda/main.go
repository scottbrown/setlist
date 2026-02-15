package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/scottbrown/setlist"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
)

func handleRequest(ctx context.Context) error {
	ssoSession := os.Getenv("SSO_SESSION")
	ssoRegion := os.Getenv("SSO_REGION")
	s3Bucket := os.Getenv("S3_BUCKET")
	s3Key := os.Getenv("S3_KEY")
	ssoFriendlyName := os.Getenv("SSO_FRIENDLY_NAME")
	nicknameMapping := os.Getenv("NICKNAME_MAPPING")
	includeAccounts := os.Getenv("INCLUDE_ACCOUNTS")
	excludeAccounts := os.Getenv("EXCLUDE_ACCOUNTS")
	includePermissionSets := os.Getenv("INCLUDE_PERMISSION_SETS")
	excludePermissionSets := os.Getenv("EXCLUDE_PERMISSION_SETS")

	if ssoSession == "" {
		return fmt.Errorf("SSO_SESSION environment variable is required")
	}
	if ssoRegion == "" {
		return fmt.Errorf("SSO_REGION environment variable is required")
	}
	if s3Bucket == "" {
		return fmt.Errorf("S3_BUCKET environment variable is required")
	}
	if s3Key == "" {
		return fmt.Errorf("S3_KEY environment variable is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(ssoRegion))
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	ssoClient := ssoadmin.NewFromConfig(cfg)
	orgClient := organizations.NewFromConfig(cfg)

	slog.Info("Generating AWS config file")
	configFile, err := setlist.Generate(ctx, setlist.GenerateInput{
		SSOClient:             ssoClient,
		OrgClient:             orgClient,
		SessionName:           ssoSession,
		Region:                ssoRegion,
		FriendlyName:          ssoFriendlyName,
		NicknameMapping:       nicknameMapping,
		IncludeAccounts:       includeAccounts,
		ExcludeAccounts:       excludeAccounts,
		IncludePermissionSets: includePermissionSets,
		ExcludePermissionSets: excludePermissionSets,
	})
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	builder := setlist.NewFileBuilder(configFile)
	payload, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to build config file: %w", err)
	}

	var buf bytes.Buffer
	if _, err := payload.WriteTo(&buf); err != nil {
		return fmt.Errorf("failed to write config to buffer: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	body := bytes.NewReader(buf.Bytes())
	contentType := "text/plain"

	slog.Info("Uploading config to S3", "bucket", s3Bucket, "key", s3Key)
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s3Bucket,
		Key:         &s3Key,
		Body:        body,
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	slog.Info("Config file uploaded successfully", "bucket", s3Bucket, "key", s3Key)
	return nil
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	lambda.Start(handleRequest)
}
