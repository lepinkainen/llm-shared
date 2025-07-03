# LLM Shared

Shared resources for LLMs to use the same libraries and tools in projects. This repository provides standardized guidelines, tools, and utilities to help AI assistants work more effectively with codebases.

## What's Included

- **Tech Stack Guidelines** (`project_tech_stack.md`) - Language-specific best practices and recommended libraries
- **Function Analyzers** (`utils/`) - Tools to extract function signatures in LLM-optimized formats
  - `gofuncs.go` - Go function analyzer
  - `pyfuncs.py` - Python function analyzer
- **Development Patterns** - Common project structures and workflows

## Quick Start

### As a Git Submodule

```bash
git submodule add https://github.com/lepinkainen/llm-shared.git llm-shared
```

### In Your CLAUDE.md

Add this to your project's `CLAUDE.md` file:

```markdown
# Project Guidelines

Refer to llm-shared/ directory for:

- Tech stack guidelines (llm-shared/project_tech_stack.md)
- Function analysis tools (llm-shared/utils/)
- Development best practices

# Function Analysis

- For Go projects: `go run llm-shared/utils/gofuncs.go -dir .`
- For Python projects: `python llm-shared/utils/pyfuncs.py --dir .`
```

## Usage Examples

### Analyzing Functions

```bash
# Go projects
go run llm-shared/utils/gofuncs.go -dir ./src

# Python projects
python llm-shared/utils/pyfuncs.py --dir ./src
```

### Following Guidelines

The `project_tech_stack.md` provides language-specific recommendations:

- Go: Standard library preferences, recommended third-party libraries
- Python: Modern tooling (uv, ruff, mypy)
- General: Build systems, testing patterns, project structure

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:

- Adding new language support
- Improving existing tools
- Submitting improvements

## License

MIT License - see [LICENSE](LICENSE) for details.
