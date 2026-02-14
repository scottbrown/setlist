package setlist

import (
	"errors"
	"fmt"
	"strings"

	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

var ErrMutuallyExclusiveFilters = errors.New("include and exclude account filters are mutually exclusive")

// ParseAccountIdList parses a comma-delimited string of AWS account IDs
// into a slice of AWSAccountId. Each ID is validated against the expected
// 12-digit format.
func ParseAccountIdList(s string) ([]AWSAccountId, error) {
	if len(s) == 0 {
		return nil, nil
	}

	tokens := strings.Split(s, ",")
	var result []AWSAccountId

	for i, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if !AccountIdPattern.MatchString(token) {
			return nil, fmt.Errorf("invalid account ID at position %d: %q", i+1, token)
		}

		result = append(result, AWSAccountId(token))
	}

	return result, nil
}

// FilterAccounts filters a list of AWS accounts based on include and exclude
// lists. If include is non-empty, only accounts in the include list are
// returned. If exclude is non-empty, accounts in the exclude list are removed.
// Setting both include and exclude is an error.
func FilterAccounts(accounts []orgtypes.Account, include, exclude []AWSAccountId) ([]orgtypes.Account, error) {
	if len(include) > 0 && len(exclude) > 0 {
		return nil, ErrMutuallyExclusiveFilters
	}

	if len(include) == 0 && len(exclude) == 0 {
		return accounts, nil
	}

	if len(include) > 0 {
		includeSet := make(map[AWSAccountId]bool, len(include))
		for _, id := range include {
			includeSet[id] = true
		}

		var filtered []orgtypes.Account
		for _, a := range accounts {
			if a.Id != nil && includeSet[AWSAccountId(*a.Id)] {
				filtered = append(filtered, a)
			}
		}
		return filtered, nil
	}

	excludeSet := make(map[AWSAccountId]bool, len(exclude))
	for _, id := range exclude {
		excludeSet[id] = true
	}

	var filtered []orgtypes.Account
	for _, a := range accounts {
		if a.Id != nil && !excludeSet[AWSAccountId(*a.Id)] {
			filtered = append(filtered, a)
		}
	}
	return filtered, nil
}
