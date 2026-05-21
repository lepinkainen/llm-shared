# Go Guidelines

IMPORTANT: Before claiming any task as completed, run `task build` to ensure all tasks are run successfully, including tests, linters, and formatters.

Be aware of file sizes, if a single file grows too large, suggest splitting it into smaller files.

## Go versioning

Always use the Go toolchain listed in `versions.md`, which is kept in sync with the latest stable release.

You can check the locally installed Go version with:

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
# Formatting + primary linter
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Advisory tools (run separately, not gated on lint)
go install golang.org/x/vuln/cmd/govulncheck@latest      # CVE scanner
go install golang.org/x/tools/cmd/deadcode@latest        # unreachable funcs
go install github.com/uudashr/gocognit/cmd/gocognit@latest # cognitive complexity
# `jq` is also required to render gocognit JSON output (brew install jq / apt install jq)
```

### Formatting

- Always run `goimports -w .` on Go code files after making changes
- IMPORTANT: DD NOT use `gofmt` - `goimports` is superior and standard

### Linting with golangci-lint

- Use golangci-lint for comprehensive code quality checks
- Configure with `.golangci.yml` in project root
- Modern govet shadow detection: use `enable: [shadow]` not deprecated `check-shadowing: true`

A ready-to-copy v2 config lives at `templates/golangci.yml`. Reproduced here for reference:

```yaml
version: "2"

run:
  timeout: 5m
  issues-exit-code: 1
  tests: false # No point in linting tests

linters:
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - ineffassign
    - misspell
    - gocritic
    - revive
    - errorlint     # forces errors.Is / errors.As over == comparison
    - bodyclose     # catches missing resp.Body.Close()
    - contextcheck  # context.Context propagation
    - sqlclosecheck # *sql.Rows / *sql.Stmt close hygiene
    - noctx         # HTTP requests without a context
    - gosec         # common security smells
    - unparam       # always-nil error returns, unused params
    - gocyclo       # cyclomatic complexity guard
  exclusions:
    paths:
      - llm-shared
  settings:
    govet:
      enable:
        - shadow
    gocyclo:
      min-complexity: 15 # raise if existing codebase has many large funcs
    gocritic:
      enabled-tags:
        - diagnostic
        - experimental
        - opinionated
        - performance
        - style
      disabled-checks:
        - exitAfterDefer
        - rangeValCopy
        - hugeParam
        - singleCaseSwitch
        - ifElseChain
        - paramTypeCombine
    revive:
      severity: warning
      rules:
        - name: exported
          severity: warning
          disabled: false
          arguments:
            - "checkPrivateReceivers"
            - "sayRepetitiveInsteadOfStutters"

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
```

#### Linter notes

- **gocyclo**: starts at `min-complexity: 15`. If an existing codebase has many over-threshold funcs, raise the bound temporarily and ratchet down. Refactor by extracting helpers — closures-in-closures and chained `if err != nil` blocks are the usual culprits.
- **unparam**: flags always-nil error returns and unused parameters. Fix by changing the signature; do not silence.
- **gosec**: noisy on `cmd/` entrypoints and test utilities. Path-scope exclusions are cleaner than `// #nosec` annotations.
- **noctx**: legitimately violated by DB-driver init code. Use a path-scoped exclusion for those files rather than disabling project-wide.

### Advisory tooling (not gated on lint)

These run as separate `task` targets and are NOT part of `task lint`, so a regression won't block the build. Wire them up in CI or run on demand.

- **govulncheck** — `task vuln-go` — scans the import graph against the Go vulnerability DB. Reports only vulns that your code actually reaches.
- **deadcode** — `task deadcode-go` — reports unreachable functions. Note: interface-satisfying methods that are never invoked (only used to prove compile-time interface compliance via `var _ Iface = (*T)(nil)`) appear as false positives — these are expected.
- **gocognit** — `task cognit-go` — reports functions with cognitive complexity > 20 plus per-branch diagnostics showing each contributing branch and its nesting depth. Cognit catches what gocyclo misses: nested closures, multi-level loops, and deep conditional pyramids.

Cognit vs. cyclo, in short: cyclomatic counts branches; cognitive multiplies by nesting depth. A function with 30 sequential `if` blocks scores 31 on cyclo but ~30 on cognit; a function with 5 triple-nested `if`s scores ~15 on cyclo but ~25 on cognit. Use both.

## Project Setup Requirements

- Use `goimports` instead of `gofmt` for code formatting (includes import management)
- Install prerequisites: `go install golang.org/x/tools/cmd/goimports@latest`
- Configure golangci-lint with `.golangci.yml` for comprehensive linting
- GitHub Actions CI should install goimports alongside other Go tools
- Use modern govet shadow detection: `enable: [shadow]` not `check-shadowing: true`

## GitHub Actions CI Best Practices

### Template Usage

Use the provided `templates/github/workflows/go-ci.yml` as a starting point. The template includes:

- **Parallel jobs**: Separate test, lint, and build jobs for faster feedback
- **Performance**: Built-in Go module caching via `actions/setup-go@v5`
- **Modern tooling**: golangci-lint-action for efficient linting

**IMPORTANT**: The `build-ci` task should ONLY compile/build the application. Do NOT add dependencies on `test` or `lint` tasks in `build-ci`. The CI workflow manages job dependencies via `needs: [test, lint]` on the build job. This ensures:

- Clean separation of concerns (test job tests, lint job lints, build job builds)
- Faster CI builds (no duplicate test/lint runs)
- No need to install testing/linting tools in the build job

### Security Best Practices

- **Regular updates**: Keep action versions current for security patches
- **Minimal permissions**: Use `permissions: contents: read` for read-only workflows

### Keeping Pinned Actions Updated

Since actions are pinned to commit SHAs for security, you need a strategy to update them:

#### Automated Updates (Recommended)

Use Dependabot to automatically create PRs for action updates. Add `.github/dependabot.yml`:

```yaml
---
version: 2
updates:
  - package-ecosystem: 'github-actions'
    directory: '/'
    schedule:
      interval: 'weekly'
    commit-message:
      prefix: 'ci'
      include: 'scope'
```

#### Manual Monitoring

- Monitor key action repositories for releases:
  - [actions/checkout](https://github.com/actions/checkout/releases)
  - [actions/setup-go](https://github.com/actions/setup-go/releases)
  - [golangci/golangci-lint-action](https://github.com/golangci/golangci-lint-action/releases)
- Update quarterly or when security advisories are published
- Get SHA for specific version: `gh api repos/actions/checkout/git/refs/tags/v4.5.0 --jq '.object.sha'`

### Performance Optimizations

- **Automatic caching**: `actions/setup-go` with `cache: true` caches `GOCACHE` and `GOMODCACHE`
- **golangci-lint-action**: More efficient than manual installation, includes result caching
- **Parallel execution**: Run tests and linting concurrently to reduce CI time

### Matrix Builds (Advanced)

For libraries or cross-platform tools, consider testing multiple Go versions:

```yaml
strategy:
  matrix:
    go-version: [1.23, 1.24]
    os: [ubuntu-latest, windows-latest, macos-latest]
```

### Local Testing

Install [`act`](https://github.com/nektos/act) to test GitHub Actions workflows locally:

```bash
brew install act
act -j test  # Run specific job
act          # Run entire workflow
```

## General Guidelines for Go

- Always run `goimports -w .` on Go code files after making changes
- Prefer `task build` over `go build` to ensure all tasks are run
  - `task build` will run tests, linters, and formatters
- Functions that are easily unit-testable, should have tests
- Don't go for 100% test coverage, test the critical parts of the code
- Always use `any` instead of `interface{}` for empty interfaces
- **Error handling**: Always use `errors.Is` and `errors.As` for robust error handling
  - Use `errors.Is(err, target)` instead of `err == target` for error comparison
  - Use `errors.As(err, &target)` instead of type assertions `err.(*Type)`
  - Create sentinel errors with `var ErrNotFound = errors.New("not found")`
  - Add `Unwrap() error` method to custom error types for proper error chaining
  - These functions work correctly with wrapped errors, unlike direct comparisons
