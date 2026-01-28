package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/resource"
)

// CursorInstaller handles installation for Cursor
type CursorInstaller struct {
	BaseInstaller
}

// NewCursorInstaller creates a new Cursor installer
func NewCursorInstaller() *CursorInstaller {
	return &CursorInstaller{
		BaseInstaller: BaseInstaller{
			Config: config.GetToolConfig(config.ToolCursor),
		},
	}
}

// SupportsType returns true if Cursor supports the given resource type
func (c *CursorInstaller) SupportsType(t resource.Type) bool {
	switch t {
	case resource.TypeRules:
		return c.Config.SupportsRules
	default:
		return false
	}
}

// Install installs a resource to Cursor's config directory
func (c *CursorInstaller) Install(res *resource.Resource, global bool, force bool) ([]string, error) {
	if !c.SupportsType(res.Type) {
		return nil, fmt.Errorf("cursor does not support resource type: %s", res.Type)
	}

	destDir, err := c.GetDestDir(res.Type, global)
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

// Uninstall removes a resource from Cursor's config directory
func (c *CursorInstaller) Uninstall(res *resource.Resource, files []string, global bool) error {
	for _, file := range files {
		if err := RemoveFile(file); err != nil {
			return err
		}
	}
	return nil
}

// InstallFromReader installs a resource from an io.Reader (for embedded resources)
func (c *CursorInstaller) InstallFromReader(res *resource.Resource, content io.Reader, global bool, force bool) (string, error) {
	if !c.SupportsType(res.Type) {
		return "", fmt.Errorf("cursor does not support resource type: %s", res.Type)
	}

	destDir, err := c.GetDestDir(res.Type, global)
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
