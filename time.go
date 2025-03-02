package setlist

import (
	"time"
)

// generateTimestamp returns the current UTC timestamp formatted as
// "2006-01-02T15:04:05 MST" for use in config file generation.
// This helps indicate when the configuration was last generated.
func generateTimestamp() string {
	now := time.Now().UTC()

	return now.Format("2006-01-02T15:04:05 MST")
}
