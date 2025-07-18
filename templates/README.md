# Project Name

Brief description of what this project does.

## Features

- Feature 1
- Feature 2
- Feature 3

## Installation

```bash
# Clone the repository
git clone https://github.com/username/project-name.git
cd project-name

# Install dependencies (choose based on your project type)
# For Go projects:
go mod download

# For Python projects:
uv sync

# For JavaScript/TypeScript projects:
pnpm install
```

## Usage

```bash
# Run the project
task dev

# Build the project
task build

# Run tests
task test

# Run linting
task lint
```

## Development

This project uses [Task](https://taskfile.dev/) for build automation. See `Taskfile.yml` for available tasks.

### Requirements

- Go 1.21+ / Python 3.11+ / Node.js 18+
- Task (install from https://taskfile.dev/)

### Project Structure

```plain
project-name/
├── cmd/           # Main applications (Go)
├── internal/      # Private application code (Go)
├── pkg/          # Public library code (Go)
├── src/          # Source code (Python/JavaScript)
├── tests/        # Test files
├── docs/         # Documentation
├── build/        # Build artifacts
├── Taskfile.yml  # Build automation
└── README.md     # This file
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Testing

Run the test suite with:

```bash
task test
```

For CI testing:

```bash
task test-ci
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes and version history.
