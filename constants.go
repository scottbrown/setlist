// Default AWS SSO configuration constants used for generating
// properly formatted config files with standard attributes.
package setlist

// SSOSessionSectionKey is the key used for defining an SSO session in
// configuration files.
const SSOSessionSectionKey string = "sso-session"

// SSOSessionAttrKey is the attribute key for storing the SSO session name.
const SSOSessionAttrKey string = "sso_session"

// SSOStartUrlKey is the attribute key for the AWS SSO start URL.
const SSOStartUrlKey string = "sso_start_url"

// SSORegionKey is the attribute key for specifying the AWS region.
const SSORegionKey string = "sso_region"

// SSORegistrationScopesKey defines the key for SSO registration scopes.
const SSORegistrationScopesKey string = "sso_registration_scopes"

// SSORegistrationScopesValue defines the default value for SSO
// registration scopes.
const SSORegistrationScopesValue string = "sso:account:access"

// SSOAccountIdKey is the key used for specifying the AWS account ID in a
// profile.
const SSOAccountIdKey string = "sso_account_id"

// SSORoleNameKey is the key used for defining the IAM role name.
const SSORoleNameKey string = "sso_role_name"

// DefaultNicknamePrefix defines the prefix used for accounts without
// explicit nicknames.
const DefaultNicknamePrefix string = "NoNickname"
