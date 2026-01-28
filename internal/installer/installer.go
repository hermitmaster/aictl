package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/resource"
)

// Installer defines the interface for tool-specific installers
type Installer interface {
	// Install installs a resource to the tool's config directory
	Install(res *resource.Resource, global bool, force bool) ([]string, error)

	// Uninstall removes a resource from the tool's config directory
	Uninstall(res *resource.Resource, files []string, global bool) error

	// SupportsType returns true if the tool supports the given resource type
	SupportsType(t resource.Type) bool

	// GetTool returns the tool this installer is for
	GetTool() config.Tool
}

// BaseInstaller provides common functionality for all installers
type BaseInstaller struct {
	Config config.ToolConfig
}

// GetTool returns the tool this installer is for
func (b *BaseInstaller) GetTool() config.Tool {
	return b.Config.Name
}

// GetDestDir returns the destination directory for a resource type
func (b *BaseInstaller) GetDestDir(t resource.Type, global bool) (string, error) {
	var baseDir string
	if global {
		baseDir = b.Config.GlobalDir
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(cwd, b.Config.LocalDir)
	}

	var subDir string
	switch t {
	case resource.TypeRules:
		subDir = b.Config.RulesDir
	case resource.TypeWorkflow:
		subDir = b.Config.WorkflowsDir
	case resource.TypeSkill:
		subDir = b.Config.SkillsDir
	case resource.TypeBin:
		subDir = b.Config.BinDir
	default:
		return "", fmt.Errorf("unsupported resource type: %s", t)
	}

	if subDir == "" {
		return "", fmt.Errorf("tool %s does not support resource type %s", b.Config.Name, t)
	}

	return filepath.Join(baseDir, subDir), nil
}

// CopyFile copies a file from src to dest, creating directories as needed
func CopyFile(src, dest string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}

// WriteContent writes content to a file, creating directories as needed
func WriteContent(content, dest string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	if err := os.WriteFile(dest, []byte(content), mode); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// RemoveFile removes a file and cleans up empty parent directories
func RemoveFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing file: %w", err)
	}

	// Try to remove empty parent directories
	dir := filepath.Dir(path)
	for {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		if err := os.Remove(dir); err != nil {
			break
		}
		dir = filepath.Dir(dir)
	}

	return nil
}

// GetInstaller returns the appropriate installer for a tool
func GetInstaller(tool config.Tool) (Installer, error) {
	switch tool {
	case config.ToolWindsurf:
		return NewWindsurfInstaller(), nil
	case config.ToolCursor:
		return NewCursorInstaller(), nil
	case config.ToolAider:
		return NewAiderInstaller(), nil
	case config.ToolContinue:
		return NewContinueInstaller(), nil
	default:
		return nil, fmt.Errorf("unsupported tool: %s", tool)
	}
}
