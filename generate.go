package setlist

import (
	"context"
	"fmt"
	"log/slog"

	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
)

// GenerateInput holds all the parameters needed to generate an AWS config file.
type GenerateInput struct {
	SSOClient             SSOAdminClient
	OrgClient             OrganizationsClient
	SessionName           string
	Region                string
	FriendlyName          string
	NicknameMapping       string
	IncludeAccounts       string
	ExcludeAccounts       string
	IncludePermissionSets string
	ExcludePermissionSets string
}

// Generate orchestrates the full config file generation workflow. It retrieves
// the SSO instance, lists and filters accounts, gathers permission sets, and
// assembles a ConfigFile ready for output.
func Generate(ctx context.Context, input GenerateInput) (ConfigFile, error) {
	slog.Info("Retrieving SSO instance")
	instance, err := SsoInstance(ctx, input.SSOClient)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("failed to retrieve SSO instance: %w", err)
	}
	slog.Info("SSO instance retrieved", "instance_arn", *instance.InstanceArn)

	slog.Info("Listing AWS accounts")
	accounts, err := ListAccounts(ctx, input.OrgClient)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("failed to list AWS accounts: %w", err)
	}
	slog.Info("AWS accounts retrieved", "count", len(accounts))

	includeList, err := ParseAccountIdList(input.IncludeAccounts)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("invalid include-accounts: %w", err)
	}

	excludeList, err := ParseAccountIdList(input.ExcludeAccounts)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("invalid exclude-accounts: %w", err)
	}

	beforeCount := len(accounts)
	accounts, err = FilterAccounts(accounts, includeList, excludeList)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("account filter error: %w", err)
	}
	slog.Info("Accounts filtered", "before", beforeCount, "after", len(accounts))

	nicknameMapping, err := ParseNicknameMapping(input.NicknameMapping)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("invalid mapping format: %w", err)
	}

	identityStoreId, err := NewIdentityStoreId(*instance.IdentityStoreId)
	if err != nil {
		return ConfigFile{}, err
	}

	region, err := NewRegion(input.Region)
	if err != nil {
		return ConfigFile{}, err
	}

	includePSList, err := ParsePermissionSetList(input.IncludePermissionSets)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("invalid include-permission-sets: %w", err)
	}

	excludePSList, err := ParsePermissionSetList(input.ExcludePermissionSets)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("invalid exclude-permission-sets: %w", err)
	}

	configFile := ConfigFile{
		SessionName:     input.SessionName,
		IdentityStoreId: identityStoreId,
		FriendlyName:    input.FriendlyName,
		Region:          region,
		NicknameMapping: nicknameMapping,
	}

	profiles, err := generateProfiles(ctx, input.SSOClient, instance, accounts, input.SessionName, includePSList, excludePSList)
	if err != nil {
		return configFile, err
	}

	configFile.Profiles = profiles
	return configFile, nil
}

func generateProfiles(
	ctx context.Context,
	ssoClient SSOAdminClient,
	instance ssotypes.InstanceMetadata,
	accounts []orgtypes.Account,
	sessionName string,
	includePS, excludePS []string,
) ([]Profile, error) {
	var profiles []Profile

	for _, account := range accounts {
		if account.Id == nil {
			slog.Warn("Found account with nil ID, skipping")
			continue
		}

		slog.Info("Processing account", "account_id", *account.Id)
		permissionSets, err := PermissionSets(ctx, ssoClient, *instance.InstanceArn, *account.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to list permission sets for account %s: %w", *account.Id, err)
		}

		permissionSets, err = FilterPermissionSets(permissionSets, includePS, excludePS)
		if err != nil {
			return nil, fmt.Errorf("permission set filter error: %w", err)
		}

		slog.Info("Permission sets retrieved", "account_id", *account.Id, "count", len(permissionSets))

		for _, p := range permissionSets {
			if p.Name == nil || p.Description == nil || p.SessionDuration == nil {
				slog.Warn("Found incomplete permission set data, skipping", "account_id", *account.Id)
				continue
			}

			profileDesc, err := NewProfileDescription(*p.Description)
			if err != nil {
				slog.Warn("Invalid profile description", "error", err.Error())
				continue
			}
			sessionDuration, err := NewSessionDuration(*p.SessionDuration)
			if err != nil {
				slog.Warn("Invalid session duration", "error", err.Error())
				continue
			}
			sName, err := NewSessionName(sessionName)
			if err != nil {
				slog.Warn("Invalid session name", "error", err.Error())
				continue
			}
			accountId, err := NewAWSAccountId(*account.Id)
			if err != nil {
				slog.Warn("Invalid AWS account ID", "error", err.Error())
				continue
			}
			roleName, err := NewRoleName(*p.Name)
			if err != nil {
				slog.Warn("Invalid role name", "error", err.Error())
				continue
			}

			profiles = append(profiles, Profile{
				Description:     profileDesc,
				SessionDuration: sessionDuration,
				SessionName:     sName,
				AccountId:       accountId,
				RoleName:        roleName,
			})
		}
	}

	return profiles, nil
}
