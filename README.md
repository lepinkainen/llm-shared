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

**Versions**: Run `python scripts/update_versions.py` (or rely on the scheduled GitHub Action) to refresh `versions.md`, which tracks recommended Go, Python, and GitHub Action versions.

**Repository hooks**:

If the repository uses hooks, enable this setting:

```sh
git config core.hooksPath scripts/hooks
```

Then all hooks should be in the `scripts/hooks` directory.

**HTTP API**:

- All HTTP APIs MUST have a `/whoami` endpoint that returns the project name, git hash, build time and version in JSON format, e.g. `{"name":"MyProject","version":"1.0.0", "hash":"abc123", "build_time":"2024-01-01T00:00:00Z"}`.
- When something is running in the port, you MUST use the `/whoami` endpoint to identify it before attempting to start it or kill it.

example:

```go
func (h *Handler) Whoami(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]string{
    "name":       h.AppName,
    "version":    h.Version,
    "hash":       h.GitHash,
    "build_time": h.BuildTime,
  })
}
```
