package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/hermitmaster/aictl/internal/config"
)

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "List supported tools and their status",
	Long: `List all supported AI coding assistant tools and their installation status.

Shows the global and local configuration directories for each tool.

Examples:
  aictl tools`,
	RunE: runTools,
}

func init() {
	rootCmd.AddCommand(toolsCmd)
}

func runTools(cmd *cobra.Command, args []string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(w, "TOOL\tSTATUS\tGLOBAL CONFIG\tLOCAL CONFIG\tSUPPORTS")
	_, _ = fmt.Fprintln(w, "----\t------\t-------------\t------------\t--------")

	for _, tool := range config.AllTools() {
		cfg := config.GetToolConfig(tool)

		status := "not found"
		statusColor := color.YellowString
		if config.IsToolInstalled(tool) {
			status = "detected"
			statusColor = color.GreenString
		}

		// Build supports string
		var supports []string
		if cfg.SupportsRules {
			supports = append(supports, "rules")
		}
		if cfg.SupportsWorkflows {
			supports = append(supports, "workflows")
		}
		if cfg.SupportsSkills {
			supports = append(supports, "skills")
		}
		if cfg.SupportsBin {
			supports = append(supports, "bin")
		}

		supportsStr := ""
		for i, s := range supports {
			if i > 0 {
				supportsStr += ", "
			}
			supportsStr += s
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			tool,
			statusColor(status),
			cfg.GlobalDir,
			cfg.LocalDir,
			supportsStr,
		)
	}

	_ = w.Flush()
	return nil
}
