package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/installer"
	"github.com/hermitmaster/aictl/internal/state"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <resource>",
	Short: "Uninstall a resource",
	Long: `Uninstall a resource from one or more AI coding assistant tools.

Examples:
  aictl uninstall jira-context
  aictl uninstall typescript-rules --tool=cursor
  aictl uninstall code-review --local`,
	Args: cobra.ExactArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	resourceName := args[0]

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
	cfg := config.DefaultConfig()
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

	// Find the installed resource
	installed := st.GetInstalled(resourceName)
	if installed == nil {
		return fmt.Errorf("resource not found: %s (scope: %s)", resourceName, scopeStr)
	}

	// Determine target tools
	var targetTools []config.Tool
	if toolFlag != "" {
		for _, t := range strings.Split(toolFlag, ",") {
			t = strings.TrimSpace(t)
			if t == "all" {
				for toolName := range installed.Tools {
					targetTools = append(targetTools, config.Tool(toolName))
				}
				break
			}
			targetTools = append(targetTools, config.Tool(t))
		}
	} else {
		// Uninstall from all tools where it's installed
		for toolName := range installed.Tools {
			targetTools = append(targetTools, config.Tool(toolName))
		}
	}

	if len(targetTools) == 0 {
		return fmt.Errorf("resource %s is not installed in any tools", resourceName)
	}

	// Uninstall from each tool
	successCount := 0
	for _, tool := range targetTools {
		toolInfo, exists := installed.Tools[string(tool)]
		if !exists {
			if verboseFlag {
				color.Yellow("⚠ Skipping %s: resource not installed there", tool)
			}
			continue
		}

		inst, err := installer.GetInstaller(tool)
		if err != nil {
			color.Yellow("⚠ Skipping %s: %v", tool, err)
			continue
		}

		if err := inst.Uninstall(nil, toolInfo.Files, global); err != nil {
			color.Red("✗ Failed to uninstall from %s: %v", tool, err)
			continue
		}

		// Remove tool from installed resource
		delete(installed.Tools, string(tool))

		color.Green("✓ Uninstalled %s from %s (%s)", resourceName, tool, scopeStr)
		if verboseFlag {
			for _, f := range toolInfo.Files {
				fmt.Printf("  → removed %s\n", f)
			}
		}
		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("failed to uninstall resource from any tool")
	}

	// Update state
	if len(installed.Tools) == 0 {
		st.RemoveInstalled(resourceName)
	}

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
