// Package setlist provides utilities for managing AWS SSO configurations.
//
// It simplifies the process of generating AWS CLI configuration files for
// organizations using AWS SSO by parsing AWS Organizations and Permission Sets
// to build a complete .aws/config file with all permission sets provisioned
// across AWS member accounts.
//
// The package includes functionality for AWS API interactions, config file
// generation, nickname mapping for accounts, and structured error handling.
package setlist
