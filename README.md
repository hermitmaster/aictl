# aictl

A registry-based CLI tool for installing AI coding assistant resources (rules, workflows, skills) across multiple tools like Windsurf, Cursor, Aider, and Continue.

Inspired by Homebrew, `aictl` supports bundled resources and custom taps (Git repositories) for sharing resources across teams.

## Features

- **Multi-tool support**: Install resources to Windsurf, Cursor, Aider, and Continue
- **Bundled resources**: Pre-packaged resources ready to install
- **Custom taps**: Add Git repositories as custom resource registries
- **Declarative config**: Use `.aiconfig` for reproducible setups
- **Global & local scopes**: Install system-wide or per-project

## Installation

### From Source

```bash
cd aictl
make install
```

This installs `aictl` to `/usr/local/bin` and generates shell completions.

### Shell Completions

After installation, add to your shell config:

```bash
# Bash (~/.bashrc)
source ~/.local/share/aictl/completions/aictl.bash

# Zsh (~/.zshrc)
source ~/.local/share/aictl/completions/aictl.zsh

# Fish (~/.config/fish/config.fish)
source ~/.local/share/aictl/completions/aictl.fish
```

## Quick Start

```bash
# Search for available resources
aictl search

# Install a bundled resource
aictl install bundled/jira-context --tool=windsurf

# Install to multiple tools
aictl install bundled/typescript-rules --tool=windsurf,cursor

# List installed resources
aictl list

# Show supported tools
aictl tools
```

## Commands

### `aictl install <source>/<name>`

Install a resource from bundled resources or a custom tap.

```bash
aictl install bundled/jira-context              # Install to detected tools
aictl install bundled/typescript-rules -t cursor # Install to specific tool
aictl install mycompany/internal-rules          # Install from tap
aictl install bundled/code-review --local       # Install to project directory
aictl install bundled/rules --force             # Overwrite existing files
```

### `aictl uninstall <name>`

Remove an installed resource.

```bash
aictl uninstall jira-context
aictl uninstall typescript-rules --tool=cursor
```

### `aictl list`

List installed resources.

```bash
aictl list                  # Show all installed
aictl list --global         # Show global only
aictl list --local          # Show local only
aictl list --tool=windsurf  # Filter by tool
```

### `aictl search [query]`

Search for available resources.

```bash
aictl search                    # List all resources
aictl search typescript         # Search by name/description
aictl search --type=rules       # Filter by type
aictl search --source=bundled   # Filter by source
```

### `aictl tools`

Show supported tools and their status.

```bash
aictl tools
```

Output:
```
TOOL      STATUS     GLOBAL CONFIG                     LOCAL CONFIG  SUPPORTS
windsurf  detected   ~/.codeium/windsurf               .windsurf     rules, workflows, skills, bin
cursor    not found  ~/.cursor                         .cursor       rules
aider     not found  ~/.aider                          .aider        rules
continue  not found  ~/.continue                       .continue     rules
```

### `aictl bundle`

Install resources from a `.aiconfig` file.

```bash
aictl bundle                     # Use .aiconfig in current directory
aictl bundle path/to/.aiconfig   # Use specific file
aictl bundle --dry-run           # Preview changes
aictl bundle --force             # Overwrite existing
```

### `aictl bundle dump`

Export current installed state to `.aiconfig` format.

```bash
aictl bundle dump > .aiconfig
```

### `aictl tap`

Manage custom resource registries (taps).

```bash
aictl tap add mycompany https://github.com/mycompany/ai-resources
aictl tap list
aictl tap update              # Update all taps
aictl tap update mycompany    # Update specific tap
aictl tap info mycompany      # Show tap details
aictl tap remove mycompany
```

## Configuration

### `.aiconfig`

Declarative manifest for reproducible resource installations:

```yaml
# Tools installed on this system
tools:
  - windsurf
  - cursor

# Custom taps (optional)
taps:
  - name: mycompany
    url: https://github.com/mycompany/ai-resources

# Resources to install
install:
  # Install to all tools
  - name: bundled/typescript-rules

  # Install to specific tools only
  - name: bundled/jira-context
    tools: [windsurf]

  # Install from a tap
  - name: mycompany/internal-standards

# Default scope (local or global)
scope: local
```

### Resource Types

| Type | Description | Supported Tools |
|------|-------------|-----------------|
| `rules` | Coding standards and guidelines | All |
| `workflow` | Multi-step automation workflows | Windsurf |
| `skill` | Reusable skill definitions | Windsurf |
| `bin` | Executable scripts | Windsurf |

## Creating Resources

Resources use YAML frontmatter for metadata:

```markdown
---
name: my-rules
version: 1.0.0
type: rules
description: My custom coding rules
author: yourname
tags:
  - coding
  - standards
tools:
  - windsurf
  - cursor
---

# My Rules

Your content here...
```

### Creating a Tap

A tap is a Git repository with the following structure:

```
my-tap/
├── rules/
│   ├── rule1.md
│   └── rule2.md
├── workflows/
│   └── workflow1.md
├── skills/
│   └── skill1.md
└── bin/
    └── script1/
        ├── resource.yaml
        ├── script.sh
        └── helper.py
```

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--tool` | `-t` | Target tool(s), comma-separated |
| `--global` | `-g` | Install to global config directory |
| `--local` | `-l` | Install to project-local directory |
| `--force` | `-f` | Overwrite existing files |
| `--verbose` | `-v` | Enable verbose output |

## Development

```bash
# Build
make build

# Run tests
make test

# Run with coverage
make coverage

# Build for all platforms
make build-all

# Create release archives
make release
```

## License

MIT
