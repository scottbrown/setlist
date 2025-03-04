package setlist

import (
	"testing"
)

// BenchmarkFileBuilder benchmarks the performance of the Build function
// with different numbers of profiles to evaluate scalability
func BenchmarkFileBuilder(b *testing.B) {
	// Define benchmark cases with different numbers of profiles
	benchCases := []struct {
		name        string
		numProfiles int
	}{
		{"Small_10Profiles", 10},
		{"Medium_100Profiles", 100},
		{"Large_1000Profiles", 1000},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			// Set up a config with the specified number of profiles
			configFile := createBenchmarkConfig(bc.numProfiles)
			builder := NewFileBuilder(configFile)

			// Reset the timer to exclude setup time
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				_, err := builder.Build()
				if err != nil {
					b.Fatalf("Build failed: %v", err)
				}
			}
		})
	}
}

// createBenchmarkConfig creates a ConfigFile with the specified number of profiles
func createBenchmarkConfig(numProfiles int) ConfigFile {
	profiles := make([]Profile, numProfiles)

	// Create dummy profiles
	for i := 0; i < numProfiles; i++ {
		profiles[i] = Profile{
			Name:            "profile-name",
			Description:     "Test Profile",
			SessionDuration: "PT12H",
			SessionName:     "test-session",
			AccountId:       "123456789012",
			RoleName:        "TestRole",
		}
	}

	// Create nickname mapping
	nicknameMapping := make(map[string]string)
	nicknameMapping["123456789012"] = "test-account"

	return ConfigFile{
		SessionName:     "test-session",
		IdentityStoreId: "d-12345abcde",
		Region:          "us-east-1",
		Profiles:        profiles,
		NicknameMapping: nicknameMapping,
	}
}
