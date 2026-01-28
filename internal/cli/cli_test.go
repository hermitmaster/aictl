package cli

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
}

func TestRootCmdExists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if rootCmd.Use != "aictl" {
		t.Errorf("rootCmd.Use = %v, want aictl", rootCmd.Use)
	}
}

func TestInstallCmdExists(t *testing.T) {
	if installCmd == nil {
		t.Fatal("installCmd should not be nil")
	}

	if installCmd.Use != "install <resource>" {
		t.Errorf("installCmd.Use = %v, want 'install <resource>'", installCmd.Use)
	}
}

func TestUninstallCmdExists(t *testing.T) {
	if uninstallCmd == nil {
		t.Fatal("uninstallCmd should not be nil")
	}

	if uninstallCmd.Use != "uninstall <resource>" {
		t.Errorf("uninstallCmd.Use = %v, want 'uninstall <resource>'", uninstallCmd.Use)
	}
}

func TestListCmdExists(t *testing.T) {
	if listCmd == nil {
		t.Fatal("listCmd should not be nil")
	}

	if listCmd.Use != "list" {
		t.Errorf("listCmd.Use = %v, want 'list'", listCmd.Use)
	}
}

func TestGetTargetToolsWithFlag(t *testing.T) {
	// Save and restore original flag
	origToolFlag := toolFlag
	defer func() { toolFlag = origToolFlag }()

	toolFlag = "windsurf"
	tools, err := getTargetTools(nil)
	if err != nil {
		t.Fatalf("getTargetTools() error = %v", err)
	}

	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	if len(tools) > 0 && string(tools[0]) != "windsurf" {
		t.Errorf("Expected windsurf, got %s", tools[0])
	}
}

func TestGetTargetToolsMultiple(t *testing.T) {
	origToolFlag := toolFlag
	defer func() { toolFlag = origToolFlag }()

	toolFlag = "windsurf,cursor"
	tools, err := getTargetTools(nil)
	if err != nil {
		t.Fatalf("getTargetTools() error = %v", err)
	}

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}
