# GitHub Issue Management for LLMs

This guide describes how LLMs and automation agents can use the GitHub CLI (`gh`) to manage issues and labels in a repository, as well as how to read and process issues for project workflows.

---

## Prerequisites

- The `gh` CLI must be installed and authenticated (`gh auth login`).
- The working directory should be the root of the target repository.

---

## Creating Labels

Before creating issues with custom labels, ensure the labels exist by running `gh label list`. If a label does not exist, you can create it using the `gh label create` command.

You can quickly initialize a repository with a common set of LLM/project management labels by running:

```sh
./llm-shared/create-gh-labels.sh
```

This script will create all recommended labels (skipping any that already exist).

To create a single label manually:

```sh
# Create a label (if it does not exist)
gh label create <label-name> --description "<description>" --color <hexcolor>
```

**Example:**

```sh
gh label create high-priority --description "High priority issue" --color FF0000
gh label create config --description "Configuration-related issue" --color 0366D6
gh label create reddit --description "Reddit provider or integration" --color FF4500
```

---

## Issue State and Workflow Labels

To track the state of issues as they move through your workflow, use these labels:

- `needs-plan`: Issue needs an implementation plan or refinement.
- `ready`: Issue is ready for implementation (plan is complete).
- `in-progress`: Work is actively being done on the issue.
- `review`: Implementation is complete, needs review or testing.
- `blocked`: Work cannot proceed due to a dependency or missing info.

**How to use state labels:**

- When an issue is first created, add `needs-plan` if it requires further planning.
- Once an implementation plan is added, replace `needs-plan` with `ready`.
- When work begins, replace `ready` with `in-progress`.
- When work is complete and needs review, replace `in-progress` with `review`.
- If the issue is blocked at any stage, add `blocked` (and remove it when unblocked).
- When the issue is resolved, close it

You can update labels using the CLI:

```sh
# Add or remove labels on an issue
gh issue edit <issue-number> --add-label <label> --remove-label <label>
```

---

## Creating Issues

To create a new issue with a title, body, and labels:

```sh
# Create an issue from a markdown file
gh issue create --title "<title>" --body-file <path-to-md> --label <label1>,<label2>
```

**If the task is not outlined in a markdown file:**

You can provide the issue body directly using the `--body` flag:

```sh
# Create an issue with an inline body
gh issue create --title "<title>" --body "<issue body text>" --label <label1>,<label2>
```

**Example:**

```sh
gh issue create --title "Add new provider interface" --body "Implement a new provider interface for XYZ API." --label enhancement,provider
```

If a label does not exist, create it first (see above).

---

## Reading Issues

To list and read issues:

```sh
# List all open issues
gh issue list

# View a specific issue by number
gh issue view <issue-number>
```

To filter or search issues:

```sh
# List issues with a specific label
gh issue list --label <label-name>

# Search issues by keyword
gh issue list --search "<keyword>"
```

---

## Notes for LLMs

- Always check if a label exists before using it; create it if missing.
- Use `--body-file` to include detailed markdown content from local files, or `--body` for inline text.
- Use `gh issue list` and `gh issue view` to programmatically read and process issues for task management.
- All commands should be run from the repository root for correct context.

---

For more, see: https://cli.github.com/manual/
