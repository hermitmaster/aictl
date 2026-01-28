package aiconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/hermitmaster/aictl/internal/config"
)

// AiConfig represents the .aiconfig file structure
type AiConfig struct {
	// Tools installed on this system (scopes 'all' operations)
	Tools []string `yaml:"tools"`

	// Custom registries
	Taps []TapConfig `yaml:"taps,omitempty"`

	// Resources to install
	Install []InstallConfig `yaml:"install"`

	// Default installation scope (global or local)
	Scope string `yaml:"scope,omitempty"`
}

// TapConfig represents a custom registry configuration
type TapConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// InstallConfig represents a resource to install
type InstallConfig struct {
	Name  string   `yaml:"name"`
	Tools []string `yaml:"tools,omitempty"`
}

// DefaultConfigFiles returns the list of config file names to search for
func DefaultConfigFiles() []string {
	return []string{".aiconfig", ".aiconfig.yaml", ".aiconfig.yml"}
}

// FindConfigFile searches for a config file in the current directory
func FindConfigFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	for _, name := range DefaultConfigFiles() {
		path := filepath.Join(cwd, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no .aiconfig file found in current directory")
}

// Load loads an AiConfig from a file path
func Load(path string) (*AiConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg AiConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &cfg, nil
}

// LoadOrFind loads a config from the specified path, or finds one in the current directory
func LoadOrFind(path string) (*AiConfig, string, error) {
	if path != "" {
		cfg, err := Load(path)
		return cfg, path, err
	}

	foundPath, err := FindConfigFile()
	if err != nil {
		return nil, "", err
	}

	cfg, err := Load(foundPath)
	return cfg, foundPath, err
}

// Validate checks if the config is valid
func (c *AiConfig) Validate() error {
	if len(c.Tools) == 0 {
		return fmt.Errorf("tools array is required in .aiconfig")
	}

	// Validate tool names
	validTools := make(map[string]bool)
	for _, t := range config.AllTools() {
		validTools[string(t)] = true
	}

	for _, t := range c.Tools {
		if !validTools[t] {
			return fmt.Errorf("invalid tool: %s (valid tools: windsurf, cursor, aider, continue)", t)
		}
	}

	// Validate install entries
	for i, inst := range c.Install {
		if inst.Name == "" {
			return fmt.Errorf("install[%d]: name is required", i)
		}

		// Validate per-resource tools are subset of global tools
		for _, t := range inst.Tools {
			found := false
			for _, globalTool := range c.Tools {
				if t == globalTool {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("install[%d]: tool '%s' is not in the global tools list", i, t)
			}
		}
	}

	// Validate scope
	if c.Scope != "" && c.Scope != "global" && c.Scope != "local" {
		return fmt.Errorf("invalid scope: %s (must be 'global' or 'local')", c.Scope)
	}

	return nil
}

// GetScope returns the scope, defaulting to "local" if not specified
func (c *AiConfig) GetScope() string {
	if c.Scope == "" {
		return "local"
	}
	return c.Scope
}

// GetToolsForResource returns the tools to install a resource to
// If the resource specifies tools, use those; otherwise use the global tools list
func (c *AiConfig) GetToolsForResource(inst InstallConfig) []string {
	if len(inst.Tools) > 0 {
		return inst.Tools
	}
	return c.Tools
}

// Save writes the config to a file
func (c *AiConfig) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// NewFromState creates an AiConfig from the current installed state
func NewFromState(tools []string, installed []InstallConfig, scope string) *AiConfig {
	return &AiConfig{
		Tools:   tools,
		Install: installed,
		Scope:   scope,
	}
}
