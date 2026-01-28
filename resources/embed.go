package resources

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/hermitmaster/aictl/internal/resource"
)

//go:embed rules workflows skills bin
var bundledFS embed.FS

// LoadBundled loads a bundled resource by name
func LoadBundled(name string) (*resource.Resource, string, error) {
	// Search in each resource type directory
	dirs := []string{"rules", "workflows", "skills", "bin"}

	for _, dir := range dirs {
		// Try as a single file (name.md)
		path := filepath.Join(dir, name+".md")
		content, err := bundledFS.ReadFile(path)
		if err == nil {
			res, _, err := resource.ParseFrontmatter(strings.NewReader(string(content)))
			if err != nil {
				return nil, "", fmt.Errorf("error parsing %s: %w", path, err)
			}
			res.SourcePath = path
			return res, string(content), nil
		}

		// Try as a directory with resource.yaml
		manifestPath := filepath.Join(dir, name, "resource.yaml")
		content, err = bundledFS.ReadFile(manifestPath)
		if err == nil {
			res, err := resource.ParseResourceYAML(strings.NewReader(string(content)))
			if err != nil {
				return nil, "", fmt.Errorf("error parsing %s: %w", manifestPath, err)
			}
			res.SourcePath = filepath.Join(dir, name)
			return res, string(content), nil
		}
	}

	return nil, "", fmt.Errorf("bundled resource not found: %s", name)
}

// ListBundled returns a list of all bundled resources
func ListBundled() ([]*resource.Resource, error) {
	var resources []*resource.Resource

	dirs := []string{"rules", "workflows", "skills", "bin"}

	for _, dir := range dirs {
		entries, err := bundledFS.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()

			// Skip hidden files, gitkeep, and README
			if strings.HasPrefix(name, ".") || strings.ToLower(name) == "readme.md" {
				continue
			}

			if entry.IsDir() {
				// Multi-file resource
				res, _, err := LoadBundled(name)
				if err == nil {
					resources = append(resources, res)
				}
			} else if strings.HasSuffix(name, ".md") {
				// Single-file resource
				resourceName := strings.TrimSuffix(name, ".md")
				res, _, err := LoadBundled(resourceName)
				if err == nil {
					resources = append(resources, res)
				}
			}
		}
	}

	return resources, nil
}

// GetBundledFile returns the content of a file from a bundled resource
func GetBundledFile(resourcePath, filePath string) ([]byte, error) {
	fullPath := filepath.Join(resourcePath, filePath)
	return bundledFS.ReadFile(fullPath)
}

// WalkBundled walks through all bundled resources
func WalkBundled(fn func(path string, d fs.DirEntry) error) error {
	dirs := []string{"rules", "workflows", "skills", "bin"}

	for _, dir := range dirs {
		err := fs.WalkDir(bundledFS, dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return fn(path, d)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
