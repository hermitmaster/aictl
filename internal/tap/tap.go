package tap

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/resource"
)

// Tap represents a custom resource registry (Git repository)
type Tap struct {
	Name      string    `yaml:"name"`
	URL       string    `yaml:"url"`
	Path      string    `yaml:"-"`
	UpdatedAt time.Time `yaml:"updated_at,omitempty"`
}

// Manager handles tap operations
type Manager struct {
	tapsDir string
}

// NewManager creates a new tap manager
func NewManager(cfg *config.Config) (*Manager, error) {
	cacheDir, err := cfg.GetCacheDir()
	if err != nil {
		return nil, err
	}

	tapsDir := filepath.Join(cacheDir, "taps")
	if err := os.MkdirAll(tapsDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating taps directory: %w", err)
	}

	return &Manager{tapsDir: tapsDir}, nil
}

// Add clones a new tap repository
func (m *Manager) Add(name, url string) (*Tap, error) {
	tapPath := filepath.Join(m.tapsDir, name)

	// Check if tap already exists
	if _, err := os.Stat(tapPath); err == nil {
		return nil, fmt.Errorf("tap '%s' already exists", name)
	}

	// Clone the repository
	cmd := exec.Command("git", "clone", "--depth", "1", url, tapPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error cloning repository: %w", err)
	}

	return &Tap{
		Name:      name,
		URL:       url,
		Path:      tapPath,
		UpdatedAt: time.Now(),
	}, nil
}

// Remove deletes a tap
func (m *Manager) Remove(name string) error {
	tapPath := filepath.Join(m.tapsDir, name)

	if _, err := os.Stat(tapPath); os.IsNotExist(err) {
		return fmt.Errorf("tap '%s' not found", name)
	}

	if err := os.RemoveAll(tapPath); err != nil {
		return fmt.Errorf("error removing tap: %w", err)
	}

	return nil
}

// Update pulls the latest changes for a tap
func (m *Manager) Update(name string) error {
	tapPath := filepath.Join(m.tapsDir, name)

	if _, err := os.Stat(tapPath); os.IsNotExist(err) {
		return fmt.Errorf("tap '%s' not found", name)
	}

	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = tapPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error updating tap: %w", err)
	}

	return nil
}

// UpdateAll updates all taps
func (m *Manager) UpdateAll() error {
	taps, err := m.List()
	if err != nil {
		return err
	}

	for _, tap := range taps {
		fmt.Printf("Updating %s...\n", tap.Name)
		if err := m.Update(tap.Name); err != nil {
			fmt.Printf("  Warning: %v\n", err)
		}
	}

	return nil
}

// List returns all installed taps
func (m *Manager) List() ([]*Tap, error) {
	entries, err := os.ReadDir(m.tapsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading taps directory: %w", err)
	}

	var taps []*Tap
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		tap, err := m.Get(entry.Name())
		if err != nil {
			continue
		}
		taps = append(taps, tap)
	}

	return taps, nil
}

// Get returns a tap by name
func (m *Manager) Get(name string) (*Tap, error) {
	tapPath := filepath.Join(m.tapsDir, name)

	if _, err := os.Stat(tapPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tap '%s' not found", name)
	}

	// Get remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = tapPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting tap URL: %w", err)
	}

	url := strings.TrimSpace(string(output))

	// Get last update time from .git directory
	gitDir := filepath.Join(tapPath, ".git")
	info, err := os.Stat(gitDir)
	var updatedAt time.Time
	if err == nil {
		updatedAt = info.ModTime()
	}

	return &Tap{
		Name:      name,
		URL:       url,
		Path:      tapPath,
		UpdatedAt: updatedAt,
	}, nil
}

// LoadResource loads a resource from a tap
func (m *Manager) LoadResource(tapName, resourceName string) (*resource.Resource, string, error) {
	tapPath := filepath.Join(m.tapsDir, tapName)

	if _, err := os.Stat(tapPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("tap '%s' not found", tapName)
	}

	// Search in standard directories
	dirs := []string{"rules", "workflows", "skills", "bin"}

	for _, dir := range dirs {
		// Try as a single file
		path := filepath.Join(tapPath, dir, resourceName+".md")
		if _, err := os.Stat(path); err == nil {
			res, err := resource.LoadFromFile(path)
			if err != nil {
				return nil, "", err
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, "", err
			}
			return res, string(content), nil
		}

		// Try as a directory with resource.yaml
		dirPath := filepath.Join(tapPath, dir, resourceName)
		manifestPath := filepath.Join(dirPath, "resource.yaml")
		if _, err := os.Stat(manifestPath); err == nil {
			res, err := resource.LoadFromDir(dirPath)
			if err != nil {
				return nil, "", err
			}
			content, err := os.ReadFile(manifestPath)
			if err != nil {
				return nil, "", err
			}
			return res, string(content), nil
		}
	}

	return nil, "", fmt.Errorf("resource '%s' not found in tap '%s'", resourceName, tapName)
}

// ListResources lists all resources in a tap
func (m *Manager) ListResources(tapName string) ([]*resource.Resource, error) {
	tapPath := filepath.Join(m.tapsDir, tapName)

	if _, err := os.Stat(tapPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tap '%s' not found", tapName)
	}

	var resources []*resource.Resource
	dirs := []string{"rules", "workflows", "skills", "bin"}

	for _, dir := range dirs {
		dirPath := filepath.Join(tapPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.IsDir() {
				// Check for resource.yaml in directory
				manifestPath := filepath.Join(path, "resource.yaml")
				if _, err := os.Stat(manifestPath); err == nil {
					res, err := resource.LoadFromDir(path)
					if err == nil {
						resources = append(resources, res)
					}
					return filepath.SkipDir
				}
				return nil
			}

			if strings.HasSuffix(d.Name(), ".md") {
				res, err := resource.LoadFromFile(path)
				if err == nil {
					resources = append(resources, res)
				}
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

// Exists checks if a tap exists
func (m *Manager) Exists(name string) bool {
	tapPath := filepath.Join(m.tapsDir, name)
	_, err := os.Stat(tapPath)
	return err == nil
}
