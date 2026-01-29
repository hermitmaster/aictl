package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllTools(t *testing.T) {
	tools := AllTools()
	if len(tools) != 5 {
		t.Errorf("AllTools() returned %d tools, expected 5", len(tools))
	}

	expected := map[Tool]bool{
		ToolWindsurf:   true,
		ToolCursor:     true,
		ToolAider:      true,
		ToolContinue:   true,
		ToolClaudeCode: true,
	}

	for _, tool := range tools {
		if !expected[tool] {
			t.Errorf("Unexpected tool: %s", tool)
		}
	}
}

func TestGetToolConfig(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name              string
		tool              Tool
		wantGlobalDir     string
		wantLocalDir      string
		wantRulesDir      string
		wantSupportsRules bool
		wantSupportsWorkflows bool
	}{
		{
			name:              "windsurf",
			tool:              ToolWindsurf,
			wantGlobalDir:     filepath.Join(home, ".codeium", "windsurf"),
			wantLocalDir:      ".windsurf",
			wantRulesDir:      "memories",
			wantSupportsRules: true,
			wantSupportsWorkflows: true,
		},
		{
			name:              "cursor",
			tool:              ToolCursor,
			wantGlobalDir:     filepath.Join(home, ".cursor"),
			wantLocalDir:      ".cursor",
			wantRulesDir:      "rules",
			wantSupportsRules: true,
			wantSupportsWorkflows: false,
		},
		{
			name:              "aider",
			tool:              ToolAider,
			wantGlobalDir:     filepath.Join(home, ".aider"),
			wantLocalDir:      ".aider",
			wantRulesDir:      "conventions",
			wantSupportsRules: true,
			wantSupportsWorkflows: false,
		},
		{
			name:              "continue",
			tool:              ToolContinue,
			wantGlobalDir:     filepath.Join(home, ".continue"),
			wantLocalDir:      ".continue",
			wantRulesDir:      "rules",
			wantSupportsRules: true,
			wantSupportsWorkflows: false,
		},
		{
			name:              "claude-code",
			tool:              ToolClaudeCode,
			wantGlobalDir:     filepath.Join(home, ".claude"),
			wantLocalDir:      ".claude",
			wantRulesDir:      "",
			wantSupportsRules: true,
			wantSupportsWorkflows: true,
		},
		{
			name:              "unknown tool",
			tool:              Tool("unknown"),
			wantGlobalDir:     "",
			wantLocalDir:      "",
			wantRulesDir:      "",
			wantSupportsRules: false,
			wantSupportsWorkflows: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetToolConfig(tt.tool)

			if cfg.GlobalDir != tt.wantGlobalDir {
				t.Errorf("GlobalDir = %v, want %v", cfg.GlobalDir, tt.wantGlobalDir)
			}
			if cfg.LocalDir != tt.wantLocalDir {
				t.Errorf("LocalDir = %v, want %v", cfg.LocalDir, tt.wantLocalDir)
			}
			if cfg.RulesDir != tt.wantRulesDir {
				t.Errorf("RulesDir = %v, want %v", cfg.RulesDir, tt.wantRulesDir)
			}
			if cfg.SupportsRules != tt.wantSupportsRules {
				t.Errorf("SupportsRules = %v, want %v", cfg.SupportsRules, tt.wantSupportsRules)
			}
			if cfg.SupportsWorkflows != tt.wantSupportsWorkflows {
				t.Errorf("SupportsWorkflows = %v, want %v", cfg.SupportsWorkflows, tt.wantSupportsWorkflows)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	home, _ := os.UserHomeDir()
	cfg := DefaultConfig()

	expectedStateDir := filepath.Join(home, ".config", "aictl")
	expectedCacheDir := filepath.Join(home, ".cache", "aictl")

	if cfg.StateDir != expectedStateDir {
		t.Errorf("StateDir = %v, want %v", cfg.StateDir, expectedStateDir)
	}
	if cfg.CacheDir != expectedCacheDir {
		t.Errorf("CacheDir = %v, want %v", cfg.CacheDir, expectedCacheDir)
	}
	if len(cfg.Tools) != 0 {
		t.Errorf("Tools should be empty by default, got %v", cfg.Tools)
	}
}

func TestConfigGetStateDir(t *testing.T) {
	// Use a temp directory for testing
	tmpDir := t.TempDir()
	cfg := &Config{
		StateDir: filepath.Join(tmpDir, "state"),
	}

	dir, err := cfg.GetStateDir()
	if err != nil {
		t.Fatalf("GetStateDir() error = %v", err)
	}

	if dir != cfg.StateDir {
		t.Errorf("GetStateDir() = %v, want %v", dir, cfg.StateDir)
	}

	// Verify directory was created
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("GetStateDir() should create the directory")
	}
}

func TestConfigGetCacheDir(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		CacheDir: filepath.Join(tmpDir, "cache"),
	}

	dir, err := cfg.GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir() error = %v", err)
	}

	if dir != cfg.CacheDir {
		t.Errorf("GetCacheDir() = %v, want %v", dir, cfg.CacheDir)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("GetCacheDir() should create the directory")
	}
}

func TestDetectInstalledTools(t *testing.T) {
	// This test depends on the actual system state
	// Just verify it returns a slice without error
	tools := DetectInstalledTools()
	if tools == nil {
		t.Error("DetectInstalledTools() should not return nil")
	}
}
