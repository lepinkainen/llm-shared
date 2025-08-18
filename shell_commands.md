# Shell Commands

Modern shell tool usage guidelines for LLM assistants.

## Preferred Tools

Use these modern alternatives instead of traditional Unix tools:

### `rg` (ripgrep) instead of `grep`

Ripgrep is faster, has better defaults, and more intuitive syntax.

**Basic usage:**

```bash
# Search for text in files
rg "pattern"

# Case insensitive search
rg -i "pattern"

# Search specific file types
rg -t go "pattern"
rg -t js "pattern" 
rg -t py "pattern"

# Search with context lines
rg -C 3 "pattern"          # 3 lines before and after
rg -A 2 -B 1 "pattern"     # 2 after, 1 before

# Show only filenames with matches
rg -l "pattern"

# Count matches
rg -c "pattern"
```

**Common patterns:**

```bash
# Find function definitions
rg "^func \w+" -t go
rg "^def \w+" -t py
rg "^function \w+" -t js

# Find TODO comments
rg "TODO|FIXME|HACK"

# Search in specific directories
rg "pattern" src/
rg "pattern" --glob "*.go"

# Exclude files/directories
rg "pattern" --glob "!test*"
```

### `fd` (fd-find) instead of `find`

Fd is simpler, faster, and respects `.gitignore` by default.

**Basic usage:**

```bash
# Find files by name
fd filename
fd "*.go"
fd "test.*\.js"

# Find directories
fd -t d dirname

# Include hidden files
fd -H filename

# Include ignored files (override .gitignore)
fd -I filename

# Search in specific directory
fd pattern src/

# Execute command on results
fd "\.go$" -x gofmt -w {}
```

**Common patterns:**

```bash
# Find source files
fd "\.(go|py|js|ts)$"

# Find test files
fd "test.*\.(go|py|js)$"
fd "_test\.go$"

# Find configuration files
fd "config\.(json|yaml|yml|toml)$"

# Find by extension
fd -e go
fd -e py
fd -e js

# Find recently modified files
fd -t f --changed-within 1d
```

## Integration with Development Workflows

**Code exploration:**

```bash
# Find all functions in Go code
rg "^func " -t go

# Find all imports
rg "^import " -t go
rg "^from .* import" -t py

# Find specific patterns across codebase
rg "http\.Client" -t go
rg "async function" -t js
```

**File discovery:**

```bash
# Find all test files
fd "_test\.go$"
fd "test_.*\.py$"
fd ".*\.test\.js$"

# Find configuration files
fd "config" -t f
fd "\.(yml|yaml|json|toml)$"

# Find documentation
fd "README"
fd "\.(md|rst)$"
```

## Benefits

- **Performance**: Both tools are significantly faster than their traditional counterparts
- **Smart defaults**: Automatically ignore binary files, respect `.gitignore`
- **Better output**: Colored output, more readable results
- **Simpler syntax**: More intuitive command-line options
- **Regular expressions**: Built-in regex support with better performance
