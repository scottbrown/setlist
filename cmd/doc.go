// Package main provides the command-line interface (CLI) entry point for
// the Setlist application.
//
// The Setlist CLI is responsible for generating an AWS SSO configuration
// file by parsing AWS Organizations and Permission Sets. It retrieves
// account information, permission sets, and session details, then formats
// them into an AWS credentials configuration file.
//
// This package leverages the Cobra library for command-line parsing and
// AWS SDK for interacting with AWS services.
//
// The primary command is executed from main.go, which invokes the root
// command defined in rootCmd.go.
//
// Usage:
//
//	setlist <command> [flags]
//
// Commands:
//
//	generate          Generate an AWS config file from SSO configuration
//	accounts          List all available AWS accounts
//	permission-sets   List all available permission sets
//	permissions       List required AWS permissions
//	check-update      Check if a newer version is available
//	init              Generate a blank configuration file
//
// Example:
//
//	setlist generate --sso-session my-session --sso-region us-east-1 --output ~/.aws/config
//	setlist accounts --sso-region us-east-1
//	setlist permission-sets --sso-region us-east-1
//	setlist permissions
//	setlist check-update
//
// Required Flags (for generate):
//
//	--sso-session    The AWS SSO session name (e.g., organization name)
//	--sso-region     The AWS region where AWS SSO is hosted
//
// Additional flags allow customization of profile mappings, output locations, and SSO-friendly names.
package main
