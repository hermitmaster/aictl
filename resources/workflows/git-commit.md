---
name: git-commit
version: 1.0.0
type: workflow
description: Generate conventional commit messages from staged changes
author: hermitmaster
tags:
  - git
  - commit
  - conventional-commits
tools:
  - windsurf
  - cursor
  - aider
---

# Git Commit Workflow

This workflow helps generate conventional commit messages based on staged changes.

## Steps

1. Review the staged changes:
```bash
git diff --staged
```

2. Based on the changes, generate a commit message following the Conventional Commits specification:
   - `feat:` - A new feature
   - `fix:` - A bug fix
   - `docs:` - Documentation only changes
   - `style:` - Changes that do not affect the meaning of the code
   - `refactor:` - A code change that neither fixes a bug nor adds a feature
   - `perf:` - A code change that improves performance
   - `test:` - Adding missing tests or correcting existing tests
   - `chore:` - Changes to the build process or auxiliary tools

3. Format the commit message:
```
<type>(<optional scope>): <description>

<optional body>

<optional footer>
```

4. Commit with the generated message:
```bash
git commit -m "<generated message>"
```

## Examples

- `feat(auth): add OAuth2 support`
- `fix: resolve null pointer in user service`
- `docs: update README with installation instructions`
- `refactor(api): simplify error handling logic`
