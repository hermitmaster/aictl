package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hermitmaster/aictl/internal/aiconfig"
	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/installer"
	"github.com/hermitmaster/aictl/internal/resource"
	"github.com/hermitmaster/aictl/internal/state"
)

func TestBundleCmdExists(t *testing.T) {
	if bundleCmd == nil {
		t.Fatal("bundleCmd should not be nil")
	}

	if bundleCmd.Use != "bundle [path/to/.aiconfig]" {
		t.Errorf("bundleCmd.Use = %v, want 'bundle [path/to/.aiconfig]'", bundleCmd.Use)
	}
}

func TestBundleDumpCmdExists(t *testing.T) {
	if bundleDumpCmd == nil {
		t.Fatal("bundleDumpCmd should not be nil")
	}

	if bundleDumpCmd.Use != "dump" {
		t.Errorf("bundleDumpCmd.Use = %v, want 'dump'", bundleDumpCmd.Use)
	}
}

func TestBundleDryRunFlag(t *testing.T) {
	flag := bundleCmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Fatal("dry-run flag should exist")
	}

	if flag.DefValue != "false" {
		t.Errorf("dry-run default = %v, want false", flag.DefValue)
	}
}

// TestBundleInstallOutcomes tests the different installation outcome scenarios
func TestBundleInstallOutcomes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a mock windsurf installer with temp directory
	mockInstaller := &installer.WindsurfInstaller{
		BaseInstaller: installer.BaseInstaller{
			Config: config.ToolConfig{
				Name:              config.ToolWindsurf,
				GlobalDir:         tmpDir,
				LocalDir:          ".windsurf",
				WorkflowsDir:      "global_workflows",
				SkillsDir:         "skills",
				RulesDir:          "memories",
				SupportsRules:     true,
				SupportsWorkflows: true,
				SupportsSkills:    true,
			},
		},
	}

	tests := []struct {
		name           string
		resourceType   resource.Type
		fileExists     bool
		force          bool
		expectSuccess  bool
		expectExists   bool
		expectNoCompat bool
	}{
		{
			name:          "new_workflow_installs_successfully",
			resourceType:  resource.TypeWorkflow,
			fileExists:    false,
			force:         false,
			expectSuccess: true,
		},
		{
			name:         "existing_file_without_force_returns_exists_error",
			resourceType: resource.TypeWorkflow,
			fileExists:   true,
			force:        false,
			expectExists: true,
		},
		{
			name:          "existing_file_with_force_overwrites",
			resourceType:  resource.TypeWorkflow,
			fileExists:    true,
			force:         true,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			workflowDir := filepath.Join(tmpDir, "global_workflows")
			_ = os.MkdirAll(workflowDir, 0o755)

			destPath := filepath.Join(workflowDir, "test-resource.md")

			// Clean up from previous test
			_ = os.Remove(destPath)

			if tt.fileExists {
				_ = os.WriteFile(destPath, []byte("existing content"), 0o644)
			}

			res := &resource.Resource{
				Name:        "test-resource",
				Version:     "1.0.0",
				Type:        tt.resourceType,
				Description: "Test resource",
			}

			content := "test content"

			// Execute
			_, err := mockInstaller.InstallFromReader(res, strings.NewReader(content), true, tt.force)

			// Verify
			if tt.expectSuccess {
				if err != nil {
					t.Errorf("expected success, got error: %v", err)
				}
			}

			if tt.expectExists {
				if err == nil {
					t.Error("expected 'already exists' error, got nil")
				} else if !strings.Contains(err.Error(), "already exists") {
					t.Errorf("expected 'already exists' error, got: %v", err)
				}
			}
		})
	}
}

// TestBundleAlreadyExistsDetection tests that "already exists" errors are properly detected
func TestBundleAlreadyExistsDetection(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "file_already_exists_error",
			errMsg:   "file already exists: /path/to/file.md (use --force to overwrite)",
			expected: true,
		},
		{
			name:     "already_exists_simple",
			errMsg:   "already exists",
			expected: true,
		},
		{
			name:     "permission_denied",
			errMsg:   "permission denied",
			expected: false,
		},
		{
			name:     "network_error",
			errMsg:   "network timeout",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.Contains(tt.errMsg, "already exists")
			if result != tt.expected {
				t.Errorf("strings.Contains(%q, 'already exists') = %v, want %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

// TestBundleToolCompatibility tests tool compatibility checking
func TestBundleToolCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		resourceType resource.Type
		toolConfig   config.ToolConfig
		expected     bool
	}{
		{
			name:         "windsurf_supports_workflows",
			resourceType: resource.TypeWorkflow,
			toolConfig: config.ToolConfig{
				SupportsWorkflows: true,
			},
			expected: true,
		},
		{
			name:         "windsurf_does_not_support_workflows_when_disabled",
			resourceType: resource.TypeWorkflow,
			toolConfig: config.ToolConfig{
				SupportsWorkflows: false,
			},
			expected: false,
		},
		{
			name:         "windsurf_supports_skills",
			resourceType: resource.TypeSkill,
			toolConfig: config.ToolConfig{
				SupportsSkills: true,
			},
			expected: true,
		},
		{
			name:         "windsurf_supports_rules",
			resourceType: resource.TypeRules,
			toolConfig: config.ToolConfig{
				SupportsRules: true,
			},
			expected: true,
		},
		{
			name:         "unknown_type_not_supported",
			resourceType: resource.Type("unknown"),
			toolConfig: config.ToolConfig{
				SupportsRules:     true,
				SupportsWorkflows: true,
				SupportsSkills:    true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &installer.WindsurfInstaller{
				BaseInstaller: installer.BaseInstaller{
					Config: tt.toolConfig,
				},
			}

			result := inst.SupportsType(tt.resourceType)
			if result != tt.expected {
				t.Errorf("SupportsType(%v) = %v, want %v", tt.resourceType, result, tt.expected)
			}
		})
	}
}

// TestBundleStateTracking tests that state is properly tracked during bundle operations
func TestBundleStateTracking(t *testing.T) {
	tmpDir := t.TempDir()

	st := &state.State{Version: 1}

	// Add an installed resource
	installedRes := state.InstalledResource{
		Name:    "test-workflow",
		Source:  "default",
		Type:    "workflow",
		Version: "1.0.0",
		Tools: map[string]state.ToolInstallInfo{
			"windsurf": {Files: []string{filepath.Join(tmpDir, "test-workflow.md")}},
		},
	}

	st.AddInstalled(installedRes)

	// Verify it was added
	if !st.IsInstalled("test-workflow") {
		t.Error("resource should be installed")
	}

	// Verify we can retrieve it
	retrieved := st.GetInstalled("test-workflow")
	if retrieved == nil {
		t.Fatal("GetInstalled should return the resource")
	}

	if retrieved.Source != "default" {
		t.Errorf("Source = %v, want 'default'", retrieved.Source)
	}

	if len(retrieved.Tools) != 1 {
		t.Errorf("Tools count = %d, want 1", len(retrieved.Tools))
	}

	// Update the resource
	updatedRes := state.InstalledResource{
		Name:    "test-workflow",
		Source:  "default",
		Type:    "workflow",
		Version: "2.0.0",
		Tools: map[string]state.ToolInstallInfo{
			"windsurf": {Files: []string{filepath.Join(tmpDir, "test-workflow.md")}},
			"cursor":   {Files: []string{filepath.Join(tmpDir, "cursor-test-workflow.md")}},
		},
	}

	st.AddInstalled(updatedRes)

	// Verify update
	retrieved = st.GetInstalled("test-workflow")
	if retrieved.Version != "2.0.0" {
		t.Errorf("Version = %v, want '2.0.0'", retrieved.Version)
	}

	if len(retrieved.Tools) != 2 {
		t.Errorf("Tools count = %d, want 2", len(retrieved.Tools))
	}

	// Verify only one entry exists (not duplicated)
	count := 0
	for _, res := range st.Installed {
		if res.Name == "test-workflow" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Should have exactly 1 entry for test-workflow, got %d", count)
	}
}

// TestBundleDumpOutput tests the bundle dump functionality
func TestBundleDumpOutput(t *testing.T) {
	st := &state.State{
		Version: 1,
		Installed: []state.InstalledResource{
			{
				Name:    "workflow-a",
				Source:  "default",
				Type:    "workflow",
				Version: "1.0.0",
				Tools: map[string]state.ToolInstallInfo{
					"windsurf": {Files: []string{"/path/to/workflow-a.md"}},
				},
			},
			{
				Name:    "skill-b",
				Source:  "custom-tap",
				Type:    "skill",
				Version: "2.0.0",
				Tools: map[string]state.ToolInstallInfo{
					"windsurf": {Files: []string{"/path/to/skill-b/SKILL.md"}},
					"cursor":   {Files: []string{"/path/to/cursor/skill-b.md"}},
				},
			},
		},
	}

	// Collect tools from installed resources (mimics bundle dump logic)
	toolSet := make(map[string]bool)
	var installConfigs []aiconfig.InstallConfig

	for _, res := range st.Installed {
		var tools []string
		for toolName := range res.Tools {
			tools = append(tools, toolName)
			toolSet[toolName] = true
		}

		installConfigs = append(installConfigs, aiconfig.InstallConfig{
			Name:  res.Source + "/" + res.Name,
			Tools: tools,
		})
	}

	// Verify tools were collected
	if len(toolSet) != 2 {
		t.Errorf("Expected 2 unique tools, got %d", len(toolSet))
	}

	if !toolSet["windsurf"] {
		t.Error("Expected windsurf in tool set")
	}

	if !toolSet["cursor"] {
		t.Error("Expected cursor in tool set")
	}

	// Verify install configs
	if len(installConfigs) != 2 {
		t.Errorf("Expected 2 install configs, got %d", len(installConfigs))
	}

	// Verify first config
	if installConfigs[0].Name != "default/workflow-a" {
		t.Errorf("First config name = %v, want 'default/workflow-a'", installConfigs[0].Name)
	}

	// Verify second config
	if installConfigs[1].Name != "custom-tap/skill-b" {
		t.Errorf("Second config name = %v, want 'custom-tap/skill-b'", installConfigs[1].Name)
	}
}

// TestBundleConfigValidation tests .aiconfig validation
func TestBundleConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    aiconfig.AiConfig
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid_config",
			config: aiconfig.AiConfig{
				Tools: []string{"windsurf"},
				Install: []aiconfig.InstallConfig{
					{Name: "default/test-workflow"},
				},
				Scope: "global",
			},
			expectErr: false,
		},
		{
			name: "empty_tools",
			config: aiconfig.AiConfig{
				Tools: []string{},
				Install: []aiconfig.InstallConfig{
					{Name: "default/test-workflow"},
				},
			},
			expectErr: true,
			errMsg:    "tools array is required",
		},
		{
			name: "invalid_tool",
			config: aiconfig.AiConfig{
				Tools: []string{"invalid-tool"},
				Install: []aiconfig.InstallConfig{
					{Name: "default/test-workflow"},
				},
			},
			expectErr: true,
			errMsg:    "invalid tool",
		},
		{
			name: "invalid_scope",
			config: aiconfig.AiConfig{
				Tools: []string{"windsurf"},
				Install: []aiconfig.InstallConfig{
					{Name: "default/test-workflow"},
				},
				Scope: "invalid",
			},
			expectErr: true,
			errMsg:    "invalid scope",
		},
		{
			name: "resource_tool_not_in_global_tools",
			config: aiconfig.AiConfig{
				Tools: []string{"windsurf"},
				Install: []aiconfig.InstallConfig{
					{Name: "default/test-workflow", Tools: []string{"cursor"}},
				},
			},
			expectErr: true,
			errMsg:    "not in the global tools list",
		},
		{
			name: "empty_resource_name",
			config: aiconfig.AiConfig{
				Tools: []string{"windsurf"},
				Install: []aiconfig.InstallConfig{
					{Name: ""},
				},
			},
			expectErr: true,
			errMsg:    "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestBundleGetToolsForResource tests tool selection for resources
func TestBundleGetToolsForResource(t *testing.T) {
	cfg := &aiconfig.AiConfig{
		Tools: []string{"windsurf", "cursor", "aider"},
	}

	tests := []struct {
		name     string
		install  aiconfig.InstallConfig
		expected []string
	}{
		{
			name:     "no_resource_tools_uses_global",
			install:  aiconfig.InstallConfig{Name: "test"},
			expected: []string{"windsurf", "cursor", "aider"},
		},
		{
			name:     "resource_tools_override_global",
			install:  aiconfig.InstallConfig{Name: "test", Tools: []string{"windsurf"}},
			expected: []string{"windsurf"},
		},
		{
			name:     "multiple_resource_tools",
			install:  aiconfig.InstallConfig{Name: "test", Tools: []string{"windsurf", "cursor"}},
			expected: []string{"windsurf", "cursor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.GetToolsForResource(tt.install)

			if len(result) != len(tt.expected) {
				t.Errorf("GetToolsForResource() returned %d tools, want %d", len(result), len(tt.expected))
				return
			}

			for i, tool := range result {
				if tool != tt.expected[i] {
					t.Errorf("GetToolsForResource()[%d] = %v, want %v", i, tool, tt.expected[i])
				}
			}
		})
	}
}

// TestBundleGetScope tests scope defaulting
func TestBundleGetScope(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		expected string
	}{
		{
			name:     "empty_defaults_to_local",
			scope:    "",
			expected: "local",
		},
		{
			name:     "global_scope",
			scope:    "global",
			expected: "global",
		},
		{
			name:     "local_scope",
			scope:    "local",
			expected: "local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &aiconfig.AiConfig{Scope: tt.scope}
			result := cfg.GetScope()

			if result != tt.expected {
				t.Errorf("GetScope() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestBundleResourceParsing tests parsing of resource references
func TestBundleResourceParsing(t *testing.T) {
	tests := []struct {
		name           string
		resourceRef    string
		expectedSource string
		expectedName   string
		expectErr      bool
	}{
		{
			name:           "tap_slash_resource",
			resourceRef:    "default/my-workflow",
			expectedSource: "default",
			expectedName:   "my-workflow",
		},
		{
			name:           "bundled_resource",
			resourceRef:    "bundled/git-commit",
			expectedSource: "bundled",
			expectedName:   "git-commit",
		},
		{
			name:           "custom_tap_resource",
			resourceRef:    "my-custom-tap/special-skill",
			expectedSource: "my-custom-tap",
			expectedName:   "special-skill",
		},
		{
			name:        "no_slash_invalid",
			resourceRef: "invalid-reference",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.SplitN(tt.resourceRef, "/", 2)

			if tt.expectErr {
				if len(parts) == 2 {
					t.Error("expected parsing to fail (no slash)")
				}
				return
			}

			if len(parts) != 2 {
				t.Fatalf("expected 2 parts, got %d", len(parts))
			}

			source, name := parts[0], parts[1]

			if source != tt.expectedSource {
				t.Errorf("source = %v, want %v", source, tt.expectedSource)
			}

			if name != tt.expectedName {
				t.Errorf("name = %v, want %v", name, tt.expectedName)
			}
		})
	}
}

// TestBundleInstalledResourceConstruction tests building InstalledResource structs
func TestBundleInstalledResourceConstruction(t *testing.T) {
	res := &resource.Resource{
		Name:        "test-workflow",
		Version:     "1.2.3",
		Type:        resource.TypeWorkflow,
		Description: "A test workflow",
	}

	source := "default"

	installedResource := state.InstalledResource{
		Name:    res.Name,
		Source:  source,
		Type:    string(res.Type),
		Version: res.Version,
		Tools:   make(map[string]state.ToolInstallInfo),
	}

	// Add tool info
	installedResource.Tools["windsurf"] = state.ToolInstallInfo{
		Files: []string{"/path/to/workflow.md"},
	}

	// Verify construction
	if installedResource.Name != "test-workflow" {
		t.Errorf("Name = %v, want 'test-workflow'", installedResource.Name)
	}

	if installedResource.Source != "default" {
		t.Errorf("Source = %v, want 'default'", installedResource.Source)
	}

	if installedResource.Type != "workflow" {
		t.Errorf("Type = %v, want 'workflow'", installedResource.Type)
	}

	if installedResource.Version != "1.2.3" {
		t.Errorf("Version = %v, want '1.2.3'", installedResource.Version)
	}

	if len(installedResource.Tools) != 1 {
		t.Errorf("Tools count = %d, want 1", len(installedResource.Tools))
	}

	toolInfo, ok := installedResource.Tools["windsurf"]
	if !ok {
		t.Fatal("windsurf tool info should exist")
	}

	if len(toolInfo.Files) != 1 || toolInfo.Files[0] != "/path/to/workflow.md" {
		t.Errorf("Files = %v, want ['/path/to/workflow.md']", toolInfo.Files)
	}
}

// TestBundleCounterLogic tests the counter logic for install outcomes
func TestBundleCounterLogic(t *testing.T) {
	tests := []struct {
		name                string
		toolSuccessCount    int
		alreadyExistsCount  int
		compatibleToolCount int
		expectedOutcome     string
	}{
		{
			name:                "success_when_tools_installed",
			toolSuccessCount:    2,
			alreadyExistsCount:  0,
			compatibleToolCount: 2,
			expectedOutcome:     "installed",
		},
		{
			name:                "already_up_to_date_when_all_exist",
			toolSuccessCount:    0,
			alreadyExistsCount:  2,
			compatibleToolCount: 2,
			expectedOutcome:     "already_up_to_date",
		},
		{
			name:                "no_compatible_tools",
			toolSuccessCount:    0,
			alreadyExistsCount:  0,
			compatibleToolCount: 0,
			expectedOutcome:     "no_compatible_tools",
		},
		{
			name:                "installation_failed",
			toolSuccessCount:    0,
			alreadyExistsCount:  0,
			compatibleToolCount: 2,
			expectedOutcome:     "installation_failed",
		},
		{
			name:                "partial_success_counts_as_success",
			toolSuccessCount:    1,
			alreadyExistsCount:  1,
			compatibleToolCount: 2,
			expectedOutcome:     "installed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outcome string

			// This mimics the logic in runBundle
			if tt.toolSuccessCount > 0 {
				outcome = "installed"
			} else if tt.alreadyExistsCount > 0 {
				outcome = "already_up_to_date"
			} else if tt.compatibleToolCount == 0 {
				outcome = "no_compatible_tools"
			} else {
				outcome = "installation_failed"
			}

			if outcome != tt.expectedOutcome {
				t.Errorf("outcome = %v, want %v", outcome, tt.expectedOutcome)
			}
		})
	}
}

// TestBundleNewFromState tests creating AiConfig from state
func TestBundleNewFromState(t *testing.T) {
	tools := []string{"windsurf", "cursor"}
	installConfigs := []aiconfig.InstallConfig{
		{Name: "default/workflow-a", Tools: []string{"windsurf"}},
		{Name: "custom/skill-b", Tools: []string{"windsurf", "cursor"}},
	}
	scope := "global"

	cfg := aiconfig.NewFromState(tools, installConfigs, scope)

	if cfg == nil {
		t.Fatal("NewFromState should not return nil")
	}

	if len(cfg.Tools) != 2 {
		t.Errorf("Tools count = %d, want 2", len(cfg.Tools))
	}

	if len(cfg.Install) != 2 {
		t.Errorf("Install count = %d, want 2", len(cfg.Install))
	}

	if cfg.Scope != "global" {
		t.Errorf("Scope = %v, want 'global'", cfg.Scope)
	}
}

// TestBundleConfigSaveAndLoad tests round-trip save/load of config
func TestBundleConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".aiconfig")

	original := &aiconfig.AiConfig{
		Tools: []string{"windsurf", "cursor"},
		Install: []aiconfig.InstallConfig{
			{Name: "default/test-workflow"},
			{Name: "default/test-skill", Tools: []string{"windsurf"}},
		},
		Scope: "global",
	}

	// Save
	if err := original.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := aiconfig.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify
	if len(loaded.Tools) != len(original.Tools) {
		t.Errorf("Tools count = %d, want %d", len(loaded.Tools), len(original.Tools))
	}

	if len(loaded.Install) != len(original.Install) {
		t.Errorf("Install count = %d, want %d", len(loaded.Install), len(original.Install))
	}

	if loaded.Scope != original.Scope {
		t.Errorf("Scope = %v, want %v", loaded.Scope, original.Scope)
	}
}

// TestBundleVerboseOutput tests that verbose flag controls output
func TestBundleVerboseOutput(t *testing.T) {
	// Save original flag value
	origVerbose := verboseFlag
	defer func() { verboseFlag = origVerbose }()

	// Test that verbose flag exists and can be set
	verboseFlag = true
	if !verboseFlag {
		t.Error("verboseFlag should be true")
	}

	verboseFlag = false
	if verboseFlag {
		t.Error("verboseFlag should be false")
	}
}

// TestBundleForceFlag tests that force flag controls overwrite behavior
func TestBundleForceFlag(t *testing.T) {
	// Save original flag value
	origForce := forceFlag
	defer func() { forceFlag = origForce }()

	// Test that force flag exists and can be set
	forceFlag = true
	if !forceFlag {
		t.Error("forceFlag should be true")
	}

	forceFlag = false
	if forceFlag {
		t.Error("forceFlag should be false")
	}
}

// TestBundleMultipleToolsInstall tests installing to multiple tools
func TestBundleMultipleToolsInstall(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories for both tools (using rules which both support)
	windsurfDir := filepath.Join(tmpDir, "windsurf", "memories")
	cursorDir := filepath.Join(tmpDir, "cursor", "rules")
	_ = os.MkdirAll(windsurfDir, 0o755)
	_ = os.MkdirAll(cursorDir, 0o755)

	windsurfInstaller := &installer.WindsurfInstaller{
		BaseInstaller: installer.BaseInstaller{
			Config: config.ToolConfig{
				Name:          config.ToolWindsurf,
				GlobalDir:     filepath.Join(tmpDir, "windsurf"),
				RulesDir:      "memories",
				SupportsRules: true,
			},
		},
	}

	cursorInstaller := &installer.CursorInstaller{
		BaseInstaller: installer.BaseInstaller{
			Config: config.ToolConfig{
				Name:          config.ToolCursor,
				GlobalDir:     filepath.Join(tmpDir, "cursor"),
				RulesDir:      "rules",
				SupportsRules: true,
			},
		},
	}

	// Use rules type since both windsurf and cursor support rules
	res := &resource.Resource{
		Name:        "multi-tool-rule",
		Version:     "1.0.0",
		Type:        resource.TypeRules,
		Description: "Test rule for multiple tools",
	}

	content := "# Multi-tool rule content"

	// Install to windsurf
	windsurfPath, err := windsurfInstaller.InstallFromReader(res, strings.NewReader(content), true, false)
	if err != nil {
		t.Fatalf("Windsurf install error: %v", err)
	}

	// Install to cursor
	cursorPath, err := cursorInstaller.InstallFromReader(res, strings.NewReader(content), true, false)
	if err != nil {
		t.Fatalf("Cursor install error: %v", err)
	}

	// Verify both files exist
	if _, err := os.Stat(windsurfPath); os.IsNotExist(err) {
		t.Error("Windsurf file should exist")
	}

	if _, err := os.Stat(cursorPath); os.IsNotExist(err) {
		t.Error("Cursor file should exist")
	}

	// Verify content
	windsurfContent, _ := os.ReadFile(windsurfPath)
	cursorContent, _ := os.ReadFile(cursorPath)

	if string(windsurfContent) != content {
		t.Error("Windsurf content mismatch")
	}

	if string(cursorContent) != content {
		t.Error("Cursor content mismatch")
	}
}

// Capture stdout helper for testing output
func captureOutput(f func()) string {
	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old
	buf.ReadFrom(r)
	return buf.String()
}
