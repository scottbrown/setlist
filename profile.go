package setlist

// Profile represents an AWS SSO profile configuration.
// It stores information about a permission set, including its metadata
// and the account it belongs to. This is used to generate AWS CLI profile
// configurations in the output file.
type Profile struct {
	Name            string
	Description     string
	SessionDuration string
	SessionName     string
	AccountId       string
	RoleName        string
}
