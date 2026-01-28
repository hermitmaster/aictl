package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/tap"
)

var tapCmd = &cobra.Command{
	Use:   "tap",
	Short: "Manage custom resource registries (taps)",
	Long: `Manage custom resource registries (taps).

Taps are Git repositories containing AI coding assistant resources
that follow the aictl resource format.

Examples:
  aictl tap add mycompany https://github.com/mycompany/ai-resources
  aictl tap list
  aictl tap update mycompany
  aictl tap remove mycompany`,
}

var tapAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a new tap",
	Long: `Add a new tap by cloning a Git repository.

The repository should contain resources in the standard directory structure:
  rules/      - Rule files (.md with frontmatter)
  workflows/  - Workflow files (.md with frontmatter)
  skills/     - Skill files (.md with frontmatter)
  bin/        - Multi-file resources (directories with resource.yaml)

Examples:
  aictl tap add mycompany https://github.com/mycompany/ai-resources
  aictl tap add internal git@github.com:mycompany/internal-resources.git`,
	Args: cobra.ExactArgs(2),
	RunE: runTapAdd,
}

var tapRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a tap",
	Long: `Remove a tap and delete its local cache.

Examples:
  aictl tap remove mycompany`,
	Args: cobra.ExactArgs(1),
	RunE: runTapRemove,
}

var tapListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed taps",
	Long: `List all installed taps.

Examples:
  aictl tap list`,
	RunE: runTapList,
}

var tapUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update tap(s)",
	Long: `Update one or all taps by pulling the latest changes.

Examples:
  aictl tap update           # Update all taps
  aictl tap update mycompany # Update specific tap`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTapUpdate,
}

var tapInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show tap details and resources",
	Long: `Show detailed information about a tap including its resources.

Examples:
  aictl tap info mycompany`,
	Args: cobra.ExactArgs(1),
	RunE: runTapInfo,
}

func init() {
	tapCmd.AddCommand(tapAddCmd)
	tapCmd.AddCommand(tapRemoveCmd)
	tapCmd.AddCommand(tapListCmd)
	tapCmd.AddCommand(tapUpdateCmd)
	tapCmd.AddCommand(tapInfoCmd)
	rootCmd.AddCommand(tapCmd)
}

func getTapManager() (*tap.Manager, error) {
	cfg := config.DefaultConfig()
	return tap.NewManager(cfg)
}

func runTapAdd(cmd *cobra.Command, args []string) error {
	name, url := args[0], args[1]

	mgr, err := getTapManager()
	if err != nil {
		return err
	}

	fmt.Printf("Adding tap '%s' from %s...\n", name, url)

	t, err := mgr.Add(name, url)
	if err != nil {
		return err
	}

	color.Green("✓ Added tap '%s'", t.Name)

	// List resources in the tap
	resources, err := mgr.ListResources(name)
	if err == nil && len(resources) > 0 {
		fmt.Printf("\nAvailable resources (%d):\n", len(resources))
		for _, res := range resources {
			fmt.Printf("  - %s/%s (%s)\n", name, res.Name, res.Type)
		}
	}

	return nil
}

func runTapRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := getTapManager()
	if err != nil {
		return err
	}

	if err := mgr.Remove(name); err != nil {
		return err
	}

	color.Green("✓ Removed tap '%s'", name)
	return nil
}

func runTapList(cmd *cobra.Command, args []string) error {
	mgr, err := getTapManager()
	if err != nil {
		return err
	}

	taps, err := mgr.List()
	if err != nil {
		return err
	}

	if len(taps) == 0 {
		fmt.Println("No taps installed.")
		fmt.Println("\nAdd a tap with: aictl tap add <name> <url>")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tURL\tUPDATED")
	_, _ = fmt.Fprintln(w, "----\t---\t-------")

	for _, t := range taps {
		updated := "unknown"
		if !t.UpdatedAt.IsZero() {
			updated = t.UpdatedAt.Format("2006-01-02 15:04")
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", t.Name, t.URL, updated)
	}

	_ = w.Flush()
	return nil
}

func runTapUpdate(cmd *cobra.Command, args []string) error {
	mgr, err := getTapManager()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		// Update all taps
		fmt.Println("Updating all taps...")
		if err := mgr.UpdateAll(); err != nil {
			return err
		}
		color.Green("✓ All taps updated")
	} else {
		// Update specific tap
		name := args[0]
		fmt.Printf("Updating tap '%s'...\n", name)
		if err := mgr.Update(name); err != nil {
			return err
		}
		color.Green("✓ Updated tap '%s'", name)
	}

	return nil
}

func runTapInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := getTapManager()
	if err != nil {
		return err
	}

	t, err := mgr.Get(name)
	if err != nil {
		return err
	}

	fmt.Printf("Name:    %s\n", t.Name)
	fmt.Printf("URL:     %s\n", t.URL)
	fmt.Printf("Path:    %s\n", t.Path)
	if !t.UpdatedAt.IsZero() {
		fmt.Printf("Updated: %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	// List resources
	resources, err := mgr.ListResources(name)
	if err != nil {
		return err
	}

	fmt.Printf("\nResources (%d):\n", len(resources))

	if len(resources) == 0 {
		fmt.Println("  No resources found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "  NAME\tTYPE\tVERSION\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "  ----\t----\t-------\t-----------")

	for _, res := range resources {
		desc := res.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		_, _ = fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", res.Name, res.Type, res.Version, desc)
	}

	_ = w.Flush()
	return nil
}
