package setlist

import (
	"fmt"
)

// ConfigFile represents the structure of an AWS CLI configuration file,
// including session details, profiles, and nickname mappings. It contains
// all the information needed to generate a complete AWS config file for
// use with AWS SSO authentication.
type ConfigFile struct {
	SessionName     string            // Name of the SSO session
	IdentityStoreId IdentityStoreId   // The unique identity store ID
	FriendlyName    string            // Alt name used for the SSO instance
	Region          Region            // AWS region
	Profiles        []Profile         // List of AWS profiles
	NicknameMapping map[string]string // Mapping of account IDs to nicknames
}

// StartURL constructs the AWS SSO start URL based on the IdentityStoreId
// or FriendlyName.
func (c *ConfigFile) StartURL() string {
	subdomain := c.IdentityStoreId.String()

	if c.hasFriendlyName() {
		subdomain = c.FriendlyName
	}

	return fmt.Sprintf("https://%s.awsapps.com/start", subdomain)
}

// hasFriendlyName checks if a friendly name has been set for the SSO
// instance.
func (c *ConfigFile) hasFriendlyName() bool {
	return c.FriendlyName != ""
}

// HasNickname determines whether an account has a mapped nickname.
func (c ConfigFile) HasNickname(accountId string) bool {
	_, exists := c.NicknameMapping[accountId]

	return exists
}
