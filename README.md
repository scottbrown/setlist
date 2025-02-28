![setlist](icon.png)

# SetList (formerly aws-config-creator)

Setlist is a CLI tool that automates the creation of AWS config files for organizations using AWS SSO. It parses AWS Organizations and Permission Sets to build a complete .aws/config file with all permission sets provisioned across your AWS member accounts.

## Why Setlist?
Managing AWS credentials across multiple AWS accounts with SSO can be challenging. While AWS provides the aws sso configure command, it's tedious to use when you have:

- Multiple AWS accounts in your organization
- Multiple permission sets per account
- Teams that need consistent configuration

Setlist solves this by automatically generating a complete .aws/config file with profiles for all accounts and permission sets you have access to, saving you time and preventing configuration errors.

## Installation

<!--
### Homebrew

```bash
brew tap scottbrown/aws-tools
brew install scottbrown/setlist
```
-->

### Download Binary

Download the latest release from the [GitHub Releases page](https://github.com/scottbrown/setlist/releases).

### Build from Source

```bash
# Clone the repository
git clone https://github.com/scottbrown/setlist.git
cd setlist

# Install Task (if not already installed)
brew install go-task

# Build the binary
task build

# The binary will be available at .build/setlist
```

## Usage

### Prerequisites

1. AWS CLI installed
1. AWS SSO configured for your organization
1. AWS credentials with permissions to access AWS Organizations and SSO Admin APIs

## Permissions Required

This tool requires some readonly permissions from your AWS organization account.  They are:

1. `organizations:ListAccounts`
1. `sso:ListInstances`
1. `sso:ListPermissionSetsProvisionedToAccount`
1. `sso:DescribePermissionSet`

You can view these in the application by running:

```bash
setlist --permissions
```

### Basic Usage

```bash
# Generate AWS config using the organization's SSO instance
setlist --sso-session myorg --sso-region us-east-1 --output ~/.aws/config
```

### Using an AWS Profile

```bash
# Use an existing AWS profile to authenticate
setlist --sso-session myorg --sso-region us-east-1 --profile admin --output ~/.aws/config
```

### Using Account Nicknames

```bash
# Map account IDs to friendly names
setlist --sso-session myorg --sso-region us-east-1 \
  --mapping "123456789012=prod,210987654321=staging" \
  --output ~/.aws/config
```

### Print to stdout

```bash
# Output to stdout instead of a file
setlist --sso-session myorg --sso-region us-east-1 --stdout
```

### Using a Friendly Name for SSO URL

```bash
# Use a friendly name instead of the identity store ID
setlist --sso-session myorg \
  --sso-region us-east-1 \
  --sso-friendly-name my-company \
  --output ~/.aws/config
```

By supplying a `--mapping` flag with a comma-delimited list of key=value pairs corresponding to AWS Account ID and its nickname, the tool will create the basic `.aws/config` profiles and then create a separate set of profiles that follow the format `[profile NICKNAME-PERMISSIONSETNAME]`.  For example: `[profile acme-AdministratorAccess]`.  This removes the need for your users to remember the 12-digit AWS Account ID, but also allows for backward-compatibility for those people that like using the AWS Account ID in the profile name.

## Configuration Options

|Flag|Short|Description|Required|
|-|-|-|-|
|--sso-session|-s|Nickname for the SSO session (e.g., organization name)|Yes
|--sso-region|-r|AWS region where AWS SSO resides|Yes|
|--profile|-p|AWS profile to use for authentication|No|
|--mapping|-m|Comma-delimited account nickname mapping (format: id=nickname)|No|
|--output|-o|Output file path (default: ./aws.config)|No|
|--stdout||Write config to stdout instead of a file|No|
|--sso-friendly-name||Alternative name for the SSO start URL|No|

## Generated Config Format

Setlist generates an AWS config file with:

1. A default section specifying the SSO session
1. An SSO session section with start URL, region, and registration scopes
1. Profile sections for each account and permission set combination

Example:

```ini
[default]
# Generated on: 2025-02-27T10:15:30 UTC
sso_session = myorg

[sso-session myorg]
sso_start_url = https://d-12345abcde.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access

# Administrator access. Session Duration: PT12H
[profile 123456789012-AdministratorAccess]
sso_session = myorg
sso_account_id = 123456789012
sso_role_name = AdministratorAccess

# Administrator access. Session Duration: PT12H
[profile prod-AdministratorAccess]
sso_session = myorg
sso_account_id = 123456789012
sso_role_name = AdministratorAccess
```

## Common Use Cases

### Team Onboarding

Simplify the onboarding process for new team members by providing a single command that generates a complete AWS config with all the accounts and permission sets they need access to.

### Credentials Refresh

When permission sets change in your AWS Organization, quickly regenerate your config file to include the new permission sets without manual configuration.

### CI/CD Pipelines

Use Setlist in CI/CD pipelines to ensure consistent AWS configuration across different environments.

## Troubleshooting

### Permission Issues

If you encounter permission errors, ensure your AWS credentials have access to:

- organizations:ListAccounts
- sso-admin:ListInstances
- sso-admin:ListPermissionSetsProvisionedToAccount
- sso-admin:DescribePermissionSet

### Region Configuration

The --sso-region parameter must specify the region where your AWS SSO instance is deployed, not necessarily the region where your resources are located.

### Output File Permissions

Ensure you have write permissions to the output file location.

## Development

### Prerequisites

- Go 1.21 or newer
- [Task](https://go-task.dev)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/scottbrown/setlist.git
cd setlist

# Run tests
task test

# Build for development
task build
```

### Running Tests

```bash
# Run all tests
task test

# Run security checks
task check
```

### Building Releases

```bash
# Build a release version
task release VERSION=v1.2.3
```

### Project Structure

- `cmd/`: Command-line interface code
- `*.go`: Core functionality for AWS interactions and config generation
- `.github/workflows/`: CI/CD pipeline definitions
- `go.mod, go.sum`: Go module definitions
- `taskfile.yml`: Task automation definitions

## Contributing

Contributions are welcome! Here's how to contribute:

1. Fork the repository
1. Create a feature branch: `git checkout -b my-new-feature`
1. Make your changes and add tests if applicable
1. Run tests to ensure everything passes: `task test`
1. Commit your changes: `git commit -am 'Add some feature'`
1. Push to the branch: `git push origin my-new-feature`
1. Submit a pull request

## Releases

Each release comes with a software bill of materials (SBOM).  It is
generated using [CycloneDX-GoMod](https://github.com/CycloneDX/cyclonedx-gomod) using the following command:

```bash
cyclonedx-gomod mod -licenses -json -output bom.json
```

Releases are typically automated via Github Actions whenever a new tag is
pushed to the default branch.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all contributors who have helped improve Setlist
- Built with Cobra and AWS SDK for Go v2
