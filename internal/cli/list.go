package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/state"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed resources",
	Long: `List all installed resources.

Examples:
  aictl list
  aictl list --global
  aictl list --local
  aictl list --tool=windsurf`,
	RunE: runList,
}

var allFlag bool

func init() {
	listCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Show both global and local resources")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	cfg := config.DefaultConfig()
	stateMgr, err := state.NewStateManager(cfg)
	if err != nil {
		return fmt.Errorf("error initializing state: %w", err)
	}

	// Determine which states to show
	showGlobal := globalFlag || (!localFlag && !allFlag) || allFlag
	showLocal := localFlag || allFlag

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if showGlobal {
		globalState, err := stateMgr.LoadGlobal()
		if err != nil {
			return fmt.Errorf("error loading global state: %w", err)
		}

		if len(globalState.Installed) > 0 || !showLocal {
			if allFlag {
				color.Cyan("\n=== Global Resources ===\n")
			}
			printResources(w, globalState, "global")
		}
	}

	if showLocal {
		localState, err := stateMgr.LoadLocal()
		if err != nil {
			return fmt.Errorf("error loading local state: %w", err)
		}

		if len(localState.Installed) > 0 || !showGlobal {
			if allFlag {
				color.Cyan("\n=== Local Resources ===\n")
			}
			printResources(w, localState, "local")
		}
	}

	_ = w.Flush()
	return nil
}

func printResources(w *tabwriter.Writer, st *state.State, scope string) {
	if len(st.Installed) == 0 {
		color.Yellow("No resources installed (%s)\n", scope)
		return
	}

	// Filter by tool if specified
	filterTool := ""
	if toolFlag != "" && toolFlag != "all" {
		filterTool = toolFlag
	}

	_, _ = fmt.Fprintln(w, "NAME\tSOURCE\tTYPE\tVERSION\tTOOLS\tSCOPE")
	_, _ = fmt.Fprintln(w, "----\t------\t----\t-------\t-----\t-----")

	for _, res := range st.Installed {
		// Get tools list
		var tools []string
		for toolName := range res.Tools {
			if filterTool != "" && toolName != filterTool {
				continue
			}
			tools = append(tools, toolName)
		}

		// Skip if no matching tools
		if filterTool != "" && len(tools) == 0 {
			continue
		}

		toolsStr := strings.Join(tools, ",")
		if len(tools) == 0 {
			toolsStr = "-"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			res.Name,
			res.Source,
			res.Type,
			res.Version,
			toolsStr,
			scope,
		)
	}
}
