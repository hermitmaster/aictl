package cli

import (
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// Version info - set via ldflags at build time
	version   = "0.1.0"
	commit    = "dev"
	buildDate = "unknown"

	// Global flags
	toolFlag    string
	globalFlag  bool
	localFlag   bool
	forceFlag   bool
	verboseFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "aictl",
	Short: "AI coding assistant resource manager",
	Long: `aictl is a registry-based CLI tool for installing AI coding assistant
resources (rules, workflows, skills) across multiple tools like Windsurf,
Cursor, and Aider.

Inspired by Homebrew, aictl supports bundled resources and custom taps
(Git repositories) for sharing resources across teams.

Quick Start:
  aictl search                     # Find available resources
  aictl install bundled/jira-context --tool=windsurf
  aictl list                       # Show installed resources
  aictl bundle                     # Install from .aiconfig

Tap Management:
  aictl tap add myco https://github.com/myco/ai-resources
  aictl install myco/custom-rules`,
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}

// GetVersion returns the full version string
func GetVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s, %s/%s)",
		version, commit, buildDate, runtime.GOOS, runtime.GOARCH)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&toolFlag, "tool", "t", "", "Target tool(s): windsurf, cursor, aider, or comma-separated list (default: from .aiconfig or all detected)")
	rootCmd.PersistentFlags().BoolVarP(&globalFlag, "global", "g", false, "Install to global config directory")
	rootCmd.PersistentFlags().BoolVarP(&localFlag, "local", "l", false, "Install to project-local config directory")
	rootCmd.PersistentFlags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing files")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")

	rootCmd.SetVersionTemplate(fmt.Sprintf("%s version %s\n", color.CyanString("aictl"), color.GreenString(GetVersion())))
}
