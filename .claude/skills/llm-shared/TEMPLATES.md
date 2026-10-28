# Project Templates Reference

Templates are located in `llm-shared/templates/`.

## Taskfile.yml

Task runner configuration (taskfile.dev - modern Make alternative).

```bash
cp llm-shared/templates/Taskfile.yml ./Taskfile.yml
```

**Required tasks:**
- `build` - Build project (depends on test, lint)
- `build-linux` - Cross-compile for Linux
- `build-ci` - CI build (compile only)
- `test` - Run tests
- `test-ci` - CI tests with build tags
- `lint` - Run linters
- `clean` - Remove build artifacts

## Docker Templates

Located in `templates/docker/`.

### For Pure Go Applications (No CGO)

```bash
cp llm-shared/templates/docker/Dockerfile-go-pure ./Dockerfile
cp llm-shared/templates/docker/docker-compose.yml ./
cp llm-shared/templates/docker/.dockerignore ./
```

### For CGO Applications

```bash
cp llm-shared/templates/docker/Dockerfile-go-cgo ./Dockerfile
```

### GitHub Actions for Docker

```bash
mkdir -p .github/workflows
cp llm-shared/templates/docker/github-workflows-docker.yml .github/workflows/docker-build.yml
```

## GitHub Actions CI Templates

Located in `templates/github/workflows/`.

### Go Projects

```bash
mkdir -p .github/workflows
cp llm-shared/templates/github/workflows/go-ci.yml .github/workflows/ci.yml
```

For advanced multi-version testing:
```bash
cp llm-shared/templates/github/workflows/go-ci-advanced.yml .github/workflows/ci.yml
```

### Python Projects

```bash
mkdir -p .github/workflows
cp llm-shared/templates/github/workflows/python-ci.yml .github/workflows/ci.yml
```

### JavaScript/TypeScript Projects

```bash
mkdir -p .github/workflows
cp llm-shared/templates/github/workflows/js-ci.yml .github/workflows/ci.yml
```

### Dependabot Configuration

```bash
mkdir -p .github
cp llm-shared/templates/github/workflows/dependabot.yml .github/dependabot.yml
```

## Linter Configurations

### Go - golangci-lint

```bash
cp llm-shared/templates/golangci.yml ./.golangci.yml
```

**Enabled linters:**
- gofmt, goimports, govet, errcheck
- staticcheck, unused, gosimple
- ineffassign, typecheck, misspell
- gocritic, revive

## .gitignore Templates

```bash
# Go projects
cp llm-shared/templates/gitignore-go ./.gitignore

# Python projects
cp llm-shared/templates/gitignore-python ./.gitignore

# JavaScript/TypeScript projects
cp llm-shared/templates/gitignore-js ./.gitignore
```

## Changelog Template

```bash
cp llm-shared/templates/CHANGELOG.md ./CHANGELOG.md
```

## GitHub Labels

Set up standard issue labels:

```bash
./llm-shared/create-gh-labels.sh
```

Creates labels: `needs-plan`, `ready`, `in-progress`, `review`, `bug`, `enhancement`, `documentation`
