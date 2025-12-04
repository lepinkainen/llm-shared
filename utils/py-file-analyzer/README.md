# Python File Analyzer

A tool similar to `go-file-analyzer` but for Python codebases. Analyzes Python files for size, complexity, and function lengths to identify refactoring opportunities.

## Usage

```bash
# Analyze current directory (excluding test files)
python llm-shared/utils/py-file-analyzer/main.py -dir . -n 10 -topfuncs 3

# Include test files in analysis
python llm-shared/utils/py-file-analyzer/main.py -dir . --include-tests
```

## Features

- **File size analysis**: Lists Python files by line count
- **Function metrics**: Counts lines and statements per function
- **Class and method tracking**: Identifies classes and method counts
- **Complexity buckets**: Categorizes functions by length
- **Refactoring notes**: Highlights potential issues:
  - Files with too many classes (>3)
  - Long functions (>50 lines)
  - Very long functions (>100 lines)
  - Complex functions (>20 statements)
  - Mixed async/sync patterns
  - Mixed web and database concerns
- **LLM-optimized output**: Compact format suitable for code analysis
- **Smart filtering**: Excludes test files, __init__.py files, and generated code by default

## Output Format

```
    452 subtrans/cli.py
    classes: 0
    methods: top-level=3
    buckets: >100=1, 50-100=1, 20-49=0, <20=1
    funcs (top 3 by lines):
      156 lines | stmts=25 | process_movie_phase1
       89 lines | stmts=12 | process_movie_phase2
       37 lines | stmts=8 | main
    note: Long functions: 1 >50 lines
```

## Options

- `-dir`: Directory to scan (default: current directory)
- `-n`: Number of files to display (default: 20)
- `-topfuncs`: Functions to list per file (default: 5)
- `--include-tests`: Include test files (test_*.py and *_test.py) in analysis (default: false)