package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/installer"
	"github.com/hermitmaster/aictl/internal/resource"
	"github.com/hermitmaster/aictl/internal/state"
	"github.com/hermitmaster/aictl/internal/tap"
)

const defaultTapName = "default"

var installCmd = &cobra.Command{
	Use:   "install <resource>",
	Short: "Install a resource",
	Long: `Install a resource to one or more AI coding assistant tools.

Resources can be specified as:
  - <name>             Install from the default registry
  - <tap>/<name>       Install from a custom tap

Examples:
  aictl install jira-context
  aictl install typescript-rules --tool=cursor
  aictl install mycompany/internal-standards --tool=windsurf,cursor`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	resourceRef := args[0]

	// Parse resource reference
	var source, name string
	parts := strings.SplitN(resourceRef, "/", 2)
	if len(parts) == 2 {
		source, name = parts[0], parts[1]
	} else {
		// No prefix - use default registry
		source, name = defaultTapName, resourceRef
	}

	// Load the resource based on source
	var res *resource.Resource
	var content string
	var err error

	// Load from a tap
	cfg := config.DefaultConfig()
	tapMgr, err := tap.NewManager(cfg)
	if err != nil {
		return fmt.Errorf("error initializing tap manager: %w", err)
	}

	// Auto-add default registry if needed
	if source == defaultTapName && !tapMgr.Exists(defaultTapName) {
		fmt.Printf("Adding default registry...\n")
		if _, err := tapMgr.Add(defaultTapName, config.DefaultRegistry); err != nil {
			return fmt.Errorf("error adding default registry: %w", err)
		}
	}

	if !tapMgr.Exists(source) {
		return fmt.Errorf("unknown source '%s' (not a tap, use 'aictl tap add' to add it)", source)
	}

	res, content, err = tapMgr.LoadResource(source, name)
	if err != nil {
		return fmt.Errorf("error loading resource from tap '%s': %w", source, err)
	}

	// Validate the resource
	if err := res.Validate(); err != nil {
		return fmt.Errorf("invalid resource: %w", err)
	}

	// Determine target tools
	tools, err := getTargetTools(res)
	if err != nil {
		return err
	}

	if len(tools) == 0 {
		return fmt.Errorf("no compatible tools found for resource type: %s", res.Type)
	}

	// Determine scope (global vs local)
	global := !localFlag
	if globalFlag {
		global = true
	}

	scopeStr := "global"
	if !global {
		scopeStr = "local"
	}

	// Initialize state manager
	stateMgr, err := state.NewStateManager(cfg)
	if err != nil {
		return fmt.Errorf("error initializing state: %w", err)
	}

	// Load appropriate state
	var st *state.State
	if global {
		st, err = stateMgr.LoadGlobal()
	} else {
		st, err = stateMgr.LoadLocal()
	}
	if err != nil {
		return fmt.Errorf("error loading state: %w", err)
	}

	// Install to each tool
	installedResource := state.InstalledResource{
		Name:        res.Name,
		Source:      source,
		Type:        string(res.Type),
		Version:     res.Version,
		InstalledAt: time.Now(),
		Tools:       make(map[string]state.ToolInstallInfo),
	}

	successCount := 0
	for _, tool := range tools {
		inst, err := installer.GetInstaller(tool)
		if err != nil {
			color.Yellow("⚠ Skipping %s: %v", tool, err)
			continue
		}

		if !inst.SupportsType(res.Type) {
			if verboseFlag {
				color.Yellow("⚠ Skipping %s: does not support %s resources", tool, res.Type)
			}
			continue
		}

		// Install using the appropriate installer
		var destPath string
		var installErr error

		switch typedInst := inst.(type) {
		case *installer.WindsurfInstaller:
			destPath, installErr = typedInst.InstallFromReader(res, strings.NewReader(content), global, forceFlag)
		case *installer.CursorInstaller:
			destPath, installErr = typedInst.InstallFromReader(res, strings.NewReader(content), global, forceFlag)
		case *installer.AiderInstaller:
			destPath, installErr = typedInst.InstallFromReader(res, strings.NewReader(content), global, forceFlag)
		case *installer.ContinueInstaller:
			destPath, installErr = typedInst.InstallFromReader(res, strings.NewReader(content), global, forceFlag)
		case *installer.ClaudeCodeInstaller:
			destPath, installErr = typedInst.InstallFromReader(res, strings.NewReader(content), global, forceFlag)
		default:
			if verboseFlag {
				color.Yellow("⚠ Skipping %s: installer type not supported", tool)
			}
			continue
		}

		if installErr != nil {
			color.Red("✗ Failed to install to %s: %v", tool, installErr)
			continue
		}

		installedResource.Tools[string(tool)] = state.ToolInstallInfo{
			Files: []string{destPath},
		}

		color.Green("✓ Installed %s to %s (%s)", res.Name, tool, scopeStr)
		if verboseFlag {
			fmt.Printf("  → %s\n", destPath)
		}
		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("failed to install resource to any tool")
	}

	// Update state
	st.AddInstalled(installedResource)
	if global {
		if err := stateMgr.SaveGlobal(st); err != nil {
			return fmt.Errorf("error saving state: %w", err)
		}
	} else {
		if err := stateMgr.SaveLocal(st); err != nil {
			return fmt.Errorf("error saving state: %w", err)
		}
	}

	return nil
}

func getTargetTools(res *resource.Resource) ([]config.Tool, error) {
	// If --tool flag is specified, use that
	if toolFlag != "" {
		var tools []config.Tool
		for _, t := range strings.Split(toolFlag, ",") {
			t = strings.TrimSpace(t)
			if t == "all" {
				return config.DetectInstalledTools(), nil
			}
			tools = append(tools, config.Tool(t))
		}
		return tools, nil
	}

	// If resource specifies tools, use those
	if len(res.Tools) > 0 {
		var tools []config.Tool
		for _, t := range res.Tools {
			tools = append(tools, config.Tool(t))
		}
		return tools, nil
	}

	// Default: detect installed tools
	return config.DetectInstalledTools(), nil
}
