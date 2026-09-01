#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.12"
# ///
"""Update versions.md with latest language and GitHub Actions versions.

Standard library only, by design: this repository takes no dependencies. The
inline metadata block above exists so the file runs as-is via `uv run` or
directly as `./scripts/update_versions.py`.

Requires the `gh` CLI on PATH for GitHub Action release tags.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import shutil
import subprocess
import sys
import textwrap
import time
import urllib.error
import urllib.request
from dataclasses import dataclass

LANGUAGE_SOURCES = {
    "Go": {
        "id": "go",
        "notes": "https://go.dev/dl/",
    },
    "Python": {
        "id": "python",
        "notes": "https://www.python.org/downloads/",
    },
    "Tailwind CSS": {
        "id": "tailwind-css",
        "notes": "https://tailwindcss.com",
    },
    "Node.js": {
        "id": "nodejs",
        "notes": "https://nodejs.org/en/download/",
    },
}

GITHUB_BASE = "https://github.com"
GITHUB_ACTIONS = [
    "actions/checkout",
    "actions/setup-go",
    "go-task/setup-task",
    "golangci/golangci-lint-action",
    "codecov/codecov-action",
    "pnpm/action-setup",
    "actions/setup-node",
    "astral-sh/setup-uv",
    "actions/setup-python",
    "stefanzweifel/git-auto-commit-action",
    "peter-evans/create-pull-request",
]

try:  # Python 3.11+
    UTC = dt.UTC
except AttributeError:  # Fallback for older interpreters
    UTC = dt.UTC

GH_PATH = shutil.which("gh")

if GH_PATH is None:
    raise SystemExit("gh CLI is required to run this script")

USER_AGENT = "llm-shared-version-bot/1.0"

END_OF_LIFE_URL = "https://endoflife.date/api/v1/products/{id}/releases/latest"


ATTEMPTS = 3
RETRY_DELAY_SECONDS = 2

TABLE_ROW = re.compile(r"^\|\s*\[([^\]]+)\]\([^)]*\)\s*\|\s*(\S+)\s*\|\s*$")


def _retry(what: str, fn):
    """Call fn, retrying transient failures.

    The upstream APIs flake often enough that a single failed call used to
    overwrite a known-good version with "unknown", and the scheduled workflow
    would then open a pull request erasing a pin.
    """
    last: Exception | None = None
    for attempt in range(1, ATTEMPTS + 1):
        try:
            return fn()
        except Exception as exc:  # pylint: disable=broad-except
            last = exc
            if attempt < ATTEMPTS:
                print(
                    f"Retrying {what} after error ({attempt}/{ATTEMPTS}): {exc}",
                    file=sys.stderr,
                )
                time.sleep(RETRY_DELAY_SECONDS * attempt)
    raise RuntimeError(f"{what} failed after {ATTEMPTS} attempts: {last}")


def read_existing_versions(path: str) -> dict[str, str]:
    """Map the label in each existing table row to its recorded version.

    Used as a fallback so a failed lookup keeps the previous value rather than
    degrading the table to "unknown".
    """
    try:
        with open(path, encoding="utf-8") as handle:
            text = handle.read()
    except OSError:
        return {}

    existing: dict[str, str] = {}
    for line in text.splitlines():
        match = TABLE_ROW.match(line)
        if match:
            label, version = match.group(1), match.group(2)
            if version != "unknown":
                existing[label] = version
    return existing


def _gh_api(path: str) -> dict:
    if not GH_PATH:
        raise SystemExit("gh CLI is required to run this script")
    try:
        result = subprocess.run(
            [GH_PATH, "api", path],
            check=True,
            capture_output=True,
            text=True,
        )
    except subprocess.CalledProcessError as exc:
        raise RuntimeError(f"gh api {path} failed: {exc}") from exc

    try:
        return json.loads(result.stdout)
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"gh api {path} returned non-JSON output") from exc


def _fetch_json(url: str, headers: dict[str, str] | None = None) -> dict:
    if headers is None:
        headers = {"User-Agent": USER_AGENT}
    request = urllib.request.Request(url, headers=headers)
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response)


def _latest_endoflife_version(product: str) -> str:
    url = END_OF_LIFE_URL.format(id=product)
    data = _fetch_json(url, headers={"User-Agent": USER_AGENT})
    if not isinstance(data, dict):
        return "unknown"

    result = data.get("result")
    if not isinstance(result, dict):
        return "unknown"

    latest = result.get("latest")
    if isinstance(latest, dict):
        name = latest.get("name")
        if isinstance(name, str) and name.strip():
            return name.strip()

    for key in ("name", "label"):
        value = result.get(key)
        if isinstance(value, str) and value.strip():
            return value.strip()

    return "unknown"


def _latest_action_tag(repo: str) -> str:
    data = _gh_api(f"repos/{repo}/releases/latest")
    tag = data.get("tag_name") or data.get("name")
    if tag:
        return tag
    return "unknown"


@dataclass
class VersionRecord:
    name: str
    version: str
    source: str


@dataclass
class ActionRecord:
    repo: str
    version: str
    url: str


def _resolve(label: str, previous: dict[str, str], fn) -> str:
    """Look up a version, falling back to the previously recorded one.

    Returns "unknown" only when the lookup fails and no earlier value exists.
    """
    try:
        return _retry(f"lookup for {label}", fn)
    except Exception as exc:  # pylint: disable=broad-except
        kept = previous.get(label)
        if kept:
            print(
                f"Warning: could not fetch {label} version ({exc}); keeping {kept}",
                file=sys.stderr,
            )
            return kept
        print(f"Warning: could not fetch {label} version: {exc}", file=sys.stderr)
        return "unknown"


def collect_language_versions(previous: dict[str, str]) -> list[VersionRecord]:
    records: list[VersionRecord] = []
    for name, meta in LANGUAGE_SOURCES.items():
        version = _resolve(
            name, previous, lambda meta=meta: _latest_endoflife_version(meta["id"])
        )
        records.append(VersionRecord(name=name, version=version, source=meta["notes"]))
    return records


def collect_action_versions(previous: dict[str, str]) -> list[ActionRecord]:
    records: list[ActionRecord] = []
    for repo in GITHUB_ACTIONS:
        version = _resolve(repo, previous, lambda repo=repo: _latest_action_tag(repo))
        url = f"{GITHUB_BASE}/{repo}"
        records.append(ActionRecord(repo=repo, version=version, url=url))
    return records


def render_markdown(languages: list[VersionRecord], actions: list[ActionRecord]) -> str:
    timestamp = dt.datetime.now(UTC).strftime("%Y-%m-%d %H:%M UTC")
    lines = [
        "# Toolchain Versions",
        "",
        f"_Last updated: {timestamp}_",
        "",
        "## Languages",
        "",
    ]
    lines.append("| Tool | Latest Version |")
    lines.append("| --- | --- |")
    for record in languages:
        tool_link = _link(record.name, record.source)
        lines.append(f"| {tool_link} | {record.version} |")
    lines.extend(
        ["", "## GitHub Actions", "", "| Action | Latest Tag |", "| --- | --- |"]
    )
    for action in actions:
        action_link = _link(action.repo, action.url)
        lines.append(f"| {action_link} | {action.version} |")
    lines.append("")
    lines.append(
        textwrap.dedent(
            """\
        > Run `./scripts/update_versions.py` locally to refresh this table immediately.
        """
        ).strip()
    )
    lines.append("")
    return "\n".join(lines)


def write_versions_file(path: str) -> None:
    previous = read_existing_versions(path)
    languages = collect_language_versions(previous)
    actions = collect_action_versions(previous)
    markdown = render_markdown(languages, actions)
    with open(path, "w", encoding="utf-8") as handle:
        handle.write(markdown + "\n")


def _link(label: str, url: str | None) -> str:
    if not url:
        return label
    return f"[{label}]({url})"


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Update versions.md with current tool versions"
    )
    parser.add_argument(
        "--output",
        default="versions.md",
        help="Path to versions.md (default: versions.md)",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    try:
        write_versions_file(args.output)
    except Exception as exc:  # pylint: disable=broad-except
        print(f"Error: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
