# Function Analyzers

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
