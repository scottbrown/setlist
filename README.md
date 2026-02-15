![setlist](icon.png)

# SetList (formerly aws-config-creator)

Setlist is a CLI tool and Go library that automates the creation of AWS config files for organizations using AWS SSO. It parses AWS Organizations and Permission Sets to build a complete .aws/config file with all permission sets provisioned across your AWS member accounts.

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
1. `sso:ListPermissionSets`
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

### Verbose Logging

```bash
# Enable verbose output to see what setlist is doing
setlist --sso-session myorg --sso-region us-east-1 --verbose --stdout

# Use JSON log format for structured logging
setlist --sso-session myorg --sso-region us-east-1 --verbose --log-format json --stdout
```

### Listing Permission Sets

```bash
# List all permission sets available in the SSO instance
setlist --list-permission-sets --sso-region us-east-1
```

## Configuration File

Setlist supports a YAML configuration file as an alternative to specifying flags on every invocation. By default, it looks for `~/.setlist.yaml`. You can specify a custom path with `--config`.

**Precedence:** Command-line flags override config file values. Config file values override flag defaults.

### Example `~/.setlist.yaml`

```yaml
sso-session: myorg
sso-region: us-east-1
profile: admin
mapping: "123456789012=prod,210987654321=staging"
output: ~/.aws/config
stdout: false
sso-friendly-name: my-company
verbose: true
log-format: plain
```

All keys are optional — only specify the ones you want as defaults. Keys match flag names exactly (hyphenated).

### Usage with Config File

```bash
# Uses values from ~/.setlist.yaml
setlist --stdout

# Override a config file value with a flag
setlist --sso-session other-org --stdout

# Use a custom config file path
setlist --config /path/to/config.yaml --stdout
```

## Configuration Options

|Flag|Short|Description|Required|
|-|-|-|-|
|--sso-session|-s|Nickname for the SSO session (e.g., organization name)|Yes|
|--sso-region|-r|AWS region where AWS SSO resides|Yes|
|--profile|-p|AWS profile to use for authentication|No|
|--mapping|-m|Comma-delimited account nickname mapping (format: id=nickname)|No|
|--output|-o|Output file path (default: ./aws.config)|No|
|--stdout||Write config to stdout instead of a file|No|
|--sso-friendly-name||Alternative name for the SSO start URL|No|
|--list-accounts||List all available AWS accounts|No|
|--list-permission-sets||List all available permission sets in the SSO instance|No|
|--verbose|-v|Enable verbose logging output|No|
|--log-format||Log output format: "plain" (default) or "json"|No|
|--include-accounts||Comma-delimited list of account IDs to include|No|
|--exclude-accounts||Comma-delimited list of account IDs to exclude|No|
|--include-permission-sets||Comma-delimited list of permission set names to include|No|
|--exclude-permission-sets||Comma-delimited list of permission set names to exclude|No|
|--config|-c|Path to config file (default: ~/.setlist.yaml)|No|
|--check-update||Check if a newer version of the tool is available|No|
|--permissions||Print the required AWS permissions and exit|No|

## Library Usage

Setlist can be used as a Go library. The `setlist.Generate()` function provides a high-level API for generating config files programmatically:

```go
package main

import (
    "context"
    "fmt"

    "github.com/scottbrown/setlist"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/organizations"
    "github.com/aws/aws-sdk-go-v2/service/ssoadmin"
)

func main() {
    ctx := context.Background()
    cfg, _ := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))

    configFile, err := setlist.Generate(ctx, setlist.GenerateInput{
        SSOClient:   ssoadmin.NewFromConfig(cfg),
        OrgClient:   organizations.NewFromConfig(cfg),
        SessionName: "myorg",
        Region:      "us-east-1",
    })
    if err != nil {
        panic(err)
    }

    builder := setlist.NewFileBuilder(configFile)
    payload, _ := builder.Build()
    payload.WriteTo(fmt.Stdout)
}
```

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

## Lambda Deployment

Setlist can be deployed as an AWS Lambda function that generates config files on a schedule and uploads them to S3. This is useful for keeping a shared config file up to date automatically.

### Building the Lambda

```bash
task lambda-build
```

### SAM Template

A SAM template (`template.yaml`) is provided. Deploy with:

```bash
# Build the Lambda binary
task lambda-build

# Deploy with SAM
sam deploy --guided \
  --parameter-overrides \
    SSOSession=myorg \
    SSORegion=us-east-1 \
    S3Bucket=my-config-bucket \
    S3Key=aws.config
```

### Environment Variables

|Variable|Description|Required|
|-|-|-|
|SSO_SESSION|Nickname for the SSO session|Yes|
|SSO_REGION|AWS region where AWS SSO resides|Yes|
|S3_BUCKET|S3 bucket for the generated config file|Yes|
|S3_KEY|S3 object key for the config file|Yes|
|SSO_FRIENDLY_NAME|Alternative name for the SSO start URL|No|
|NICKNAME_MAPPING|Comma-delimited account nickname mapping|No|
|INCLUDE_ACCOUNTS|Comma-delimited list of account IDs to include|No|
|EXCLUDE_ACCOUNTS|Comma-delimited list of account IDs to exclude|No|
|INCLUDE_PERMISSION_SETS|Comma-delimited list of permission set names to include|No|
|EXCLUDE_PERMISSION_SETS|Comma-delimited list of permission set names to exclude|No|

### Required IAM Permissions

The Lambda execution role needs:

- `organizations:ListAccounts`
- `sso:ListInstances`
- `sso:ListPermissionSets`
- `sso:ListPermissionSetsProvisionedToAccount`
- `sso:DescribePermissionSet`
- `s3:PutObject` on the target S3 bucket/key

## Common Use Cases

### Team Onboarding

Simplify the onboarding process for new team members by providing a single command that generates a complete AWS config with all the accounts and permission sets they need access to.

### Credentials Refresh

When permission sets change in your AWS Organization, quickly regenerate your config file to include the new permission sets without manual configuration.

### CI/CD Pipelines

Use Setlist in CI/CD pipelines to ensure consistent AWS configuration across different environments.

### Automated Config Updates

Deploy as a Lambda function to automatically regenerate and upload shared config files to S3 on a schedule.

## Troubleshooting

### Permission Issues

If you encounter permission errors, ensure your AWS credentials have access to:

- organizations:ListAccounts
- sso:ListInstances
- sso:ListPermissionSets
- sso:ListPermissionSetsProvisionedToAccount
- sso:DescribePermissionSet

### Region Configuration

The --sso-region parameter must specify the region where your AWS SSO instance is deployed, not necessarily the region where your resources are located.

### Output File Permissions

Ensure you have write permissions to the output file location.

## Development

### Prerequisites

- Go 1.24 or newer
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

- `cmd/`: CLI entry point
- `cmd/lambda/`: Lambda function entry point
- `*.go`: Core library for AWS interactions and config generation
- `.github/workflows/`: CI/CD pipeline definitions
- `template.yaml`: SAM template for Lambda deployment
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
