# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SetList is a Go CLI tool that automates AWS configuration file generation for organizations using AWS SSO. It parses AWS Organizations and Permission Sets to build complete `.aws/config` files with all permission sets provisioned across member accounts.

## Development Commands

### Task Runner
This project uses [go-task](https://go-task.dev) for build automation. All commands use `task`:

```bash
# Core development
task build          # Build local binary (→ .build/setlist)
task test           # Run unit tests  
task fmt            # Format code
task check          # Run security scans (SAST, vet, vulnerability checks)
task bench          # Run benchmarks
task clean          # Clean build directories

# Release management
task release VERSION=x.y.z    # Create release artifacts for all platforms
task sbom                     # Generate software bill of materials
```

### Prerequisites
- Go 1.24+ (required)
- Task runner: `brew install go-task`
- For security scans: `gosec`, `govulncheck`
- For SBOM: `cyclonedx-gomod`

## Architecture

### Core Package Structure
- **Module:** `github.com/scottbrown/setlist`
- **Entry Point:** `/cmd/main.go` (Cobra CLI)
- **Business Logic:** Root directory Go files
- **Dependencies:** AWS SDK v2, Cobra, go-ini

### Key Components

**CLI Layer (`/cmd/`):**
- `main.go` - Application entry point
- `rootCmd.go` - Main CLI logic and command handling
- `constants.go`, `flags.go` - CLI configuration

**Core Business Logic:**
- `config_file.go` - ConfigFile struct and SSO URL generation
- `file_builder.go` - INI file generation for AWS config
- `profile.go` - Profile data structures with validation
- `types.go` - Custom types with validation (IdentityStoreId, Region, etc.)
- `organizations.go` - AWS Organizations API client
- `ssoadmin.go` - AWS SSO Admin API client
- `permissions.go` - Permission set handling
- `nickname.go` - Account nickname mapping logic

**AWS Integration:**
- Uses AWS SDK v2 with context handling and pagination
- Required permissions: `organizations:ListAccounts`, `sso:ListInstances`, `sso:ListPermissionSetsProvisionedToAccount`, `sso:DescribePermissionSet`
- Interface-based clients for testability

### Type Safety Patterns
The codebase uses extensive custom types with validation:
- Constructor functions validate input and return errors
- String methods for serialization
- Examples: `AWSAccountId`, `IdentityStoreId`, `SessionDuration`

### Testing Strategy
- Comprehensive unit tests for all components (`*_test.go`)
- Benchmark tests for performance-critical code (`*_benchmark_test.go`)
- Table-driven test patterns throughout
- Interface-based design enables thorough testing

## Code Conventions

### Error Handling
- Proper error wrapping with context
- Validation at type constructors
- Context-aware operations with timeouts

### AWS SDK Patterns
- Use AWS SDK v2 interfaces for testability
- Handle pagination for large result sets
- Proper context cancellation and timeouts
- Error wrapping preserves AWS error details

### File Generation
- Uses go-ini library for INI file handling
- Template-based profile generation
- Automatic timestamp generation in output

## CI/CD Pipeline

### GitHub Actions Workflows
- `ci.yaml` - Test and build on every push
- `release.yaml` - Automated releases on tags
- `sast.yaml` - Security analysis
- `automerge-dependabot.yaml` - Dependency management

### Security and Quality
- Multiple security scanners: gosec, go vet, govulncheck
- Dependabot for automated dependency updates
- SBOM generation using CycloneDX-GoMod
- 5-minute test timeout in CI

### Release Process
- Multi-platform builds (Linux, Windows, macOS - amd64/arm64)
- Automated GitHub releases with binaries
- Software Bill of Materials included
- Semantic versioning