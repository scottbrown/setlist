package setlist

// ListPermissionsRequired returns a slice of AWS IAM permission strings
// that are required for this application to function correctly.
// These permissions are needed to access AWS Organizations and SSO Admin APIs.
func ListPermissionsRequired() []string {
	return []string{
		"organizations:ListAccounts",
		"sso:ListInstances",
		"sso:ListPermissionSetsProvisionedToAccount",
		"sso:DescribePermissionSet",
	}
}
