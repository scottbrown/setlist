package main

var (
	ssoSession            string // SSO session nickname
	profile               string // AWS profile name
	ssoRegion             string // AWS region
	mapping               string // Mapping of account IDs to nicknames
	filename              string // Output filename
	stdout                bool   // Flag to print output to stdout instead of a file
	ssoFriendlyName       string // Optional friendly name for the SSO instance
	includeAccounts       string // Comma-delimited list of account IDs to include
	excludeAccounts       string // Comma-delimited list of account IDs to exclude
	includePermissionSets string // Comma-delimited list of permission set names to include
	excludePermissionSets string // Comma-delimited list of permission set names to exclude
	verbose               bool   // Flag to enable verbose logging
	logFormat             string // Log format: "plain" or "json"
	configFile            string // Path to YAML config file
)
