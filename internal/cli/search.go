package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/hermitmaster/aictl/internal/config"
	"github.com/hermitmaster/aictl/internal/tap"
	"github.com/hermitmaster/aictl/resources"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for resources",
	Long: `Search for resources across bundled resources and installed taps.

If no query is provided, lists all available resources.

Examples:
  aictl search                    # List all resources
  aictl search typescript         # Search for 'typescript'
  aictl search --type=rules       # Filter by type
  aictl search --source=bundled   # Filter by source`,
	RunE: runSearch,
}

var (
	searchTypeFlag   string
	searchSourceFlag string
)

func init() {
	searchCmd.Flags().StringVar(&searchTypeFlag, "type", "", "Filter by resource type (rules, workflow, skill, bin)")
	searchCmd.Flags().StringVar(&searchSourceFlag, "source", "", "Filter by source (bundled or tap name)")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 0 {
		query = strings.ToLower(args[0])
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "SOURCE\tNAME\tTYPE\tVERSION\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "------\t----\t----\t-------\t-----------")

	count := 0

	// Search bundled resources
	if searchSourceFlag == "" || searchSourceFlag == "bundled" {
		bundled, err := resources.ListBundled()
		if err == nil {
			for _, res := range bundled {
				if matchesSearch(res.Name, res.Description, query, string(res.Type), searchTypeFlag) {
					desc := truncate(res.Description, 40)
					_, _ = fmt.Fprintf(w, "bundled\t%s\t%s\t%s\t%s\n", res.Name, res.Type, res.Version, desc)
					count++
				}
			}
		}
	}

	// Search taps
	if searchSourceFlag == "" || (searchSourceFlag != "" && searchSourceFlag != "bundled") {
		cfg := config.DefaultConfig()
		tapMgr, err := tap.NewManager(cfg)
		if err == nil {
			taps, err := tapMgr.List()
			if err == nil {
				for _, t := range taps {
					if searchSourceFlag != "" && searchSourceFlag != t.Name {
						continue
					}

					tapResources, err := tapMgr.ListResources(t.Name)
					if err != nil {
						continue
					}

					for _, res := range tapResources {
						if matchesSearch(res.Name, res.Description, query, string(res.Type), searchTypeFlag) {
							desc := truncate(res.Description, 40)
							_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.Name, res.Name, res.Type, res.Version, desc)
							count++
						}
					}
				}
			}
		}
	}

	_ = w.Flush()

	if count == 0 {
		if query != "" {
			fmt.Printf("\nNo resources found matching '%s'\n", query)
		} else {
			fmt.Println("\nNo resources found")
		}
	} else {
		fmt.Printf("\n%d resource(s) found\n", count)
	}

	return nil
}

func matchesSearch(name, description, query, resType, typeFilter string) bool {
	// Type filter
	if typeFilter != "" && resType != typeFilter {
		return false
	}

	// Query filter
	if query == "" {
		return true
	}

	nameLower := strings.ToLower(name)
	descLower := strings.ToLower(description)

	return strings.Contains(nameLower, query) || strings.Contains(descLower, query)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
