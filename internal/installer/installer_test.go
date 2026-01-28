package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/resource"
)

func TestGetInstaller(t *testing.T) {
	tests := []struct {
		name    string
		tool    config.Tool
		wantErr bool
	}{
		{"windsurf", config.ToolWindsurf, false},
		{"cursor", config.ToolCursor, false},
		{"aider", config.ToolAider, false},
		{"continue", config.ToolContinue, false},
		{"unknown", config.Tool("unknown"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst, err := GetInstaller(tt.tool)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInstaller() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && inst == nil {
				t.Error("GetInstaller() returned nil installer")
			}
			if !tt.wantErr && inst.GetTool() != tt.tool {
				t.Errorf("GetTool() = %v, want %v", inst.GetTool(), tt.tool)
			}
		})
	}
}

func TestBaseInstallerGetDestDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	bi := &BaseInstaller{
		Config: config.ToolConfig{
			Name:         config.ToolWindsurf,
			GlobalDir:    filepath.Join(tmpDir, "global"),
			LocalDir:     ".windsurf",
			RulesDir:     "memories",
			WorkflowsDir: "global_workflows",
			SkillsDir:    "skills",
			BinDir:       "bin",
		},
	}

	tests := []struct {
		name       string
		resType    resource.Type
		global     bool
		wantSuffix string
		wantErr    bool
	}{
		{"global rules", resource.TypeRules, true, "memories", false},
		{"local rules", resource.TypeRules, false, "memories", false},
		{"global workflow", resource.TypeWorkflow, true, "global_workflows", false},
		{"global skill", resource.TypeSkill, true, "skills", false},
		{"global bin", resource.TypeBin, true, "bin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := bi.GetDestDir(tt.resType, tt.global)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDestDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !filepath.IsAbs(dir) && tt.global {
				t.Errorf("GetDestDir() should return absolute path for global, got %v", dir)
			}
			if !tt.wantErr && !contains(dir, tt.wantSuffix) {
				t.Errorf("GetDestDir() = %v, should contain %v", dir, tt.wantSuffix)
			}
		})
	}
}

func TestBaseInstallerGetDestDirUnsupported(t *testing.T) {
	bi := &BaseInstaller{
		Config: config.ToolConfig{
			Name:         config.ToolCursor,
			GlobalDir:    "/tmp/cursor",
			LocalDir:     ".cursor",
			RulesDir:     "rules",
			WorkflowsDir: "", // Cursor doesn't support workflows
		},
	}

	_, err := bi.GetDestDir(resource.TypeWorkflow, true)
	if err == nil {
		t.Error("GetDestDir() should error for unsupported resource type")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "source.txt")
	destPath := filepath.Join(tmpDir, "subdir", "dest.txt")

	content := "test content"
	if err := os.WriteFile(srcPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := CopyFile(srcPath, destPath, 0644); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	// Verify content
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read dest file: %v", err)
	}
	if string(data) != content {
		t.Errorf("Content = %v, want %v", string(data), content)
	}
}

func TestCopyFileExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "script.sh")
	destPath := filepath.Join(tmpDir, "bin", "script.sh")

	if err := os.WriteFile(srcPath, []byte("#!/bin/bash"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := CopyFile(srcPath, destPath, 0755); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("Failed to stat dest file: %v", err)
	}

	// Check executable bit
	if info.Mode()&0100 == 0 {
		t.Error("File should be executable")
	}
}

func TestWriteContent(t *testing.T) {
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "nested", "dir", "file.md")

	content := "# Test\n\nContent here."
	if err := WriteContent(content, destPath, 0644); err != nil {
		t.Fatalf("WriteContent() error = %v", err)
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != content {
		t.Errorf("Content = %v, want %v", string(data), content)
	}
}

func TestRemoveFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested file
	filePath := filepath.Join(tmpDir, "a", "b", "c", "file.txt")
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	if err := RemoveFile(filePath); err != nil {
		t.Fatalf("RemoveFile() error = %v", err)
	}

	// File should be gone
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File should have been removed")
	}

	// Empty parent directories should be cleaned up
	if _, err := os.Stat(filepath.Join(tmpDir, "a", "b", "c")); !os.IsNotExist(err) {
		t.Error("Empty directory 'c' should have been removed")
	}
}

func TestRemoveFileNonexistent(t *testing.T) {
	// Should not error for nonexistent file
	if err := RemoveFile("/nonexistent/path/file.txt"); err != nil {
		t.Errorf("RemoveFile() should not error for nonexistent file, got %v", err)
	}
}

func contains(s, substr string) bool {
	return filepath.Base(filepath.Dir(s)) == substr || filepath.Base(s) == substr ||
		len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
