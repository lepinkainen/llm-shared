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

## Features

- AST parsing for accuracy
- LLM-optimized compact format
- Sorted by file then line number
- Python: handles async, decorators, type hints
