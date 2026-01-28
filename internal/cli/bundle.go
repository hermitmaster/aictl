package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/hermitmaster/aictl/internal/aiconfig"
	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/installer"
	"github.com/hermitmaster/aictl/internal/resource"
	"github.com/hermitmaster/aictl/internal/state"
	"github.com/hermitmaster/aictl/internal/tap"
	"github.com/hermitmaster/aictl/resources"
)

var bundleCmd = &cobra.Command{
	Use:   "bundle [path/to/.aiconfig]",
	Short: "Install resources from a .aiconfig file",
	Long: `Install all resources defined in a .aiconfig file.

If no path is specified, searches for .aiconfig in the current directory.

Examples:
  aictl bundle                     # Uses ./.aiconfig
  aictl bundle path/to/.aiconfig   # Use specific file
  aictl bundle --dry-run           # Preview changes`,
	RunE: runBundle,
}

var bundleDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Export current state to .aiconfig format",
	Long: `Export the currently installed resources to .aiconfig format.

Examples:
  aictl bundle dump                # Output to stdout
  aictl bundle dump > .aiconfig    # Save to file`,
	RunE: runBundleDump,
}

var dryRunFlag bool

func init() {
	bundleCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview changes without installing")
	bundleCmd.AddCommand(bundleDumpCmd)
	rootCmd.AddCommand(bundleCmd)
}

func runBundle(cmd *cobra.Command, args []string) error {
	// Load config
	var configPath string
	if len(args) > 0 {
		configPath = args[0]
	}

	cfg, foundPath, err := aiconfig.LoadOrFind(configPath)
	if err != nil {
		return err
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid .aiconfig: %w", err)
	}

	fmt.Printf("Using config: %s\n", foundPath)
	fmt.Printf("Tools: %s\n", strings.Join(cfg.Tools, ", "))
	fmt.Printf("Scope: %s\n", cfg.GetScope())
	fmt.Println()

	if dryRunFlag {
		color.Cyan("=== Dry Run (no changes will be made) ===\n")
	}

	// Determine scope
	global := cfg.GetScope() == "global"
	scopeStr := cfg.GetScope()

	// Initialize app config and tap manager
	appCfg := config.DefaultConfig()

	// Auto-add and update taps declared in config
	tapMgr, err := tap.NewManager(appCfg)
	if err != nil {
		return fmt.Errorf("error initializing tap manager: %w", err)
	}

	// Add missing taps from config
	for _, tapCfg := range cfg.Taps {
		if !tapMgr.Exists(tapCfg.Name) {
			if dryRunFlag {
				fmt.Printf("Would add tap: %s (%s)\n", tapCfg.Name, tapCfg.URL)
			} else {
				fmt.Printf("Adding tap: %s (%s)\n", tapCfg.Name, tapCfg.URL)
				if _, err := tapMgr.Add(tapCfg.Name, tapCfg.URL); err != nil {
					color.Yellow("⚠ Failed to add tap %s: %v", tapCfg.Name, err)
				} else {
					color.Green("✓ Added tap: %s", tapCfg.Name)
				}
			}
		}
	}

	// Update all taps to get latest resources
	if !dryRunFlag {
		fmt.Println("Updating taps...")
		if err := tapMgr.UpdateAll(); err != nil {
			color.Yellow("⚠ Failed to update taps: %v", err)
		}
	}
	fmt.Println()

	// Initialize state manager
	stateMgr, err := state.NewStateManager(appCfg)
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

	// Process each install entry
	successCount := 0
	skipCount := 0
	failCount := 0

	for _, inst := range cfg.Install {
		// Parse resource reference
		parts := strings.SplitN(inst.Name, "/", 2)
		if len(parts) != 2 {
			color.Red("✗ Invalid resource reference: %s", inst.Name)
			failCount++
			continue
		}

		source, name := parts[0], parts[1]

		// Load the resource based on source
		var res *resource.Resource
		var content string
		var loadErr error

		if source == "bundled" {
			res, content, loadErr = resources.LoadBundled(name)
		} else {
			// Try to load from a tap
			tapMgr, err := tap.NewManager(appCfg)
			if err != nil {
				color.Red("✗ Error initializing tap manager: %v", err)
				failCount++
				continue
			}

			if !tapMgr.Exists(source) {
				color.Red("✗ Unknown source '%s' (not a tap)", source)
				failCount++
				continue
			}

			res, content, loadErr = tapMgr.LoadResource(source, name)
		}

		if loadErr != nil {
			color.Red("✗ Error loading %s: %v", inst.Name, loadErr)
			failCount++
			continue
		}

		// Get tools for this resource
		tools := cfg.GetToolsForResource(inst)

		if dryRunFlag {
			fmt.Printf("Would install %s to: %s\n", inst.Name, strings.Join(tools, ", "))
			successCount++
			continue
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

		toolSuccessCount := 0
		for _, toolName := range tools {
			tool := config.Tool(toolName)
			inst, err := installer.GetInstaller(tool)
			if err != nil {
				if verboseFlag {
					color.Yellow("  ⚠ Skipping %s: %v", tool, err)
				}
				continue
			}

			if !inst.SupportsType(res.Type) {
				if verboseFlag {
					color.Yellow("  ⚠ Skipping %s: does not support %s resources", tool, res.Type)
				}
				continue
			}

			// Install using the appropriate installer
			var installedFiles []string
			var installErr error

			// Use full Install method for multi-file resources (which have SourcePath set)
			// Use InstallFromReader for bundled resources (content passed as string)
			if res.IsMultiFile() && res.SourcePath != "" {
				installedFiles, installErr = inst.Install(res, global, forceFlag)
			} else {
				var destPath string
				switch typedInst := inst.(type) {
				case *installer.WindsurfInstaller:
					destPath, installErr = typedInst.InstallFromReader(res, strings.NewReader(content), global, forceFlag)
				case *installer.CursorInstaller:
					destPath, installErr = typedInst.InstallFromReader(res, strings.NewReader(content), global, forceFlag)
				case *installer.AiderInstaller:
					destPath, installErr = typedInst.InstallFromReader(res, strings.NewReader(content), global, forceFlag)
				case *installer.ContinueInstaller:
					destPath, installErr = typedInst.InstallFromReader(res, strings.NewReader(content), global, forceFlag)
				default:
					continue
				}
				if destPath != "" {
					installedFiles = []string{destPath}
				}
			}

			if installErr != nil {
				if verboseFlag {
					color.Yellow("  ⚠ Failed to install to %s: %v", tool, installErr)
				}
				continue
			}

			installedResource.Tools[string(tool)] = state.ToolInstallInfo{
				Files: installedFiles,
			}

			if verboseFlag {
				fmt.Printf("  → %s: %v\n", tool, installedFiles)
			}
			toolSuccessCount++
		}

		if toolSuccessCount > 0 {
			color.Green("✓ Installed %s to %d tool(s) (%s)", res.Name, toolSuccessCount, scopeStr)
			st.AddInstalled(installedResource)
			successCount++
		} else {
			color.Yellow("⚠ Skipped %s: no compatible tools", res.Name)
			skipCount++
		}
	}

	// Save state
	if !dryRunFlag && successCount > 0 {
		if global {
			if err := stateMgr.SaveGlobal(st); err != nil {
				return fmt.Errorf("error saving state: %w", err)
			}
		} else {
			if err := stateMgr.SaveLocal(st); err != nil {
				return fmt.Errorf("error saving state: %w", err)
			}
		}
	}

	// Summary
	fmt.Println()
	if dryRunFlag {
		fmt.Printf("Dry run complete: %d resource(s) would be installed\n", successCount)
	} else {
		fmt.Printf("Bundle complete: %d installed, %d skipped, %d failed\n", successCount, skipCount, failCount)
	}

	return nil
}

func runBundleDump(cmd *cobra.Command, args []string) error {
	// Initialize state manager
	appCfg := config.DefaultConfig()
	stateMgr, err := state.NewStateManager(appCfg)
	if err != nil {
		return fmt.Errorf("error initializing state: %w", err)
	}

	// Determine which state to dump
	var st *state.State
	var scope string

	if localFlag {
		st, err = stateMgr.LoadLocal()
		scope = "local"
	} else {
		st, err = stateMgr.LoadGlobal()
		scope = "global"
	}
	if err != nil {
		return fmt.Errorf("error loading state: %w", err)
	}

	// Collect tools from installed resources
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

	// Build tools list
	var tools []string
	for t := range toolSet {
		tools = append(tools, t)
	}

	// Create config
	cfg := aiconfig.NewFromState(tools, installConfigs, scope)

	// Output as YAML
	if err := cfg.Save("/dev/stdout"); err != nil {
		// Fallback: manual output
		fmt.Println("# .aiconfig")
		fmt.Println("# Generated by aictl bundle dump")
		fmt.Println()
		fmt.Println("tools:")
		for _, t := range tools {
			fmt.Printf("  - %s\n", t)
		}
		fmt.Println()
		fmt.Println("install:")
		for _, inst := range installConfigs {
			fmt.Printf("  - name: %s\n", inst.Name)
			if len(inst.Tools) > 0 {
				fmt.Printf("    tools: [%s]\n", strings.Join(inst.Tools, ", "))
			}
		}
		fmt.Println()
		fmt.Printf("scope: %s\n", scope)
	}

	return nil
}
