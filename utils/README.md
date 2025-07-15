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

## Features

- AST parsing for accuracy (Go, Python)
- Regex parsing for JavaScript/TypeScript (AST parsing available with optional dependencies)
- LLM-optimized compact format
- Sorted by file then line number
- Language-specific features:
  - **Go**: Full AST parsing, methods, receivers, type information
  - **Python**: Async functions, decorators, type hints, class methods
  - **JavaScript/TypeScript**: Arrow functions, async/await, class methods, generators
