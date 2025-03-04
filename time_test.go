package setlist

import (
	"regexp"
	"testing"
	"time"
)

// TestGenerateTimestamp verifies the timestamp format
func TestGenerateTimestamp(t *testing.T) {
	// The expected format is "2006-01-02T15:04:05 UTC"
	expectedPattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2} UTC$`

	result := generateTimestamp()

	matched, err := regexp.MatchString(expectedPattern, result)
	if err != nil {
		t.Fatalf("Error matching regex pattern: %v", err)
	}

	if !matched {
		t.Errorf("generateTimestamp() = %q, does not match expected pattern %q",
			result, expectedPattern)
	}
}

// TestTimestampIsUTC verifies that the timestamp is in UTC
func TestTimestampIsUTC(t *testing.T) {
	// Parse the generated timestamp
	result := generateTimestamp()

	// The timestamp should end with " UTC"
	if len(result) < 4 || result[len(result)-3:] != "UTC" {
		t.Errorf("generateTimestamp() = %q, should end with 'UTC'", result)
	}

	// The actual time should be within a reasonable range of current time
	timestamp, err := time.Parse("2006-01-02T15:04:05 MST", result)
	if err != nil {
		t.Fatalf("Failed to parse timestamp %q: %v", result, err)
	}

	// Check that the timestamp is within 5 seconds of the current time
	now := time.Now().UTC()
	diff := now.Sub(timestamp)
	if diff < 0 {
		diff = -diff
	}

	if diff > 5*time.Second {
		t.Errorf("Timestamp difference too large: %v", diff)
	}
}
