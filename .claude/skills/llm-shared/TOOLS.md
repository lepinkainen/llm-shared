# Code Analysis Tools Reference

## validate-docs - Project Structure Validator

Validates that projects follow standard conventions.

```bash
# Auto-detect project type and validate
go run llm-shared/utils/validate-docs/validate-docs.go --auto-detect

# Generate config template for specific project type
go run llm-shared/utils/validate-docs/validate-docs.go --init go
go run llm-shared/utils/validate-docs/validate-docs.go --init python
go run llm-shared/utils/validate-docs/validate-docs.go --init node

# Validate with custom config
go run llm-shared/utils/validate-docs/validate-docs.go --config .doc-validator.yml
```

**Features:**
- Auto-detects Go, Node.js, Python, Rust projects
- Validates standard directory structure
- Checks required files (go.mod, requirements.txt, etc.)
- Validates Taskfile.yml configuration
- Checks code patterns using gofuncs/pyfuncs
- Validates Go interface implementations

**Output:** Colored validation results with summary (success/warning/error)

## gofuncs - Go Function Extractor

Lists all functions and methods in a Go project.

```bash
go run llm-shared/utils/gofuncs/gofuncs.go -dir /path/to/project
```

**Output format:** `file:line:type:exported:name:receiver:signature`

**Field meanings:**
- `type`: `f` = function, `m` = method
- `exported`: `y` = public (uppercase), `n` = private (lowercase)
- `receiver`: Empty for functions, type name for methods

**Example output:**
```
api.go:15:f:n:fetchItems:():[]Item
config.go:144:m:y:GetCategory:*Mapper:(string)string
```

## pyfuncs - Python Function Extractor

Lists all functions and methods in a Python project.

```bash
python3 llm-shared/utils/pyfuncs.py --dir /path/to/project
```

**Output format:** `file:line:type:exported:name:class:signature:decorators`

**Field meanings:**
- `type`: `f` = function, `m` = method, `s` = staticmethod, `c` = classmethod, `p` = property
- `exported`: `y` = public, `n` = private (underscore prefix)
- `class`: Empty for top-level functions, class name for methods
- `decorators`: Comma-separated decorator names

**Example output:**
```
main.py:15:f:y:process_data::(data:List[str])->Dict[str,int]:
api.py:45:m:y:fetch:APIClient:async (url:str)->Response:cache,retry
utils.py:23:s:y:helper:Utils:(value:int)->str:staticmethod
```

## jsfuncs - JavaScript/TypeScript Function Extractor

Lists all functions and methods in a JS/TS project.

```bash
node llm-shared/utils/jsfuncs.js --dir /path/to/project
```

**Output format:** `file:line:type:exported:name:class:signature:decorators`

**Field meanings:**
- `type`: `f` = function, `m` = method, `a` = arrow function, `c` = constructor, `g` = getter, `s` = setter
- `exported`: `y` = public/exported, `n` = private

**Example output:**
```
main.js:15:f:y:processData::(data:string[])=>Promise<Object>:
api.ts:45:m:y:fetch:APIClient:async (url:string)=>Response:
utils.js:23:a:y:helper::(value:number)=>string:
```

## py-file-analyzer - Python Complexity Analyzer

Analyzes Python files for size and complexity.

```bash
python3 llm-shared/utils/py-file-analyzer/main.py -dir /path/to/project -n 20 -topfuncs 5
```

**Options:**
- `-dir`: Directory to analyze
- `-n`: Number of files to show (default: 20)
- `-topfuncs`: Number of top functions per file (default: 5)

**Output includes:**
- Total lines per file
- Class and method counts
- Function length buckets
- Top functions by lines and statements
- Refactoring suggestions
- Mixed concern detection (web + database, async + sync)

## go-file-analyzer - Go Complexity Analyzer

Analyzes Go files for size and complexity.

```bash
go run llm-shared/utils/go-file-analyzer/main.go -dir /path/to/project -n 20 -topfuncs 5
```

**Options:**
- `-dir`: Directory to analyze
- `-n`: Number of files to show (default: 20)
- `-topfuncs`: Number of top functions per file (default: 5)

**Output includes:**
- Total lines per file
- Type and method counts
- Length buckets
- Top functions by lines and statements
- Refactoring notes for crowded files
