package tap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hermitmaster/aictl/internal/config"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	// Verify taps directory was created
	tapsDir := filepath.Join(tmpDir, "taps")
	if _, err := os.Stat(tapsDir); os.IsNotExist(err) {
		t.Error("NewManager() should create taps directory")
	}
}

func TestManagerExists(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, _ := NewManager(cfg)

	// Non-existent tap
	if mgr.Exists("nonexistent") {
		t.Error("Exists() should return false for non-existent tap")
	}

	// Create a fake tap directory
	tapPath := filepath.Join(tmpDir, "taps", "test-tap")
	_ = os.MkdirAll(tapPath, 0755)

	if !mgr.Exists("test-tap") {
		t.Error("Exists() should return true for existing tap")
	}
}

func TestManagerList(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, _ := NewManager(cfg)

	// Empty list
	taps, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(taps) != 0 {
		t.Errorf("List() should return empty list, got %d", len(taps))
	}
}

func TestManagerRemoveNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, _ := NewManager(cfg)

	err := mgr.Remove("nonexistent")
	if err == nil {
		t.Error("Remove() should error for non-existent tap")
	}
}

func TestManagerGetNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, _ := NewManager(cfg)

	_, err := mgr.Get("nonexistent")
	if err == nil {
		t.Error("Get() should error for non-existent tap")
	}
}

func TestManagerUpdateNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, _ := NewManager(cfg)

	err := mgr.Update("nonexistent")
	if err == nil {
		t.Error("Update() should error for non-existent tap")
	}
}

func TestManagerLoadResourceNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, _ := NewManager(cfg)

	_, _, err := mgr.LoadResource("nonexistent", "resource")
	if err == nil {
		t.Error("LoadResource() should error for non-existent tap")
	}
}

func TestManagerListResourcesNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, _ := NewManager(cfg)

	_, err := mgr.ListResources("nonexistent")
	if err == nil {
		t.Error("ListResources() should error for non-existent tap")
	}
}

func TestManagerLoadResourceFromFakeTap(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, _ := NewManager(cfg)

	// Create a fake tap with a resource
	tapPath := filepath.Join(tmpDir, "taps", "test-tap")
	rulesDir := filepath.Join(tapPath, "rules")
	_ = os.MkdirAll(rulesDir, 0755)

	resourceContent := `---
name: test-rule
version: 1.0.0
type: rules
description: A test rule
---

# Test Rule

This is a test.
`
	resourcePath := filepath.Join(rulesDir, "test-rule.md")
	_ = os.WriteFile(resourcePath, []byte(resourceContent), 0644)

	// Load the resource
	res, content, err := mgr.LoadResource("test-tap", "test-rule")
	if err != nil {
		t.Fatalf("LoadResource() error = %v", err)
	}

	if res.Name != "test-rule" {
		t.Errorf("Name = %s, want test-rule", res.Name)
	}
	if res.Version != "1.0.0" {
		t.Errorf("Version = %s, want 1.0.0", res.Version)
	}
	if content == "" {
		t.Error("Content should not be empty")
	}
}

func TestManagerListResourcesFromFakeTap(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		CacheDir: tmpDir,
	}

	mgr, _ := NewManager(cfg)

	// Create a fake tap with resources
	tapPath := filepath.Join(tmpDir, "taps", "test-tap")
	rulesDir := filepath.Join(tapPath, "rules")
	workflowsDir := filepath.Join(tapPath, "workflows")
	_ = os.MkdirAll(rulesDir, 0755)
	_ = os.MkdirAll(workflowsDir, 0755)

	rule1 := `---
name: rule1
version: 1.0.0
type: rules
description: Rule 1
---
Content
`
	rule2 := `---
name: rule2
version: 1.0.0
type: rules
description: Rule 2
---
Content
`
	workflow1 := `---
name: workflow1
version: 1.0.0
type: workflow
description: Workflow 1
---
Content
`

	_ = os.WriteFile(filepath.Join(rulesDir, "rule1.md"), []byte(rule1), 0644)
	_ = os.WriteFile(filepath.Join(rulesDir, "rule2.md"), []byte(rule2), 0644)
	_ = os.WriteFile(filepath.Join(workflowsDir, "workflow1.md"), []byte(workflow1), 0644)

	resources, err := mgr.ListResources("test-tap")
	if err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}

	if len(resources) != 3 {
		t.Errorf("ListResources() returned %d resources, want 3", len(resources))
	}
}
