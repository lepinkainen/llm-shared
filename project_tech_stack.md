# Project Tech Stack and Guidelines

## Git use

My Github repository root is at <https://github.com/lepinkainen/>

- IMPORTANT: NEVER commit to the "main" or "master" branch directly
- Use feature branches for new features or bug fixes
- Keep commits small and focused
- Write clear, descriptive commit messages
- Rebase branches before merging to keep history clean
- Use pull requests for code reviews and discussions

## Project management

- When working from a markdown checklist of tasks, check off the tasks as you complete them
- Task is not complete until `task build` succeeds, which includes:
  - Running tests
  - Linting the code
  - Building the project (if applicable)
- Task is not complete until it has even basic unit tests, even if they are not comprehensive
  - No need to mock external dependencies, just test the logic of the code

## Project analysis with Gemini CLI

- When analysing a large codebase that might exceed context limits, use the Gemini CLI
- Use gemini -p when:
  - Analysing entire codebases or large directories
  - Comparing multiple large files
  - Need to understand project-wide patterns or architecture
  - Checking for presence of certain coding patterns or practices

Examples:

```bash
gemini -p "@src/main.go Explain this file's purpose and functionality"
gemini -p "@src/ Summarise the architecture of this codebase"
gemini -p "@src/ Is the project test coverage on par with industry standards?"
```

## Project validation

- Use the `validate-docs` tool to check if projects follow standard structure conventions
- The tool auto-detects Go and Python projects and validates:
  - Standard directory structure (cmd/internal/pkg for Go, src for Python)
  - Required files (go.mod, requirements.txt, etc.)
  - Build system configuration (Taskfile.yml)
  - Code patterns (main functions, test functions) using gofuncs/pyfuncs tools

```bash
# Validate current directory
go run utils/validate-docs/validate-docs.go

# Validate specific project
go run utils/validate-docs/validate-docs.go --dir /path/to/project
```

## Common project guidelines

- taskfile "taskfile.dev/task" (for task management instead of makefiles)
  - should containt the following tasks:
    - build
    - build-linux
    - build-ci (for building in CI)
    - test
    - test-ci (for build-ci tests)
      - go test -tags=ci -cover -v ./...
      - allow skipping tests with //go:build !ci
    - lint
    - build tasks need to depend on test and lint tasks
  - All build artefacts should be placed in the `build/` directory if the language builds to a binary
  - Projects should have a basic Github Actions setup that uses the build-ci task to run tests and linting on push and pull requests
  - See `templates/github/workflows/` for CI templates for Go, Python, and JavaScript projects
  - Always keep `.gitignore` up to date with the language-specific ignores so that build artefacts and temporary files are not committed
  - See `templates/gitignore-*` files for language-specific .gitignore templates
  - When doing HTTP requests, use a custom user agent that includes the project name and version, e.g. `MyProject/1.0.0`
  - See `templates/Taskfile.yml` for a comprehensive example template that follows these guidelines

## Language-Specific Guidelines

For detailed language-specific guidelines, libraries, and best practices, see:

- **[Go Guidelines](languages/go.md)** - Go libraries, tools, and conventions
- **[Python Guidelines](languages/python.md)** - Python libraries, tools, and conventions
- **[JavaScript/TypeScript Guidelines](languages/javascript.md)** - JS/TS libraries, frameworks, and tools
