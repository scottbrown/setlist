package main

import (
	"time"
)

const AppName string = "setlist"

const AppDescShort string = "Creates an AWS config file from AWS SSO configuration"

const AppDescLong string = `
Parses an AWS organizations permission set structure to build a complete .aws/config file with all permission sets provisioned across all AWS member accounts
`

const (
	FlagSSOSession            string = "sso-session"
	FlagSSORegion             string = "sso-region"
	FlagProfile               string = "profile"
	FlagMapping               string = "mapping"
	FlagOutput                string = "output"
	FlagStdout                string = "stdout"
	FlagSSOFriendlyName       string = "sso-friendly-name"
	FlagIncludeAccounts       string = "include-accounts"
	FlagExcludeAccounts       string = "exclude-accounts"
	FlagIncludePermissionSets string = "include-permission-sets"
	FlagExcludePermissionSets string = "exclude-permission-sets"
	FlagVerbose               string = "verbose"
	FlagLogFormat             string = "log-format"
	FlagConfig                string = "config"
)

const DEFAULT_FILENAME string = "aws.config"

const DEFAULT_TIMEOUT time.Duration = 2 * time.Minute
