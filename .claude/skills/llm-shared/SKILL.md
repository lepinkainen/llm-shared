---
name: llm-shared
description: Development guidelines and code analysis tools for Go, Python, and JavaScript/TypeScript projects. Use when setting up new projects, analyzing code structure, validating project configuration, writing code, or applying language-specific best practices. Activates for any development task.
---

# LLM Shared Development Guidelines

This skill provides standardized development practices, code analysis tools, and project templates for software development.

## Quick Reference

### Code Analysis Tools

Run these from the project root (adjust path to llm-shared location):

```bash
# List all Go functions (compact LLM-friendly format)
go run llm-shared/utils/gofuncs/gofuncs.go -dir .

# List all Python functions
python3 llm-shared/utils/pyfuncs.py --dir .

# List all JavaScript/TypeScript functions
node llm-shared/utils/jsfuncs.js --dir .

# Validate project structure
go run llm-shared/utils/validate-docs/validate-docs.go --auto-detect

# Analyze Go file complexity
go run llm-shared/utils/go-file-analyzer/main.go -dir . -n 20 -topfuncs 5

# Analyze Python file complexity
python3 llm-shared/utils/py-file-analyzer/main.py -dir . -n 20 -topfuncs 5
```

### Output Formats

**gofuncs**: `file:line:type:exported:name:receiver:signature`
- type: `f`=function, `m`=method
- exported: `y`=public, `n`=private

**pyfuncs**: `file:line:type:exported:name:class:signature:decorators`
- type: `f`=function, `m`=method, `s`=staticmethod, `c`=classmethod, `p`=property

**jsfuncs**: `file:line:type:exported:name:class:signature:decorators`
- type: `f`=function, `m`=method, `a`=arrow, `c`=constructor, `g`=getter, `s`=setter

## Core Principles

1. **Never commit to main/master directly** - Use feature branches
2. **Task is not complete until `task build` succeeds** (tests, lint, build)
3. **Task is not complete without basic unit tests**
4. **Use taskfile.dev** instead of Makefiles
5. **Keep .gitignore up to date** with language-specific ignores

## Language Guidelines

### Go

For detailed Go guidelines, see [languages/go.md](../../../languages/go.md)

Key points:
- Use `goimports -w .` (not gofmt) after changes
- Use `golangci-lint` with `.golangci.yml` config
- Prefer stdlib; if SQLite needed use `modernc.org/sqlite` (no CGO)
- Config: `github.com/spf13/viper`
- CLI args: `github.com/alecthomas/kong`
- Logging: `log/slog` for cron jobs, `fmt.Println` for CLI
- Always use `errors.Is` and `errors.As` for error handling
- Use `any` instead of `interface{}`

### Python

For detailed Python guidelines, see [languages/python.md](../../../languages/python.md)

Key points:
- Use `uv` for dependency management
- Use `ruff` for linting and formatting
- Use `mypy` for type checking

### TypeScript/JavaScript

For detailed JS/TS guidelines, see [languages/javascript.md](../../../languages/javascript.md)

Key points:
- **Always use TypeScript** for new projects
- **Prefer Deno** over Node.js for new projects
- Use `pnpm` for Node.js package management
- Use `zod` for validation
- Use ES modules and async/await

## Taskfile Structure

Every project should have a `Taskfile.yml` with these tasks:
- `build` - Build the project (depends on test and lint)
- `build-linux` - Cross-compile for Linux
- `build-ci` - CI build (compile only, no test/lint dependencies)
- `test` - Run tests
- `test-ci` - Run CI tests (allows skipping with build tags)
- `lint` - Run linters

Build artifacts go in `build/` directory.

See [templates/Taskfile.yml](../../../templates/Taskfile.yml) for a complete example.

## Docker Deployment

For containerized applications:
- Use multi-stage builds for minimal image sizes
- Prefer pure Go (`modernc.org/sqlite`) over CGO when possible
- Include `/health` endpoint for health checks
- Run as non-root user
- Host on GitHub Container Registry (GHCR)

Templates available in `templates/docker/`:
- `Dockerfile-go-pure` - Pure Go apps
- `Dockerfile-go-cgo` - CGO-dependent apps
- `docker-compose.yml` - Local development
- `github-workflows-docker.yml` - CI/CD for GHCR

## HTTP Services

All HTTP services must implement `/whoami` endpoint returning:
```json
{
  "name": "service-name",
  "version": "1.0.0",
  "hash": "abc123",
  "build_time": "2024-01-01T00:00:00Z"
}
```

## Current Toolchain Versions

For canonical versions, see [versions.md](../../../versions.md)

Key versions (check versions.md for latest):
- Go: 1.25.1
- Python: 3.13.7
- Node.js: 24.8.0

## GitHub Actions

CI templates in `templates/github/workflows/`:
- `go-ci.yml` - Basic Go testing/linting
- `go-ci-advanced.yml` - Multi-version Go CI
- `python-ci.yml` - Python workflow
- `js-ci.yml` - JavaScript/TypeScript workflow

Use Dependabot for automated action updates - see template in `templates/github/workflows/dependabot.yml`.

## Shell Commands

Prefer modern tools:
- Use `rg` (ripgrep) instead of `grep`
- Use `fd` instead of `find`

See [shell_commands.md](../../../shell_commands.md) for usage patterns.

## GitHub Issues

Use the `gh` CLI for issue management. Standard workflow labels:
- `needs-plan` - Needs implementation plan
- `ready` - Ready for work
- `in-progress` - Being worked on
- `review` - Needs review

See [GITHUB.md](../../../GITHUB.md) for detailed workflow.
