# Python Guidelines

When looking for functions, use the `pyfuncs` tool to list all functions in a Python project. It provides a compact format that is optimized for LLM context.

- Example usage:

  ```bash
  python3 pyfuncs.py --dir /path/to/project
  ```

## Python Libraries

- Use "uv" for dependency management (similar to pipenv)
- Use "ruff" for linting and formatting
- Use "mypy" for type checking as much as possible