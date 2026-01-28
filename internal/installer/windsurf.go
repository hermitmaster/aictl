package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/resource"
)

// WindsurfInstaller handles installation for Windsurf
type WindsurfInstaller struct {
	BaseInstaller
}

// NewWindsurfInstaller creates a new Windsurf installer
func NewWindsurfInstaller() *WindsurfInstaller {
	return &WindsurfInstaller{
		BaseInstaller: BaseInstaller{
			Config: config.GetToolConfig(config.ToolWindsurf),
		},
	}
}

// SupportsType returns true if Windsurf supports the given resource type
func (w *WindsurfInstaller) SupportsType(t resource.Type) bool {
	switch t {
	case resource.TypeRules:
		return w.Config.SupportsRules
	case resource.TypeWorkflow:
		return w.Config.SupportsWorkflows
	case resource.TypeSkill:
		return w.Config.SupportsSkills
	case resource.TypeBin:
		return w.Config.SupportsBin
	default:
		return false
	}
}

// Install installs a resource to Windsurf's config directory
func (w *WindsurfInstaller) Install(res *resource.Resource, global bool, force bool) ([]string, error) {
	if !w.SupportsType(res.Type) {
		return nil, fmt.Errorf("windsurf does not support resource type: %s", res.Type)
	}

	var installedFiles []string

	if res.IsMultiFile() {
		files, err := w.installMultiFile(res, global, force)
		if err != nil {
			return nil, err
		}
		installedFiles = files
	} else {
		file, err := w.installSingleFile(res, global, force)
		if err != nil {
			return nil, err
		}
		installedFiles = []string{file}
	}

	return installedFiles, nil
}

func (w *WindsurfInstaller) installSingleFile(res *resource.Resource, global bool, force bool) (string, error) {
	destDir, err := w.GetDestDir(res.Type, global)
	if err != nil {
		return "", err
	}

	// Skills require a subdirectory: skills/<name>/SKILL.md
	var destPath string
	if res.Type == resource.TypeSkill {
		destPath = filepath.Join(destDir, res.Name, "SKILL.md")
	} else {
		destPath = filepath.Join(destDir, res.Name+".md")
	}

	// Check if file exists
	if _, err := os.Stat(destPath); err == nil && !force {
		return "", fmt.Errorf("file already exists: %s (use --force to overwrite)", destPath)
	}

	// Read the full source file content (including frontmatter)
	var content []byte
	if res.SourcePath != "" {
		content, err = os.ReadFile(res.SourcePath)
		if err != nil {
			return "", fmt.Errorf("error reading source file: %w", err)
		}
	} else {
		return "", fmt.Errorf("no source path for resource")
	}

	if err := WriteContent(string(content), destPath, 0644); err != nil {
		return "", err
	}

	return destPath, nil
}

func (w *WindsurfInstaller) installMultiFile(res *resource.Resource, global bool, force bool) ([]string, error) {
	var baseDir string
	if global {
		baseDir = w.Config.GlobalDir
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		baseDir = filepath.Join(cwd, w.Config.LocalDir)
	}

	var installedFiles []string
	sourceDir := res.SourcePath // SourcePath is the directory for multi-file resources

	for _, fm := range res.Files {
		srcPath := filepath.Join(sourceDir, fm.Src)
		destPath := filepath.Join(baseDir, fm.Dest)

		// Check if file exists
		if _, err := os.Stat(destPath); err == nil && !force {
			return installedFiles, fmt.Errorf("file already exists: %s (use --force to overwrite)", destPath)
		}

		mode := os.FileMode(0644)
		if fm.Mode != 0 {
			mode = os.FileMode(fm.Mode)
		}

		if err := CopyFile(srcPath, destPath, mode); err != nil {
			return installedFiles, err
		}

		installedFiles = append(installedFiles, destPath)
	}

	return installedFiles, nil
}

// Uninstall removes a resource from Windsurf's config directory
func (w *WindsurfInstaller) Uninstall(res *resource.Resource, files []string, global bool) error {
	for _, file := range files {
		if err := RemoveFile(file); err != nil {
			return err
		}
	}
	return nil
}

// InstallFromReader installs a resource from an io.Reader (for embedded resources)
func (w *WindsurfInstaller) InstallFromReader(res *resource.Resource, content io.Reader, global bool, force bool) (string, error) {
	if !w.SupportsType(res.Type) {
		return "", fmt.Errorf("windsurf does not support resource type: %s", res.Type)
	}

	destDir, err := w.GetDestDir(res.Type, global)
	if err != nil {
		return "", err
	}

	// Skills require a subdirectory: skills/<name>/SKILL.md
	var destPath string
	if res.Type == resource.TypeSkill {
		destPath = filepath.Join(destDir, res.Name, "SKILL.md")
	} else {
		destPath = filepath.Join(destDir, res.Name+".md")
	}

	// Check if file exists
	if _, err := os.Stat(destPath); err == nil && !force {
		return "", fmt.Errorf("file already exists: %s (use --force to overwrite)", destPath)
	}

	data, err := io.ReadAll(content)
	if err != nil {
		return "", fmt.Errorf("error reading content: %w", err)
	}

	if err := WriteContent(string(data), destPath, 0644); err != nil {
		return "", err
	}

	return destPath, nil
}
