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
  - Projects should have a basic Github Actions setup that uses the build-ci task to run tests and linting on push and pull requests

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
- Command-line arguments: "github.com/alecthomas/kong" (only if the project requires complex CLI)

### General Guidelines for Go

- Always run "gofmt -w ." on the Go code files after making changes
- Prefer `task build` over `go build` to ensure all tasks are run
- Functions that are easily unit-testable, should have tests
- Don't go for 100% test coverage, test the critical parts of the code

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

## JavaScript/TypeScript

When looking for functions, use the `jsfuncs` tool to list all functions in a JavaScript/TypeScript project. It provides a compact format that is optimized for LLM context.

- Example usage:

  ```bash
  node jsfuncs.js --dir /path/to/project
  ```

### JavaScript/TypeScript Libraries and Tools

#### Package Management

- Use "pnpm" for package management (faster than npm/yarn)
- Alternative: "npm" or "yarn" for simpler projects

#### Development Tools

- **TypeScript**: Always prefer TypeScript over JavaScript for new projects
- **ESLint**: For linting (`@typescript-eslint/parser` for TS projects)
- **Prettier**: For code formatting
- **Vitest**: For testing (modern alternative to Jest)
- **tsx**: For running TypeScript directly (`npx tsx script.ts`)

#### Framework Preferences

- **Node.js Backend**: Fastify (lightweight) or Express.js (established)
- **Frontend**: React with TypeScript, Next.js for full-stack
- **Build Tools**: Vite (modern) or esbuild (fast)

#### Database Libraries

- **SQL**: Drizzle ORM (type-safe) or Prisma (feature-rich)
- **SQLite**: better-sqlite3
- **PostgreSQL**: pg with @types/pg

#### Utility Libraries

- **Date handling**: date-fns (modular) or dayjs (lightweight)
- **Validation**: zod (TypeScript-first schema validation)
- **Environment**: dotenv for configuration
- **Logging**: pino (fast structured logging)

### General Guidelines for Javascript/TypeScript

- Always use TypeScript for new projects
- Configure strict TypeScript settings in tsconfig.json
- Use ES modules (import/export) over CommonJS
- Prefer async/await over promises and callbacks
- Use meaningful variable and function names
- Always build/compile before finishing a task
