package main

import (
	"fmt"

	"github.com/scottbrown/setlist"
)

// handleListPermissions lists all required permissions
func handleListPermissions() error {
	for _, p := range setlist.ListPermissionsRequired() {
		fmt.Println(p)
	}
	return nil
}
