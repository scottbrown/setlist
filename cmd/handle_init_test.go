package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleInit_CreatesFile(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	path := filepath.Join(dir, ".setlist.yaml")
	configFile = path

	cmd := initCmd
	cmd.SetArgs([]string{})

	if err := handleInit(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "sso-session:") {
		t.Error("expected config file to contain sso-session key")
	}
	if !strings.Contains(content, "sso-region:") {
		t.Error("expected config file to contain sso-region key")
	}
	if !strings.Contains(content, "include-accounts:") {
		t.Error("expected config file to contain include-accounts key")
	}
	if !strings.Contains(content, "exclude-permission-sets:") {
		t.Error("expected config file to contain exclude-permission-sets key")
	}
}

func TestHandleInit_RefusesToOverwrite(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	path := filepath.Join(dir, ".setlist.yaml")
	configFile = path
	forceOverwrite = false

	if err := os.WriteFile(path, []byte("existing content"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := initCmd
	cmd.SetArgs([]string{})

	err := handleInit(cmd, nil)
	if err == nil {
		t.Fatal("expected error when file already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "already exists")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("error = %q, want it to mention --force", err.Error())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing content" {
		t.Error("existing file should not have been modified")
	}
}

func TestHandleInit_OverwriteWithForce(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	path := filepath.Join(dir, ".setlist.yaml")
	configFile = path
	forceOverwrite = true

	if err := os.WriteFile(path, []byte("existing content"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := initCmd
	cmd.SetArgs([]string{})

	if err := handleInit(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) == "existing content" {
		t.Error("file should have been overwritten")
	}
	if !strings.Contains(string(data), "sso-session:") {
		t.Error("overwritten file should contain template content")
	}
}

func TestHandleInit_CustomPathViaConfigFlag(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	customPath := filepath.Join(dir, "custom", "my-config.yaml")
	if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
		t.Fatal(err)
	}
	configFile = customPath

	cmd := initCmd
	cmd.SetArgs([]string{})

	if err := handleInit(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		t.Error("expected config file to be created at custom path")
	}
}

func TestHandleInit_DefaultPath(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	configFile = ""

	cmd := initCmd
	cmd.SetArgs([]string{})

	if err := handleInit(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(dir, ".setlist.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected config file at %s", expectedPath)
	}
}

func TestHandleInit_TemplateIsValidYAML(t *testing.T) {
	resetGlobals()

	dir := t.TempDir()
	path := filepath.Join(dir, ".setlist.yaml")
	configFile = path

	cmd := initCmd
	cmd.SetArgs([]string{})

	if err := handleInit(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the generated file can be loaded by the config loader
	testCmd := newTestCommand()
	configFile = path
	testCmd.Flags().Set(FlagConfig, path)

	if err := loadConfigFile(testCmd); err != nil {
		t.Fatalf("generated template is not valid config: %v", err)
	}
}
