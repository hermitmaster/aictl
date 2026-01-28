package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/resource"
)

// AiderInstaller handles installation for Aider
type AiderInstaller struct {
	BaseInstaller
}

// NewAiderInstaller creates a new Aider installer
func NewAiderInstaller() *AiderInstaller {
	return &AiderInstaller{
		BaseInstaller: BaseInstaller{
			Config: config.GetToolConfig(config.ToolAider),
		},
	}
}

// SupportsType returns true if Aider supports the given resource type
func (a *AiderInstaller) SupportsType(t resource.Type) bool {
	switch t {
	case resource.TypeRules:
		return a.Config.SupportsRules
	default:
		return false
	}
}

// Install installs a resource to Aider's config directory
func (a *AiderInstaller) Install(res *resource.Resource, global bool, force bool) ([]string, error) {
	if !a.SupportsType(res.Type) {
		return nil, fmt.Errorf("aider does not support resource type: %s", res.Type)
	}

	destDir, err := a.GetDestDir(res.Type, global)
	if err != nil {
		return nil, err
	}

	destPath := filepath.Join(destDir, res.Name+".md")

	if _, err := os.Stat(destPath); err == nil && !force {
		return nil, fmt.Errorf("file already exists: %s (use --force to overwrite)", destPath)
	}

	var content []byte
	if res.SourcePath != "" {
		content, err = os.ReadFile(res.SourcePath)
		if err != nil {
			return nil, fmt.Errorf("error reading source file: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no source path for resource")
	}

	if err := WriteContent(string(content), destPath, 0o644); err != nil {
		return nil, err
	}

	return []string{destPath}, nil
}

// Uninstall removes a resource from Aider's config directory
func (a *AiderInstaller) Uninstall(res *resource.Resource, files []string, global bool) error {
	for _, file := range files {
		if err := RemoveFile(file); err != nil {
			return err
		}
	}
	return nil
}

// InstallFromReader installs a resource from an io.Reader (for embedded resources)
func (a *AiderInstaller) InstallFromReader(res *resource.Resource, content io.Reader, global bool, force bool) (string, error) {
	if !a.SupportsType(res.Type) {
		return "", fmt.Errorf("aider does not support resource type: %s", res.Type)
	}

	destDir, err := a.GetDestDir(res.Type, global)
	if err != nil {
		return "", err
	}

	destPath := filepath.Join(destDir, res.Name+".md")

	if _, err := os.Stat(destPath); err == nil && !force {
		return "", fmt.Errorf("file already exists: %s (use --force to overwrite)", destPath)
	}

	data, err := io.ReadAll(content)
	if err != nil {
		return "", fmt.Errorf("error reading content: %w", err)
	}

	if err := WriteContent(string(data), destPath, 0o644); err != nil {
		return "", err
	}

	return destPath, nil
}
