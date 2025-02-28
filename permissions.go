package setlist

func ListPermissionsRequired() []string {
	return []string{
		"organizations:ListAccounts",
		"sso:ListInstances",
		"sso:ListPermissionSetsProvisionedToAccount",
		"sso:DescribePermissionSet",
	}
}
