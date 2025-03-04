package setlist

import (
	"strings"
	"testing"
)

// BenchmarkParseNicknameMapping benchmarks the parsing of nickname mappings
// with different sizes and complexities
func BenchmarkParseNicknameMapping(b *testing.B) {
	benchCases := []struct {
		name    string
		mapping string
	}{
		{"Empty", ""},
		{"Small_5Mappings", generateMapping(5)},
		{"Medium_50Mappings", generateMapping(50)},
		{"Large_500Mappings", generateMapping(500)},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			// Reset the timer for each test case
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				_, err := ParseNicknameMapping(bc.mapping)
				if err != nil {
					b.Fatalf("ParseNicknameMapping failed: %v", err)
				}
			}
		})
	}
}

// generateMapping creates a string with the specified number of nickname mappings
func generateMapping(count int) string {
	mappings := make([]string, count)

	// Create a slice of account IDs and nicknames to use
	accountIDs := []string{
		"111122223333", "444455556666", "777788889999", "999900001111", "222233334444",
		"555566667777", "888899990000", "123412341234", "432143214321", "654365436543",
	}

	nicknames := []string{
		"dev", "prod", "stage", "test", "qa",
		"sandbox", "training", "demo", "temp", "archive",
	}

	for i := 0; i < count; i++ {
		// Use modulo to cycle through the available IDs and nicknames
		accountIdx := i % len(accountIDs)
		nicknameIdx := i % len(nicknames)

		// Format as "accountID=nickname"
		mappings[i] = accountIDs[accountIdx] + "=" + nicknames[nicknameIdx]
	}

	return strings.Join(mappings, ",")
}
