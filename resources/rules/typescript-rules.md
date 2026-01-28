---
name: typescript-rules
version: 1.0.0
type: rules
description: TypeScript coding standards and best practices
author: drausch
tags:
  - typescript
  - standards
  - coding
tools:
  - windsurf
  - cursor
  - aider
  - continue
---

# TypeScript Coding Standards

## General Principles

- Use TypeScript strict mode
- Prefer `const` over `let`, avoid `var`
- Use explicit type annotations for function parameters and return types
- Avoid `any` type - use `unknown` when type is truly unknown

## Naming Conventions

- **Variables/Functions**: camelCase (`getUserName`, `isValid`)
- **Classes/Interfaces/Types**: PascalCase (`UserService`, `ApiResponse`)
- **Constants**: SCREAMING_SNAKE_CASE (`MAX_RETRIES`, `API_BASE_URL`)
- **Files**: kebab-case (`user-service.ts`, `api-client.ts`)

## Functions

- Keep functions small and focused (single responsibility)
- Use arrow functions for callbacks and short functions
- Use named functions for complex logic
- Document public functions with JSDoc

## Error Handling

- Use typed errors when possible
- Always handle promise rejections
- Use try/catch for async/await
- Log errors with context

## Imports

- Group imports: external, internal, relative
- Use named imports over default imports
- Avoid circular dependencies
