package resource

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Type represents the type of resource
type Type string

const (
	TypeRules    Type = "rules"
	TypeWorkflow Type = "workflow"
	TypeSkill    Type = "skill"
	TypeBin      Type = "bin"
)

// Resource represents a single resource with its metadata
type Resource struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Type         Type              `yaml:"type"`
	Description  string            `yaml:"description"`
	Author       string            `yaml:"author,omitempty"`
	Tags         []string          `yaml:"tags,omitempty"`
	Tools        []string          `yaml:"tools,omitempty"`
	Dependencies []string          `yaml:"dependencies,omitempty"`
	EnvVars      *EnvVars          `yaml:"env_vars,omitempty"`
	Files        []FileMapping     `yaml:"files,omitempty"`
	Content      string            `yaml:"-"`
	SourcePath   string            `yaml:"-"`
}

// EnvVars holds required and optional environment variables
type EnvVars struct {
	Required []string `yaml:"required,omitempty"`
	Optional []string `yaml:"optional,omitempty"`
}

// FileMapping maps source files to destination paths (for multi-file resources)
type FileMapping struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
	Mode uint32 `yaml:"mode,omitempty"`
}

// ParseFrontmatter extracts YAML frontmatter from a markdown file
func ParseFrontmatter(r io.Reader) (*Resource, string, error) {
	scanner := bufio.NewScanner(r)
	var frontmatter bytes.Buffer
	var content bytes.Buffer
	inFrontmatter := false
	frontmatterDone := false

	for scanner.Scan() {
		line := scanner.Text()

		if !inFrontmatter && !frontmatterDone && line == "---" {
			inFrontmatter = true
			continue
		}

		if inFrontmatter && line == "---" {
			inFrontmatter = false
			frontmatterDone = true
			continue
		}

		if inFrontmatter {
			frontmatter.WriteString(line)
			frontmatter.WriteString("\n")
		} else if frontmatterDone {
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("error reading file: %w", err)
	}

	var res Resource
	if frontmatter.Len() > 0 {
		if err := yaml.Unmarshal(frontmatter.Bytes(), &res); err != nil {
			return nil, "", fmt.Errorf("error parsing frontmatter: %w", err)
		}
	}

	res.Content = strings.TrimSpace(content.String())
	return &res, res.Content, nil
}

// ParseResourceYAML parses a resource.yaml manifest file
func ParseResourceYAML(r io.Reader) (*Resource, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var res Resource
	if err := yaml.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	return &res, nil
}

// LoadFromFile loads a resource from a file path
func LoadFromFile(path string) (*Resource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer func() { _ = f.Close() }()

	ext := strings.ToLower(filepath.Ext(path))
	var res *Resource

	switch ext {
	case ".md":
		res, _, err = ParseFrontmatter(f)
	case ".yaml", ".yml":
		res, err = ParseResourceYAML(f)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	res.SourcePath = path
	return res, nil
}

// LoadFromDir loads a multi-file resource from a directory containing resource.yaml
func LoadFromDir(dir string) (*Resource, error) {
	manifestPath := filepath.Join(dir, "resource.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("resource.yaml not found in %s", dir)
	}

	res, err := LoadFromFile(manifestPath)
	if err != nil {
		return nil, err
	}

	res.SourcePath = dir
	return res, nil
}

// Validate checks if the resource has all required fields
func (r *Resource) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("resource name is required")
	}
	if r.Version == "" {
		return fmt.Errorf("resource version is required")
	}
	if r.Type == "" {
		return fmt.Errorf("resource type is required")
	}
	if r.Description == "" {
		return fmt.Errorf("resource description is required")
	}

	validTypes := map[Type]bool{
		TypeRules:    true,
		TypeWorkflow: true,
		TypeSkill:    true,
		TypeBin:      true,
	}
	if !validTypes[r.Type] {
		return fmt.Errorf("invalid resource type: %s", r.Type)
	}

	return nil
}

// IsMultiFile returns true if this is a multi-file resource (has file mappings)
func (r *Resource) IsMultiFile() bool {
	return len(r.Files) > 0
}

// SupportsTools returns the list of tools this resource supports
// If no tools are specified, it returns nil (meaning all compatible tools)
func (r *Resource) SupportsTools() []string {
	return r.Tools
}
