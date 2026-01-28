package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hermitmaster/aictl/internal/config"
)

func TestStateAddInstalled(t *testing.T) {
	st := &State{Version: 1}

	res := InstalledResource{
		Name:        "test-resource",
		Source:      "bundled",
		Type:        "workflow",
		Version:     "1.0.0",
		InstalledAt: time.Now(),
		Tools: map[string]ToolInstallInfo{
			"windsurf": {Files: []string{"/path/to/file.md"}},
		},
	}

	st.AddInstalled(res)

	if len(st.Installed) != 1 {
		t.Fatalf("Expected 1 installed resource, got %d", len(st.Installed))
	}
	if st.Installed[0].Name != "test-resource" {
		t.Errorf("Expected name 'test-resource', got '%s'", st.Installed[0].Name)
	}
}

func TestStateAddInstalledUpdate(t *testing.T) {
	st := &State{Version: 1}

	res1 := InstalledResource{
		Name:    "test-resource",
		Version: "1.0.0",
	}
	st.AddInstalled(res1)

	res2 := InstalledResource{
		Name:    "test-resource",
		Version: "2.0.0",
	}
	st.AddInstalled(res2)

	if len(st.Installed) != 1 {
		t.Fatalf("Expected 1 installed resource after update, got %d", len(st.Installed))
	}
	if st.Installed[0].Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", st.Installed[0].Version)
	}
}

func TestStateRemoveInstalled(t *testing.T) {
	st := &State{
		Version: 1,
		Installed: []InstalledResource{
			{Name: "resource-1"},
			{Name: "resource-2"},
			{Name: "resource-3"},
		},
	}

	removed := st.RemoveInstalled("resource-2")
	if !removed {
		t.Error("RemoveInstalled should return true for existing resource")
	}
	if len(st.Installed) != 2 {
		t.Fatalf("Expected 2 installed resources after removal, got %d", len(st.Installed))
	}

	// Verify resource-2 is gone
	for _, res := range st.Installed {
		if res.Name == "resource-2" {
			t.Error("resource-2 should have been removed")
		}
	}
}

func TestStateRemoveInstalledNotFound(t *testing.T) {
	st := &State{
		Version: 1,
		Installed: []InstalledResource{
			{Name: "resource-1"},
		},
	}

	removed := st.RemoveInstalled("nonexistent")
	if removed {
		t.Error("RemoveInstalled should return false for nonexistent resource")
	}
	if len(st.Installed) != 1 {
		t.Error("Installed list should be unchanged")
	}
}

func TestStateGetInstalled(t *testing.T) {
	st := &State{
		Version: 1,
		Installed: []InstalledResource{
			{Name: "resource-1", Version: "1.0.0"},
			{Name: "resource-2", Version: "2.0.0"},
		},
	}

	res := st.GetInstalled("resource-2")
	if res == nil {
		t.Fatal("GetInstalled should return the resource")
	}
	if res.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", res.Version)
	}

	notFound := st.GetInstalled("nonexistent")
	if notFound != nil {
		t.Error("GetInstalled should return nil for nonexistent resource")
	}
}

func TestStateIsInstalled(t *testing.T) {
	st := &State{
		Version: 1,
		Installed: []InstalledResource{
			{Name: "resource-1"},
		},
	}

	if !st.IsInstalled("resource-1") {
		t.Error("IsInstalled should return true for existing resource")
	}
	if st.IsInstalled("nonexistent") {
		t.Error("IsInstalled should return false for nonexistent resource")
	}
}

func TestStateManagerLoadSave(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		StateDir: tmpDir,
	}

	mgr, err := NewStateManager(cfg)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Test saving and loading global state
	st := &State{
		Version: 1,
		Taps: []Tap{
			{Name: "mycompany", URL: "https://github.com/mycompany/resources"},
		},
		Installed: []InstalledResource{
			{
				Name:        "test-resource",
				Source:      "bundled",
				Type:        "workflow",
				Version:     "1.0.0",
				InstalledAt: time.Now().Truncate(time.Second),
				Tools: map[string]ToolInstallInfo{
					"windsurf": {Files: []string{"/path/to/file.md"}},
				},
			},
		},
	}

	if err := mgr.SaveGlobal(st); err != nil {
		t.Fatalf("SaveGlobal() error = %v", err)
	}

	loaded, err := mgr.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}

	if loaded.Version != st.Version {
		t.Errorf("Version = %d, want %d", loaded.Version, st.Version)
	}
	if len(loaded.Taps) != 1 {
		t.Errorf("Taps count = %d, want 1", len(loaded.Taps))
	}
	if len(loaded.Installed) != 1 {
		t.Errorf("Installed count = %d, want 1", len(loaded.Installed))
	}
	if loaded.Installed[0].Name != "test-resource" {
		t.Errorf("Installed name = %s, want test-resource", loaded.Installed[0].Name)
	}
}

func TestStateManagerLoadNonexistent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		StateDir: tmpDir,
	}

	mgr, err := NewStateManager(cfg)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Loading nonexistent file should return empty state
	st, err := mgr.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() should not error for nonexistent file, got %v", err)
	}
	if st.Version != 1 {
		t.Errorf("Version should be 1 for new state, got %d", st.Version)
	}
	if len(st.Installed) != 0 {
		t.Error("Installed should be empty for new state")
	}
}

func TestStateManagerLocalState(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory for local state test
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	cfg := &config.Config{
		StateDir: filepath.Join(tmpDir, "global"),
	}

	mgr, err := NewStateManager(cfg)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	st := &State{
		Version: 1,
		Installed: []InstalledResource{
			{Name: "local-resource"},
		},
	}

	if err := mgr.SaveLocal(st); err != nil {
		t.Fatalf("SaveLocal() error = %v", err)
	}

	loaded, err := mgr.LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() error = %v", err)
	}

	if len(loaded.Installed) != 1 {
		t.Errorf("Installed count = %d, want 1", len(loaded.Installed))
	}
}
