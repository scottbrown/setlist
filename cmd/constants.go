package main

import (
	"time"
)

// AppName represents the application name.
const AppName string = "setlist"

// AppDescShort provides a short description of the application.
const AppDescShort string = "Creates an AWS config file from AWS SSO configuration"

// AppDescLong provides a detailed description of the application's
// functionality.
const AppDescLong string = `
Parses an AWS organizations permission set structure to build a complete .aws/config file with all permission sets provisioned across all AWS member accounts
`

// Various flag constants for command-line options.
const (
	FlagSSOSession      string = "sso-session"
	FlagSSORegion       string = "sso-region"
	FlagProfile         string = "profile"
	FlagMapping         string = "mapping"
	FlagOutput          string = "output"
  FlagPermissions     string = "permissions"
	FlagStdout          string = "stdout"
	FlagSSOFriendlyName string = "sso-friendly-name"
)

// Default output filename if no filename is specified
const DEFAULT_FILENAME string = "aws.config"

// Default timeout (in minutes) for all calls that use a context
const DEFAULT_TIMEOUT time.Duration = 2 * time.Minute
