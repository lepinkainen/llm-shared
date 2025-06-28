# Libraries to use

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

### Libraries

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

### Libraries

- Use "uv" for dependency management (similar to pipenv)
- Use "ruff" for linting and formatting
- Use "mypy" for type checking as much as possible
