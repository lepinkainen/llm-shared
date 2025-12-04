# Development Utilities

## validate-docs - Documentation Validator

```bash
go run validate-docs.go --auto-detect
go run validate-docs.go --config .doc-validator.yml
go run validate-docs.go --init go
```

**Purpose**: Validates that documentation accurately reflects the current codebase structure across multiple project types.

**Features**:

- **Auto-detection**: Automatically detects project type (Go, Node.js, Python, Rust)
- **Configuration-driven**: YAML-based validation rules with project-specific customization
- **Multi-language support**: Built-in templates for common project structures
- **Interface validation**: Checks Go interface implementations
- **Build system validation**: Validates Taskfile, Makefile, npm scripts, Cargo commands
- **Dependency checking**: Verifies expected dependencies in go.mod, package.json, requirements.txt, Cargo.toml

**Output**: Colored validation results with summary (✅ success, ⚠️ warning, ❌ error)

**Examples**:

```bash
# Auto-detect and validate current project
go run validate-docs.go --auto-detect

# Generate config template for Go project
go run validate-docs.go --init go

# Validate with custom configuration
go run validate-docs.go --config .doc-validator.yml
```

**Template configs** available in `../examples/`:

- `go-project.doc-validator.yml`
- `node-project.doc-validator.yml`
- `python-project.doc-validator.yml`
- `rust-project.doc-validator.yml`

## gofuncs - Go Function Lister

```bash
go run gofuncs.go -dir /path/to/project
```

**Output**: `file:line:type:exported:name:receiver:signature`

- **type**: `f`=function, `m`=method
- **exported**: `y`=public, `n`=private

```plain
api.go:15:f:n:fetchHackerNewsItems:()[]HackerNewsItem
config.go:144:m:y:GetCategoryForDomain:*CategoryMapper:(string)string
```

## pyfuncs - Python Function Lister

```bash
python pyfuncs.py --dir /path/to/project
```

**Output**: `file:line:type:exported:name:class:signature:decorators`

- **type**: `f`=function, `m`=method, `s`=staticmethod, `c`=classmethod, `p`=property
- **exported**: `y`=public, `n`=private (underscore prefix)

```plain
main.py:15:f:y:process_data::(data:List[str])->Dict[str,int]:
api.py:45:m:y:fetch:APIClient:async (url:str)->Response:cache,retry
utils.py:23:s:y:helper:Utils:(value:int)->str:staticmethod
```

## jsfuncs - JavaScript/TypeScript Function Lister

```bash
node jsfuncs.js --dir /path/to/project
```

**Output**: `file:line:type:exported:name:class:signature:decorators`

- **type**: `f`=function, `m`=method, `a`=arrow, `c`=constructor, `g`=getter, `s`=setter
- **exported**: `y`=public, `n`=private (underscore prefix or not module-level)

```plain
main.js:15:f:y:processData::(data:string[])=>Promise<Object>:
api.ts:45:m:y:fetch:APIClient:async (url:string)=>Response:
utils.js:23:a:y:helper::(value:number)=>string:
```

## py-file-analyzer - Python File Size & Function Report

```bash
python py-file-analyzer/main.py -dir /path/to/project -n 20 -topfuncs 5
```

**Purpose**: Lists the longest Python source files by line count, ignoring test files, `__init__.py` files, and common build artifacts.

**Output**: LLM-friendly summary per file: total lines, class count, method/top-level func counts, function length buckets, top functions (lines + statements), and refactoring notes.

**Features**:
- AST parsing for accurate function boundaries and statement counting
- Complexity analysis with configurable buckets
- Automatic refactoring suggestions
- Mixed concern detection (web + database, async + sync)
- Class and method organization analysis
- Smart filtering (excludes tests, __init__.py, generated code)

## go-file-analyzer - Go File Size & Function Report

```bash
go run ./go-file-analyzer/main.go -dir /path/to/project -n 20 -topfuncs 5
```

**Purpose**: Lists the longest Go source files by line count, ignoring `_test.go` files and common build artifacts.

**Output**: LLM-friendly summary per file: total lines, type count, method/top-level func counts, length buckets, top functions (lines + stmts), and short notes when things look crowded.

## Features

- AST parsing for accuracy (Go, Python)
- Regex parsing for JavaScript/TypeScript (AST parsing available with optional dependencies)
- LLM-optimized compact format
- Sorted by file then line number
- Language-specific features:
  - **Go**: Full AST parsing, methods, receivers, type information
  - **Python**: Full AST parsing, async functions, decorators, type hints, class methods
  - **JavaScript/TypeScript**: Arrow functions, async/await, class methods, generators
