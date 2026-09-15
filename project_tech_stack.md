# Project Tech Stack and Guidelines

## Git use

My Github repository root is at <https://github.com/lepinkainen/>

- IMPORTANT: NEVER commit to the "main" or "master" branch directly
- Use feature branches for new features or bug fixes
- Keep commits small and focused
- Write clear, descriptive commit messages
- Rebase branches before merging to keep history clean
- Use pull requests for code reviews and discussions
- **Commit messages, PR descriptions and review comments are not a
  documentation store.** Not because they get lost, but because nobody reads
  them. Humans read the file they are editing; agents read the files they are
  handed and grep the tree. Neither goes digging through `git log`, and nobody
  reopens a months-old review thread to find out why a line is the way it is.
  Assume anything recorded *only* in a commit message or a PR comment will
  never be read again. If a future reader needs the reasoning to change the
  code safely, put it in the code as a comment, or in the project's docs.
  Commit messages carry the narrative of a change; they are not its record.

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
- Consult `versions.md` for the canonical language runtime and GitHub Actions versions before updating tooling

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
    - build tasks need to run lint and test first — but call them **sequentially
      from `cmds:`, never via `deps:`**. Task v3 runs deps in parallel, and a
      lint task that rewrites files (`go mod tidy`, `go fix`, `goimports -w`)
      racing `go test` fails with a spurious `missing go.sum entry`. See
      `templates/Taskfile.yml` and `languages/go.md`.
  - All build artefacts should be placed in the `build/` directory if the language builds to a binary
  - Projects should have a basic Github Actions setup that uses the build-ci task to run tests and linting on push and pull requests
  - See `templates/github/workflows/` for CI templates for Go, Python, and JavaScript projects
  - Always keep `.gitignore` up to date with the language-specific ignores so that build artefacts and temporary files are not committed
  - See `templates/gitignore-*` files for language-specific .gitignore templates
  - When doing HTTP requests, use a custom user agent that includes the project name and version, e.g. `MyProject/1.0.0`
  - See `templates/Taskfile.yml` for a comprehensive example template that follows these guidelines

## Docker Deployment

For containerized applications (web services, APIs, long-running daemons):

- **Documentation**: See `docker.md` for comprehensive Docker deployment guide
- **Templates**: Use `templates/docker/` for starter files:
  - `Dockerfile-go-pure` - For pure Go applications (no CGO)
  - `Dockerfile-go-cgo` - For applications requiring CGO (e.g., mattn/go-sqlite3)
  - `docker-compose.yml` - Local development setup
  - `github-workflows-docker.yml` - Automated builds to GHCR
  - `.dockerignore` - Optimize build context

**Key principles:**

- Use multi-stage builds for minimal image sizes
- Prefer pure Go (modernc.org/sqlite) over CGO when possible
- Always include a `/health` endpoint for health checks
- Use named volumes for data persistence
- Run containers as non-root user
- Host images on GitHub Container Registry (GHCR)

**Quick start:**

```bash
# Copy appropriate template
cp llm-shared/templates/docker/Dockerfile-go-pure ./Dockerfile
cp llm-shared/templates/docker/docker-compose.yml ./
cp llm-shared/templates/docker/.dockerignore ./

# Copy GitHub Actions workflow
cp llm-shared/templates/docker/github-workflows-docker.yml .github/workflows/docker-build.yml
```

## Language-Specific Guidelines

For detailed language-specific guidelines, libraries, and best practices, see:

- **[Go Guidelines](languages/go.md)** - Go libraries, tools, and conventions
- **[Python Guidelines](languages/python.md)** - Python libraries, tools, and conventions
- **[JavaScript/TypeScript Guidelines](languages/javascript.md)** - JS/TS libraries, frameworks, and tools
