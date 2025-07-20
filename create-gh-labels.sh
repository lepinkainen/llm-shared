#!/usr/bin/env bash
# llm-shared/create-gh-labels.sh
# Usage: ./llm-shared/create-gh-labels.sh
# Creates a set of common labels for LLM-driven GitHub project management.

set -e

gh label create llm-task --description "Task created or managed by LLM" --color 5319e7 || true
gh label create high-priority --description "High priority issue" --color FF0000 || true
gh label create medium-priority --description "Medium priority issue" --color FFA500 || true
gh label create low-priority --description "Low priority issue" --color 1D76DB || true
# Issue state labels
gh label create needs-plan --description "Needs implementation plan or refinement" --color 5319e7 || true
gh label create ready --description "Ready for implementation" --color 0e8a16 || true
gh label create in-progress --description "Task in progress" --color 0052cc || true
gh label create blocked --description "Task is blocked" --color b60205 || true
gh label create review --description "Needs review or testing" --color fbca04 || true
# Removed 'done' label as closed issues are considered done
# Other useful labels
gh label create bug --description "Bug report" --color d73a4a || true
gh label create enhancement --description "Enhancement or feature" --color a2eeef || true
gh label create documentation --description "Documentation" --color 0075ca || true
gh label create config --description "Configuration-related issue" --color 0366D6 || true
gh label create provider --description "Provider-specific issue" --color 5319e7 || true
gh label create needs-info --description "Needs more information" --color d876e3 || true
gh label create automation --description "Created by automation or bot" --color 8a2be2 || true
gh label create duplicate --description "Duplicate issue" --color cfd3d7 || true
gh label create wontfix --description "Won't fix" --color ffffff || true
gh label create invalid --description "Invalid issue" --color e4e669 || true
#gh label create good-first-issue --description "Good for newcomers" --color 7057ff || true
#gh label create help-wanted --description "Help wanted" --color 008672 || true

echo "All common labels created (existing labels were skipped)."
