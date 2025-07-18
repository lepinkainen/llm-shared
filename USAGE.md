# Usage Guide

This guide provides detailed instructions on how to integrate and use llm-shared in your projects.

## Installation Methods

### Method 1: Git Submodule (Recommended)

```bash
# Add as submodule to your project
git submodule add https://github.com/lepinkainen/llm-shared.git llm-shared

# Initialize and update submodule
git submodule update --init --recursive
```

### Method 2: Direct Clone

```bash
# Clone into your project directory
git clone https://github.com/lepinkainen/llm-shared.git llm-shared
```

## Project Integration

### CLAUDE.md Configuration

Create or update your project's `CLAUDE.md` file:

```markdown
# Project Setup

This project uses llm-shared for standardized development practices.

# Guidelines Reference

- Tech stack guidelines: llm-shared/project_tech_stack.md
- Function analysis tools: llm-shared/utils/

# Function Analysis Commands

- Go: `go run llm-shared/utils/gofuncs/gofuncs.go -dir .`
- Python: `python llm-shared/utils/pyfuncs.py --dir .`

# Build Commands

- Build: task build
- Test: task test
- Lint: task lint
```

## Function Analysis Tools

### Go Projects

```bash
# Analyze all Go files in current directory
go run llm-shared/utils/gofuncs/gofuncs.go -dir .

# Analyze specific directory
go run llm-shared/utils/gofuncs/gofuncs.go -dir ./src

# Example output:
# main.go:15:f:y:main:()
# api.go:23:m:y:GetUser:*UserService:(id string)(*User,error)
```

### Python Projects

```bash
# Analyze all Python files in current directory
python llm-shared/utils/pyfuncs.py --dir .

# Analyze specific directory
python llm-shared/utils/pyfuncs.py --dir ./src

# Example output:
# main.py:15:f:y:process_data::(data:List[str])->Dict[str,int]:
# api.py:45:m:y:get_user:UserService:(user_id:str)->Optional[User]:async
```

## Real-World Integration Examples

### Example 1: Go Web API Project

**Project Structure:**

```
my-api/
├── llm-shared/          # Git submodule
├── CLAUDE.md           # LLM instructions
├── Taskfile.yml        # Build tasks
├── cmd/
│   └── server/
│       └── main.go
└── internal/
    ├── api/
    └── models/
```

**CLAUDE.md:**

```markdown
# API Project Guidelines

Follow llm-shared/project_tech_stack.md for Go development.

# Function Analysis

Use: `go run llm-shared/utils/gofuncs/gofuncs.go -dir .`

# Build Commands

- `task build` - Build server binary
- `task test` - Run tests
- `task lint` - Run linter
```

### Example 2: Python Data Processing Project

**Project Structure:**

```
data-processor/
├── llm-shared/          # Git submodule
├── CLAUDE.md           # LLM instructions
├── pyproject.toml      # Python dependencies
├── src/
│   ├── processor/
│   └── utils/
└── tests/
```

**CLAUDE.md:**

```markdown
# Data Processing Project

Follow llm-shared/project_tech_stack.md for Python development.

# Function Analysis

Use: `python llm-shared/utils/pyfuncs.py --dir src/`

# Development Commands

- `uv sync` - Install dependencies
- `ruff check` - Lint code
- `mypy src/` - Type check
- `pytest` - Run tests
```

## Working with Guidelines

### Tech Stack Guidelines

The `project_tech_stack.md` file provides:

1. **Language-specific recommendations**

   - Preferred libraries and frameworks
   - Standard patterns and practices
   - Tooling recommendations

2. **Project structure guidelines**

   - Build system setup (Taskfile)
   - Testing strategies
   - CI/CD patterns

3. **Common workflows**
   - Development process
   - Code review practices
   - Deployment patterns

### Following Recommendations

**For Go projects:**

- Use standard library when possible
- Prefer `modernc.org/sqlite` for SQLite
- Use `log/slog` for cron jobs, `fmt.Println` for CLI
- Always run `gofmt -w .` after code changes

**For Python projects:**

- Use `uv` for dependency management
- Use `ruff` for linting and formatting
- Use `mypy` for type checking
- Follow modern Python practices

## Maintenance

### Updating llm-shared

```bash
# If using as submodule
cd llm-shared
git pull origin main
cd ..
git add llm-shared
git commit -m "Update llm-shared"

# If using direct clone
cd llm-shared
git pull origin main
```

### Customization

You can extend the guidelines by:

1. Adding project-specific rules to your `CLAUDE.md`
2. Creating additional analysis tools in your project
3. Extending the existing tools for your needs

## Troubleshooting

### Common Issues

**Go function analyzer not working:**

```bash
# Ensure Go is installed and GOPATH is set
go version
go run llm-shared/utils/gofuncs.go -dir .
```

**Python function analyzer errors:**

```bash
# Ensure Python 3.7+ is installed
python --version
python llm-shared/utils/pyfuncs.py --dir .
```

**Submodule not updating:**

```bash
git submodule update --remote llm-shared
```

## Advanced Usage

### Custom Analysis Scripts

You can create project-specific wrappers:

```bash
#!/bin/bash
# analyze.sh - Project analysis script
echo "=== Go Functions ==="
go run llm-shared/utils/gofuncs/gofuncs.go -dir ./cmd
go run llm-shared/utils/gofuncs/gofuncs.go -dir ./internal

echo "=== Python Functions ==="
python llm-shared/utils/pyfuncs.py --dir ./scripts
```

### Integration with IDEs

Configure your IDE to run analysis tools:

**VS Code tasks.json:**

```json
{
  "tasks": [
    {
      "label": "Analyze Go Functions",
      "type": "shell",
      "command": "go run llm-shared/utils/gofuncs/gofuncs.go -dir ."
    }
  ]
}
```
