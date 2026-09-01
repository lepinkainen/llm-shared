# llm-shared — Project Purpose

## What it is

The shared conventions layer for every project in this workspace. It is vendored into ~30 repos as a git submodule and holds the guidelines an AI assistant (or a returning human) needs to work consistently across all of them: preferred libraries, build and validation practices, shell tooling, Docker patterns and GitHub workflow.

It is documentation and small stdlib-only utilities — not an application.

## Current capabilities

- `project_tech_stack.md` — universal development guidelines: project management, validation, common practices
- `languages/` — per-language conventions (`go.md`, `python.md`, `javascript.md`)
- `GITHUB.md` — issue creation and management conventions
- `shell_commands.md` — modern shell tooling (`rg` over `grep`, `fd` over `find`)
- `docker.md` — multi-stage builds, GHCR, docker-compose patterns
- `versions.md`, `templates/`, `scripts/`, `create-gh-labels.sh` — version pins, CI templates and setup helpers
- `utils/` — small stdlib-only helpers such as `gofuncs.go`

## Direction

Stay project-independent. Anything that only makes sense for one repo belongs in that repo's own `CLAUDE.md` or `PROJECT.md`, not here.

## Constraints

- **This repo is consumed as a submodule and must not be edited from within a consuming project.** Changes are made here, in the canonical repo, and pulled downstream.
- **Stdlib-only.** No dependencies may be added to code in this repo.
- Consuming projects exclude `llm-shared/` from their tests, linters and searches.
- The convention this file participates in is defined here: `PROJECT.md` in a project root states purpose, current state and intended direction, and is the first thing to read.
