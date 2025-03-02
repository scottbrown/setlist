package setlist

import (
	"fmt"
	"strings"
)

// ParseNicknameMapping parses a comma-delimited string of account ID to nickname
// mappings into a map. The expected format is "accountID1=nickname1,accountID2=nickname2".
// This enables users to reference AWS accounts by friendly names rather than
// numeric IDs. The function validates the format and returns appropriate errors
// for malformed input.
func ParseNicknameMapping(mapping string) (map[string]string, error) {
	nicknameMapping := make(map[string]string)

	if len(mapping) == 0 {
		return nicknameMapping, nil
	}

	tokens := strings.Split(mapping, ",")
	for i, token := range tokens {
		if token == "" {
			continue // skip empty token
		}

		parts := strings.Split(token, "=")
		if len(parts) != 2 {
			return nicknameMapping, fmt.Errorf("invalid nickname mapping format at entry %d: %q, expected format 'accountID=nickname'", i+1, token)
		}

		accountID := strings.TrimSpace(parts[0])
		nickname := strings.TrimSpace(parts[1])

		if accountID == "" {
			return nil, fmt.Errorf("empty account ID in mapping entry %d", i+1)
		}

		if nickname == "" {
			return nil, fmt.Errorf("empty nickname in mapping entry %d", i+1)
		}

		nicknameMapping[accountID] = nickname
	}

	return nicknameMapping, nil
}
