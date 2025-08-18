# LLM Shared

Standardized development guidelines and tools for LLM assistants.

**For LLMs**:

- Refer to `project_tech_stack.md` for universal development guidelines (project management, validation, common practices)
- Refer to `GITHUB.md` for GitHub issue management (creating, reading, and managing issues)
- Refer to `shell_commands.md` for modern shell tool usage (`rg` instead of `grep`, `fd` instead of `find`)
- Language-specific guidelines are in the `languages/` directory:
  - `languages/go.md` - Go libraries, tools, and conventions
  - `languages/python.md` - Python libraries, tools, and conventions
  - `languages/javascript.md` - JavaScript/TypeScript libraries, frameworks, and tools

**Tools**: The `utils/` directory contains function analysis tools for code exploration:

- `gofuncs.go` - Go function analyzer
- `pyfuncs.py` - Python function analyzer
- `jsfuncs.js` - JavaScript/TypeScript function analyzer
- `validate-docs.go` - Project structure validator
