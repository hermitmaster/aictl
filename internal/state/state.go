package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/hermitmaster/aictl/internal/config"
)

// State represents the installed resources state
type State struct {
	Version   int                `yaml:"version"`
	Taps      []Tap              `yaml:"taps,omitempty"`
	Installed []InstalledResource `yaml:"installed,omitempty"`
}

// Tap represents a custom registry
type Tap struct {
	Name      string    `yaml:"name"`
	URL       string    `yaml:"url"`
	UpdatedAt time.Time `yaml:"updated_at,omitempty"`
}

// InstalledResource represents an installed resource
type InstalledResource struct {
	Name        string                     `yaml:"name"`
	Source      string                     `yaml:"source"`
	Type        string                     `yaml:"type"`
	Version     string                     `yaml:"version"`
	InstalledAt time.Time                  `yaml:"installed_at"`
	Tools       map[string]ToolInstallInfo `yaml:"tools"`
}

// ToolInstallInfo holds installation info for a specific tool
type ToolInstallInfo struct {
	Files []string `yaml:"files"`
}

// StateManager handles reading and writing state
type StateManager struct {
	globalPath string
	localPath  string
}

// NewStateManager creates a new state manager
func NewStateManager(cfg *config.Config) (*StateManager, error) {
	stateDir, err := cfg.GetStateDir()
	if err != nil {
		return nil, err
	}

	return &StateManager{
		globalPath: filepath.Join(stateDir, "state.yaml"),
		localPath:  ".aictl/state.yaml",
	}, nil
}

// LoadGlobal loads the global state
func (m *StateManager) LoadGlobal() (*State, error) {
	return m.load(m.globalPath)
}

// LoadLocal loads the local (project) state
func (m *StateManager) LoadLocal() (*State, error) {
	return m.load(m.localPath)
}

// SaveGlobal saves the global state
func (m *StateManager) SaveGlobal(state *State) error {
	return m.save(m.globalPath, state)
}

// SaveLocal saves the local (project) state
func (m *StateManager) SaveLocal(state *State) error {
	return m.save(m.localPath, state)
}

func (m *StateManager) load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &State{Version: 1}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading state file: %w", err)
	}

	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("error parsing state file: %w", err)
	}

	return &state, nil
}

func (m *StateManager) save(path string, state *State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("error creating state directory: %w", err)
	}

	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("error marshaling state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing state file: %w", err)
	}

	return nil
}

// AddInstalled adds or updates an installed resource in the state
func (s *State) AddInstalled(res InstalledResource) {
	for i, existing := range s.Installed {
		if existing.Name == res.Name {
			s.Installed[i] = res
			return
		}
	}
	s.Installed = append(s.Installed, res)
}

// RemoveInstalled removes an installed resource from the state
func (s *State) RemoveInstalled(name string) bool {
	for i, res := range s.Installed {
		if res.Name == name {
			s.Installed = append(s.Installed[:i], s.Installed[i+1:]...)
			return true
		}
	}
	return false
}

// GetInstalled returns an installed resource by name
func (s *State) GetInstalled(name string) *InstalledResource {
	for i := range s.Installed {
		if s.Installed[i].Name == name {
			return &s.Installed[i]
		}
	}
	return nil
}

// IsInstalled checks if a resource is installed
func (s *State) IsInstalled(name string) bool {
	return s.GetInstalled(name) != nil
}
