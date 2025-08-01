# Go Guidelines

IMPORTANT: Before claiming any task as completed, run `task build` to ensure all tasks are run successfully, including tests, linters, and formatters.

## Go versioning

Always use the latest stable Go version for development. Currentlyu, the latest stable version is Go 1.24.

You can check the latest version at [Go Releases](https://go.dev/dl/).

You can check the current installed Go version with:

```bash
go version
```

## Investigating project structure

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

  - Example Kong usage:

    ```go
    type CLI struct {
        Config  string `help:"Path to config file" default:"config.yaml"`
        Verbose bool   `help:"Enable verbose logging"`
    }
    var cli CLI
    kong.Parse(&cli)
    ```

- If the application needs an interactive CLI, use "github.com/charmbracelet/bubbletea" for TUI applications

## Code Formatting and Linting

### Prerequisites

Install required tools:

```bash
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Formatting

- Always run `goimports -w .` on Go code files after making changes
- IMPORTANT: DD NOT use `gofmt` - `goimports` is superior and standard

### Linting with golangci-lint

- Use golangci-lint for comprehensive code quality checks
- Configure with `.golangci.yml` in project root
- Modern govet shadow detection: use `enable: [shadow]` not deprecated `check-shadowing: true`

Example `.golangci.yml`:

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - misspell
    - gocritic
    - revive

linters-settings:
  govet:
    enable:
      - shadow
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style
    disabled-checks:
      - exitAfterDefer # CLI apps often exit after defer
      - rangeValCopy # Sometimes acceptable for readability

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
    - path: llm-shared/
      linters:
        - gocritic
        - revive
```

## Project Setup Requirements

- Use `goimports` instead of `gofmt` for code formatting (includes import management)
- Install prerequisites: `go install golang.org/x/tools/cmd/goimports@latest`
- Configure golangci-lint with `.golangci.yml` for comprehensive linting
- GitHub Actions CI should install goimports alongside other Go tools
- Use modern govet shadow detection: `enable: [shadow]` not `check-shadowing: true`

## General Guidelines for Go

- Always run `goimports -w .` on Go code files after making changes
- Prefer `task build` over `go build` to ensure all tasks are run
  - `task build` will run tests, linters, and formatters
- Functions that are easily unit-testable, should have tests
- Don't go for 100% test coverage, test the critical parts of the code
