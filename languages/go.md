# Go Guidelines

When looking for functions, use the `gofuncs` tool to list all functions in a Go project. It provides a compact format that is optimized for LLM context.

- Example usage:

  ```bash
  go run gofuncs.go -dir /path/to/project
  ```

## Go Libraries

- Provide justification when adding new third-party dependencies. Keep dependencies updated.
- Prefer using standard library packages when possible
- If SQLite is used, use "modernc.org/sqlite" as the library (no dependency on cgo)
- Logging in applications run from cron: "log/slog" (standard library)
  - Use "github.com/lepinkainen/humanlog" to format slog output
- Logging in applications run from CLI: "fmt.Println" (standard library, use emojis for better UX)
- Configuration management: "github.com/spf13/viper"
- Command-line arguments: "github.com/alecthomas/kong" (only if the project requires complex CLI)
- If the application needs an interactive CLI, use "github.com/charmbracelet/bubbletea" for TUI applications

## General Guidelines for Go

- Always run "gofmt -w -s ." on Go code files after making changes
- Prefer `task build` over `go build` to ensure all tasks are run
  - `task build` will run tests, linters, and formatters
- Functions that are easily unit-testable, should have tests
- Don't go for 100% test coverage, test the critical parts of the code
