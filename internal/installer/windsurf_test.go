package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/resource"
)

func TestWindsurfInstallerSupportsType(t *testing.T) {
	inst := NewWindsurfInstaller()

	tests := []struct {
		resType resource.Type
		want    bool
	}{
		{resource.TypeRules, true},
		{resource.TypeWorkflow, true},
		{resource.TypeSkill, true},
		{resource.TypeBin, true},
		{resource.Type("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.resType), func(t *testing.T) {
			if got := inst.SupportsType(tt.resType); got != tt.want {
				t.Errorf("SupportsType(%v) = %v, want %v", tt.resType, got, tt.want)
			}
		})
	}
}

func TestWindsurfInstallerGetTool(t *testing.T) {
	inst := NewWindsurfInstaller()
	if inst.GetTool() != config.ToolWindsurf {
		t.Errorf("GetTool() = %v, want %v", inst.GetTool(), config.ToolWindsurf)
	}
}

func TestWindsurfInstallerInstallFromReader(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a custom installer with temp directory
	inst := &WindsurfInstaller{
		BaseInstaller: BaseInstaller{
			Config: config.ToolConfig{
				Name:         config.ToolWindsurf,
				GlobalDir:    tmpDir,
				LocalDir:     ".windsurf",
				RulesDir:     "memories",
				WorkflowsDir: "global_workflows",
				SkillsDir:    "skills",
				BinDir:       "bin",
				SupportsRules:    true,
				SupportsWorkflows: true,
				SupportsSkills:   true,
				SupportsBin:      true,
			},
		},
	}

	res := &resource.Resource{
		Name:        "test-workflow",
		Version:     "1.0.0",
		Type:        resource.TypeWorkflow,
		Description: "Test workflow",
	}

	content := `---
name: test-workflow
version: 1.0.0
type: workflow
description: Test workflow
---

# Test Workflow

This is test content.`

	destPath, err := inst.InstallFromReader(res, strings.NewReader(content), true, false)
	if err != nil {
		t.Fatalf("InstallFromReader() error = %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "global_workflows", "test-workflow.md")
	if destPath != expectedPath {
		t.Errorf("destPath = %v, want %v", destPath, expectedPath)
	}

	// Verify file was created
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read installed file: %v", err)
	}
	if string(data) != content {
		t.Errorf("Content mismatch")
	}
}

func TestWindsurfInstallerInstallFromReaderNoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	inst := &WindsurfInstaller{
		BaseInstaller: BaseInstaller{
			Config: config.ToolConfig{
				Name:         config.ToolWindsurf,
				GlobalDir:    tmpDir,
				WorkflowsDir: "global_workflows",
				SupportsWorkflows: true,
			},
		},
	}

	res := &resource.Resource{
		Name: "existing",
		Type: resource.TypeWorkflow,
	}

	// Create existing file
	destDir := filepath.Join(tmpDir, "global_workflows")
	_ = os.MkdirAll(destDir, 0755)
	existingPath := filepath.Join(destDir, "existing.md")
	_ = os.WriteFile(existingPath, []byte("original"), 0644)

	// Try to install without force
	_, err := inst.InstallFromReader(res, strings.NewReader("new content"), true, false)
	if err == nil {
		t.Error("InstallFromReader() should error when file exists and force=false")
	}

	// Verify original content unchanged
	data, _ := os.ReadFile(existingPath)
	if string(data) != "original" {
		t.Error("Original file should not be modified")
	}
}

func TestWindsurfInstallerInstallFromReaderWithForce(t *testing.T) {
	tmpDir := t.TempDir()

	inst := &WindsurfInstaller{
		BaseInstaller: BaseInstaller{
			Config: config.ToolConfig{
				Name:         config.ToolWindsurf,
				GlobalDir:    tmpDir,
				WorkflowsDir: "global_workflows",
				SupportsWorkflows: true,
			},
		},
	}

	res := &resource.Resource{
		Name: "existing",
		Type: resource.TypeWorkflow,
	}

	// Create existing file
	destDir := filepath.Join(tmpDir, "global_workflows")
	_ = os.MkdirAll(destDir, 0755)
	existingPath := filepath.Join(destDir, "existing.md")
	_ = os.WriteFile(existingPath, []byte("original"), 0644)

	// Install with force
	_, err := inst.InstallFromReader(res, strings.NewReader("new content"), true, true)
	if err != nil {
		t.Fatalf("InstallFromReader() with force should succeed, got error: %v", err)
	}

	// Verify content was overwritten
	data, _ := os.ReadFile(existingPath)
	if string(data) != "new content" {
		t.Errorf("File should be overwritten, got: %s", string(data))
	}
}

func TestWindsurfInstallerInstallUnsupportedType(t *testing.T) {
	inst := &WindsurfInstaller{
		BaseInstaller: BaseInstaller{
			Config: config.ToolConfig{
				Name:              config.ToolWindsurf,
				SupportsWorkflows: false, // Disable workflows for test
			},
		},
	}

	res := &resource.Resource{
		Name: "test",
		Type: resource.TypeWorkflow,
	}

	_, err := inst.InstallFromReader(res, strings.NewReader("content"), true, false)
	if err == nil {
		t.Error("InstallFromReader() should error for unsupported type")
	}
}

func TestWindsurfInstallerInstallSkillInSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	inst := &WindsurfInstaller{
		BaseInstaller: BaseInstaller{
			Config: config.ToolConfig{
				Name:           config.ToolWindsurf,
				GlobalDir:      tmpDir,
				SkillsDir:      "skills",
				SupportsSkills: true,
			},
		},
	}

	res := &resource.Resource{
		Name:        "test-skill",
		Version:     "1.0.0",
		Type:        resource.TypeSkill,
		Description: "Test skill",
	}

	content := `---
name: test-skill
version: 1.0.0
type: skill
description: Test skill
---

# Test Skill

This is test content.`

	destPath, err := inst.InstallFromReader(res, strings.NewReader(content), true, false)
	if err != nil {
		t.Fatalf("InstallFromReader() error = %v", err)
	}

	// Skills should be installed in subdirectory: skills/<name>/SKILL.md
	expectedPath := filepath.Join(tmpDir, "skills", "test-skill", "SKILL.md")
	if destPath != expectedPath {
		t.Errorf("destPath = %v, want %v", destPath, expectedPath)
	}

	// Verify file was created
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read installed file: %v", err)
	}
	if string(data) != content {
		t.Errorf("Content mismatch")
	}
}

func TestWindsurfInstallerUninstall(t *testing.T) {
	tmpDir := t.TempDir()

	inst := NewWindsurfInstaller()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.md")
	file2 := filepath.Join(tmpDir, "subdir", "file2.md")

	_ = os.WriteFile(file1, []byte("content1"), 0644)
	_ = os.MkdirAll(filepath.Dir(file2), 0755)
	_ = os.WriteFile(file2, []byte("content2"), 0644)

	files := []string{file1, file2}

	if err := inst.Uninstall(nil, files, true); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	// Verify files are removed
	for _, f := range files {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("File %s should have been removed", f)
		}
	}
}
