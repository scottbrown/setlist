package setlist

import (
	"testing"
)

func TestListPermissions(t *testing.T) {
	result := ListPermissionsRequired()

	if len(result) == 0 {
		t.Errorf("ListPermissionsRequired() must have at least one permission")
	}
}
