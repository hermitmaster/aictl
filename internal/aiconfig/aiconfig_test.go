package aiconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".aiconfig")

	content := `tools:
  - windsurf
  - cursor

taps:
  - name: mycompany
    url: https://github.com/mycompany/resources

install:
  - name: bundled/typescript-rules
  - name: bundled/jira-context
    tools: [windsurf]

scope: local
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Tools) != 2 {
		t.Errorf("Tools count = %d, want 2", len(cfg.Tools))
	}
	if cfg.Tools[0] != "windsurf" {
		t.Errorf("Tools[0] = %s, want windsurf", cfg.Tools[0])
	}

	if len(cfg.Taps) != 1 {
		t.Errorf("Taps count = %d, want 1", len(cfg.Taps))
	}
	if cfg.Taps[0].Name != "mycompany" {
		t.Errorf("Taps[0].Name = %s, want mycompany", cfg.Taps[0].Name)
	}

	if len(cfg.Install) != 2 {
		t.Errorf("Install count = %d, want 2", len(cfg.Install))
	}
	if cfg.Install[0].Name != "bundled/typescript-rules" {
		t.Errorf("Install[0].Name = %s, want bundled/typescript-rules", cfg.Install[0].Name)
	}
	if len(cfg.Install[1].Tools) != 1 {
		t.Errorf("Install[1].Tools count = %d, want 1", len(cfg.Install[1].Tools))
	}

	if cfg.Scope != "local" {
		t.Errorf("Scope = %s, want local", cfg.Scope)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     AiConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: AiConfig{
				Tools: []string{"windsurf", "cursor"},
				Install: []InstallConfig{
					{Name: "bundled/test"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing tools",
			cfg: AiConfig{
				Install: []InstallConfig{
					{Name: "bundled/test"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid tool",
			cfg: AiConfig{
				Tools: []string{"invalid-tool"},
				Install: []InstallConfig{
					{Name: "bundled/test"},
				},
			},
			wantErr: true,
		},
		{
			name: "resource tool not in global tools",
			cfg: AiConfig{
				Tools: []string{"windsurf"},
				Install: []InstallConfig{
					{Name: "bundled/test", Tools: []string{"cursor"}},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid scope",
			cfg: AiConfig{
				Tools: []string{"windsurf"},
				Install: []InstallConfig{
					{Name: "bundled/test"},
				},
				Scope: "invalid",
			},
			wantErr: true,
		},
		{
			name: "missing install name",
			cfg: AiConfig{
				Tools: []string{"windsurf"},
				Install: []InstallConfig{
					{Name: ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetScope(t *testing.T) {
	tests := []struct {
		name  string
		scope string
		want  string
	}{
		{"empty defaults to local", "", "local"},
		{"explicit local", "local", "local"},
		{"explicit global", "global", "global"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AiConfig{Scope: tt.scope}
			if got := cfg.GetScope(); got != tt.want {
				t.Errorf("GetScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetToolsForResource(t *testing.T) {
	cfg := &AiConfig{
		Tools: []string{"windsurf", "cursor"},
	}

	t.Run("uses global tools when resource has none", func(t *testing.T) {
		inst := InstallConfig{Name: "bundled/test"}
		tools := cfg.GetToolsForResource(inst)
		if len(tools) != 2 {
			t.Errorf("Expected 2 tools, got %d", len(tools))
		}
	})

	t.Run("uses resource tools when specified", func(t *testing.T) {
		inst := InstallConfig{Name: "bundled/test", Tools: []string{"windsurf"}}
		tools := cfg.GetToolsForResource(inst)
		if len(tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(tools))
		}
		if tools[0] != "windsurf" {
			t.Errorf("Expected windsurf, got %s", tools[0])
		}
	})
}

func TestFindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// No config file
	_, err := FindConfigFile()
	if err == nil {
		t.Error("FindConfigFile() should error when no config exists")
	}

	// Create .aiconfig
	_ = os.WriteFile(filepath.Join(tmpDir, ".aiconfig"), []byte("tools: [windsurf]"), 0644)

	path, err := FindConfigFile()
	if err != nil {
		t.Fatalf("FindConfigFile() error = %v", err)
	}
	if filepath.Base(path) != ".aiconfig" {
		t.Errorf("FindConfigFile() = %s, want .aiconfig", filepath.Base(path))
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".aiconfig")

	cfg := &AiConfig{
		Tools: []string{"windsurf", "cursor"},
		Install: []InstallConfig{
			{Name: "bundled/test"},
		},
		Scope: "local",
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify by loading
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Tools) != 2 {
		t.Errorf("Loaded tools count = %d, want 2", len(loaded.Tools))
	}
	if loaded.Scope != "local" {
		t.Errorf("Loaded scope = %s, want local", loaded.Scope)
	}
}

func TestNewFromState(t *testing.T) {
	tools := []string{"windsurf", "cursor"}
	install := []InstallConfig{
		{Name: "bundled/test", Tools: []string{"windsurf"}},
	}

	cfg := NewFromState(tools, install, "global")

	if len(cfg.Tools) != 2 {
		t.Errorf("Tools count = %d, want 2", len(cfg.Tools))
	}
	if len(cfg.Install) != 1 {
		t.Errorf("Install count = %d, want 1", len(cfg.Install))
	}
	if cfg.Scope != "global" {
		t.Errorf("Scope = %s, want global", cfg.Scope)
	}
}
