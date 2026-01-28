package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new resource registry (tap)",
	Long: `Initialize a new resource registry (tap) directory structure.

Creates the standard directory structure for a custom tap:
  <path>/
  ├── rules/          # Coding standards and guidelines
  ├── workflows/      # Multi-step automation workflows
  ├── skills/         # Reusable skill definitions
  ├── bin/            # Executable scripts (multi-file resources)
  └── README.md       # Registry documentation

If no path is specified, creates the structure in the current directory.

Examples:
  aictl init                      # Initialize in current directory
  aictl init my-tap               # Create my-tap/ directory
  aictl init ~/repos/company-tap  # Create at specific path`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Expand ~ if present
	if targetPath[:1] == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error expanding home directory: %w", err)
		}
		targetPath = filepath.Join(home, targetPath[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("error resolving path: %w", err)
	}

	fmt.Printf("Initializing tap at: %s\n\n", absPath)

	// Create directories
	dirs := []string{
		"rules",
		"workflows",
		"skills",
		"bin",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(absPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}
		fmt.Printf("  Created %s/\n", dir)
	}

	// Create README.md
	readmePath := filepath.Join(absPath, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		tapName := filepath.Base(absPath)
		if tapName == "." {
			tapName = "my-tap"
		}
		readmeContent := generateReadme(tapName)
		if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
			return fmt.Errorf("error creating README.md: %w", err)
		}
		fmt.Println("  Created README.md")
	}

	// Create sample rule
	sampleRulePath := filepath.Join(absPath, "rules", "example-rules.md")
	if _, err := os.Stat(sampleRulePath); os.IsNotExist(err) {
		if err := os.WriteFile(sampleRulePath, []byte(sampleRule), 0644); err != nil {
			return fmt.Errorf("error creating sample rule: %w", err)
		}
		fmt.Println("  Created rules/example-rules.md (sample)")
	}

	// Create sample workflow
	sampleWorkflowPath := filepath.Join(absPath, "workflows", "example-workflow.md")
	if _, err := os.Stat(sampleWorkflowPath); os.IsNotExist(err) {
		if err := os.WriteFile(sampleWorkflowPath, []byte(sampleWorkflow), 0644); err != nil {
			return fmt.Errorf("error creating sample workflow: %w", err)
		}
		fmt.Println("  Created workflows/example-workflow.md (sample)")
	}

	// Create .gitignore
	gitignorePath := filepath.Join(absPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignoreContent := "# OS files\n.DS_Store\nThumbs.db\n\n# Editor files\n*.swp\n*.swo\n*~\n.idea/\n.vscode/\n"
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			return fmt.Errorf("error creating .gitignore: %w", err)
		}
		fmt.Println("  Created .gitignore")
	}

	fmt.Println()
	color.Green("✓ Tap initialized successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. cd %s\n", absPath)
	fmt.Println("  2. Edit the sample resources or create your own")
	fmt.Println("  3. git init && git add . && git commit -m 'Initial tap'")
	fmt.Println("  4. Push to a Git repository")
	fmt.Println()
	fmt.Println("Others can then add your tap with:")
	fmt.Printf("  aictl tap add <name> <git-url>\n")

	return nil
}

func generateReadme(tapName string) string {
	return fmt.Sprintf(`# %s

A custom resource registry for aictl.

## Installation

`+"```bash"+`
aictl tap add %s <git-url>
`+"```"+`

## Available Resources

### Rules

| Name | Description |
|------|-------------|
| example-rules | Example coding rules (sample) |

### Workflows

| Name | Description |
|------|-------------|
| example-workflow | Example workflow (sample) |

## Usage

`+"```bash"+`
# Install a resource from this tap
aictl install %s/example-rules --tool=windsurf

# List available resources
aictl tap info %s
`+"```"+`

## Contributing

1. Create a new resource file in the appropriate directory
2. Add YAML frontmatter with required metadata
3. Submit a pull request

## Resource Format

Resources use YAML frontmatter for metadata:

`+"```markdown"+`
---
name: my-resource
version: 1.0.0
type: rules  # or workflow, skill, bin
description: A brief description
author: your-name
tags:
  - tag1
  - tag2
tools:
  - windsurf
  - cursor
---

# Resource Content

Your content here...
`+"```"+`
`, tapName, tapName, tapName, tapName)
}

const sampleRule = `---
name: example-rules
version: 1.0.0
type: rules
description: Example coding rules - customize or replace this file
author: your-name
tags:
  - example
  - template
tools:
  - windsurf
  - cursor
  - aider
  - continue
---

# Example Coding Rules

This is a sample rules file. Replace this content with your own coding standards.

## Code Style

- Use consistent indentation
- Write descriptive variable names
- Keep functions small and focused

## Documentation

- Document public APIs
- Include usage examples
- Keep documentation up to date

## Testing

- Write tests for new features
- Maintain test coverage
- Use descriptive test names
`

const sampleWorkflow = `---
name: example-workflow
version: 1.0.0
type: workflow
description: Example workflow - customize or replace this file
author: your-name
tags:
  - example
  - template
tools:
  - windsurf
---

# Example Workflow

This is a sample workflow file. Replace this content with your own workflow.

## Steps

1. First, gather context about the task
2. Then, implement the solution
3. Finally, verify the results

## Usage

Invoke this workflow with: /example-workflow
`
