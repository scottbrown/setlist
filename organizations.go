package setlist

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

// Define interface for the Organizations client to make testing easier
type OrganizationsClient interface {
	ListAccounts(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error)
}

// ListAccounts retrieves all accounts within an AWS Organization using
// the provided Organizations client. It handles pagination automatically
// to ensure all accounts are retrieved, even when the organization contains
// a large number of accounts. The function respects context cancellation for
// proper timeout handling.
func ListAccounts(ctx context.Context, client OrganizationsClient) ([]types.Account, error) {
	var accounts []types.Account

	var token *string
	for {
		// Check if context is cancelled before making the API call
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		orgOutput, err := client.ListAccounts(ctx, &organizations.ListAccountsInput{NextToken: token})
		if err != nil {
			return accounts, fmt.Errorf("failed to list AWS accounts: %w", err)
		}

		for _, v := range orgOutput.Accounts {
			accounts = append(accounts, v)
		}

		if orgOutput.NextToken == nil {
			break
		}

		token = orgOutput.NextToken
	}

	return accounts, nil
}
