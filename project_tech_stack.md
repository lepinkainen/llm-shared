# Project Tech Stack and Guidelines

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

## Go

When looking for functions, use the `gofuncs` tool to list all functions in a Go project. It provides a compact format that is optimized for LLM context.

- Example usage:

  ```bash
  go run gofuncs.go -dir /path/to/project
  ```

### Go Libraries

- Provide justification when adding new third-party dependencies. Keep dependencies updated.
- Prefer using standard library packages when possible
- If SQLite is used, use "modernc.org/sqlite" as the library (no dependency on cgo)
- Logging in applications run from cron: "log/slog" (standard library)
- Logging in applications run from CLI: "fmt.Println" (standard library, use emojis for better UX)
- Configuration management: "github.com/spf13/viper"
- Command-line arguments: "github.com/spf13/cobra" (only if the project requires complex CLI)

### General Guidelines

- Always run "gofmt -w ." on the Go code files after making changes
- Always build the project using the taskfile before finishing a task

## Python

When looking for functions, use the `pyfuncs` tool to list all functions in a Python project. It provides a compact format that is optimized for LLM context.

- Example usage:

  ```bash
  python pyfuncs.py --dir /path/to/project
  ```

### Python Libraries

- Use "uv" for dependency management (similar to pipenv)
- Use "ruff" for linting and formatting
- Use "mypy" for type checking as much as possible
