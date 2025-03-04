package setlist

import (
	"testing"
)

// BenchmarkCompareVersions benchmarks the performance of comparing different
// version string combinations
func BenchmarkCompareVersions(b *testing.B) {
	benchCases := []struct {
		name           string
		currentVersion string
		latestVersion  string
	}{
		{"Simple_Equal", "1.0.0", "1.0.0"},
		{"Simple_Newer", "1.0.0", "1.1.0"},
		{"Simple_Older", "1.1.0", "1.0.0"},
		{"Complex_Equal", "1.23.456", "1.23.456"},
		{"Complex_Newer", "1.23.456", "1.24.1"},
		{"Complex_Older", "2.0.0", "1.99.999"},
		{"DifferentLength_Shorter", "1.2", "1.2.3"},
		{"DifferentLength_Longer", "1.2.3.4", "1.2.3"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			// Reset the timer for each test case
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				_ = compareVersions(bc.currentVersion, bc.latestVersion)
			}
		})
	}
}

// BenchmarkCompareVersionsWithPrefixes benchmarks version comparison
// with and without 'v' prefixes to evaluate any performance differences
func BenchmarkCompareVersionsWithPrefixes(b *testing.B) {
	benchCases := []struct {
		name           string
		currentVersion string
		latestVersion  string
	}{
		{"NoPrefixes", "1.0.0", "1.1.0"},
		{"CurrentWithPrefix", "v1.0.0", "1.1.0"},
		{"LatestWithPrefix", "1.0.0", "v1.1.0"},
		{"BothWithPrefix", "v1.0.0", "v1.1.0"},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			// Prepare the versions by removing 'v' prefixes as done in the actual code
			current := bc.currentVersion
			latest := bc.latestVersion

			if len(current) > 0 && current[0] == 'v' {
				current = current[1:]
			}
			if len(latest) > 0 && latest[0] == 'v' {
				latest = latest[1:]
			}

			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				_ = compareVersions(current, latest)
			}
		})
	}
}
