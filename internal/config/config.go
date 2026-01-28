package config

import (
	"os"
	"path/filepath"
)

// Tool represents a supported AI coding assistant
type Tool string

const (
	ToolWindsurf Tool = "windsurf"
	ToolCursor   Tool = "cursor"
	ToolAider    Tool = "aider"
	ToolContinue Tool = "continue"

	// DefaultRegistry is the default resource registry URL
	DefaultRegistry = "https://github.com/hermitmaster/aictl-registry"
)

// AllTools returns all supported tools
func AllTools() []Tool {
	return []Tool{ToolWindsurf, ToolCursor, ToolAider, ToolContinue}
}

// ToolConfig holds configuration for a specific tool
type ToolConfig struct {
	Name              Tool
	GlobalDir         string
	LocalDir          string
	RulesDir          string
	WorkflowsDir      string
	SkillsDir         string
	BinDir            string
	ScriptsDir        string
	SupportsRules     bool
	SupportsWorkflows bool
	SupportsSkills    bool
	SupportsBin       bool
}

// GetToolConfig returns the configuration for a specific tool
func GetToolConfig(tool Tool) ToolConfig {
	home, _ := os.UserHomeDir()

	switch tool {
	case ToolWindsurf:
		return ToolConfig{
			Name:              tool,
			GlobalDir:         filepath.Join(home, ".codeium", "windsurf"),
			LocalDir:          ".windsurf",
			RulesDir:          "memories",
			WorkflowsDir:      "global_workflows",
			SkillsDir:         "skills",
			BinDir:            "bin",
			ScriptsDir:        "scripts",
			SupportsRules:     true,
			SupportsWorkflows: true,
			SupportsSkills:    true,
			SupportsBin:       true,
		}
	case ToolCursor:
		return ToolConfig{
			Name:              tool,
			GlobalDir:         filepath.Join(home, ".cursor"),
			LocalDir:          ".cursor",
			RulesDir:          "rules",
			WorkflowsDir:      "",
			SkillsDir:         "",
			BinDir:            "",
			ScriptsDir:        "",
			SupportsRules:     true,
			SupportsWorkflows: false,
			SupportsSkills:    false,
			SupportsBin:       false,
		}
	case ToolAider:
		return ToolConfig{
			Name:              tool,
			GlobalDir:         filepath.Join(home, ".aider"),
			LocalDir:          ".aider",
			RulesDir:          "conventions",
			WorkflowsDir:      "",
			SkillsDir:         "",
			BinDir:            "",
			ScriptsDir:        "",
			SupportsRules:     true,
			SupportsWorkflows: false,
			SupportsSkills:    false,
			SupportsBin:       false,
		}
	case ToolContinue:
		return ToolConfig{
			Name:              tool,
			GlobalDir:         filepath.Join(home, ".continue"),
			LocalDir:          ".continue",
			RulesDir:          "rules",
			WorkflowsDir:      "",
			SkillsDir:         "",
			BinDir:            "",
			ScriptsDir:        "",
			SupportsRules:     true,
			SupportsWorkflows: false,
			SupportsSkills:    false,
			SupportsBin:       false,
		}
	default:
		return ToolConfig{}
	}
}

// Config holds the global aictl configuration
type Config struct {
	Tools    []Tool `yaml:"tools"`
	StateDir string `yaml:"-"`
	CacheDir string `yaml:"-"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Tools:    []Tool{},
		StateDir: filepath.Join(home, ".config", "aictl"),
		CacheDir: filepath.Join(home, ".cache", "aictl"),
	}
}

// GetStateDir returns the state directory, creating it if necessary
func (c *Config) GetStateDir() (string, error) {
	if err := os.MkdirAll(c.StateDir, 0o755); err != nil {
		return "", err
	}
	return c.StateDir, nil
}

// GetCacheDir returns the cache directory, creating it if necessary
func (c *Config) GetCacheDir() (string, error) {
	if err := os.MkdirAll(c.CacheDir, 0o755); err != nil {
		return "", err
	}
	return c.CacheDir, nil
}

// IsToolInstalled checks if a tool appears to be installed
func IsToolInstalled(tool Tool) bool {
	cfg := GetToolConfig(tool)
	_, err := os.Stat(cfg.GlobalDir)
	return err == nil
}

// DetectInstalledTools returns a list of tools that appear to be installed
func DetectInstalledTools() []Tool {
	installed := []Tool{}
	for _, tool := range AllTools() {
		if IsToolInstalled(tool) {
			installed = append(installed, tool)
		}
	}
	return installed
}
