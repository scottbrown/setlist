package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHandleListPermissions(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := handleListPermissions()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := []string{
		"organizations:ListAccounts",
		"sso:ListInstances",
		"sso:ListPermissionSets",
		"sso:ListPermissionSetsProvisionedToAccount",
		"sso:DescribePermissionSet",
	}

	for _, perm := range expected {
		if !strings.Contains(output, perm) {
			t.Errorf("Expected output to contain %q, got: %s", perm, output)
		}
	}
}
