#!/usr/bin/env python3
"""Update versions.md with latest language and GitHub Actions versions."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import shutil
import subprocess
import sys
import textwrap
import urllib.error
import urllib.request
from dataclasses import dataclass
from typing import Optional

LANGUAGE_SOURCES = {
    "Go": {
        "url": "https://endoflife.date/api/v1/products/go/releases/latest",
        "type": "endoflife_latest",
        "notes": "https://go.dev/dl/",
    },
    "Python": {
        "url": "https://endoflife.date/api/v1/products/python/releases/latest",
        "type": "endoflife_latest",
        "notes": "https://www.python.org/downloads/",
    },
}

GITHUB_BASE = "https://github.com"
GITHUB_ACTIONS = [
    "actions/checkout",
    "actions/setup-go",
    "arduino/setup-task",
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
    UTC = dt.timezone.utc

GH_PATH = shutil.which("gh")

if GH_PATH is None:
    raise SystemExit("gh CLI is required to run this script")

USER_AGENT = "llm-shared-version-bot/1.0"

END_OF_LIFE_BASE = "https://endoflife.date/api/v1/products"


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


def _fetch_json(url: str, headers: Optional[dict[str, str]] = None) -> dict:
    if headers is None:
        headers = {"User-Agent": USER_AGENT}
    request = urllib.request.Request(url, headers=headers)
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response)


def _latest_go_version() -> str:
    return _latest_endoflife_version("go")


def _latest_python_version() -> str:
    return _latest_endoflife_version("python")


def _latest_endoflife_version(product: str) -> str:
    url = f"{END_OF_LIFE_BASE}/{product}/releases/latest"
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


def collect_language_versions() -> list[VersionRecord]:
    records: list[VersionRecord] = []
    for name, meta in LANGUAGE_SOURCES.items():
        try:
            if name == "Go":
                version = _latest_go_version()
            elif name == "Python":
                version = _latest_python_version()
            else:
                version = "unknown"
        except Exception as exc:  # pylint: disable=broad-except
            print(f"Warning: could not fetch {name} version: {exc}", file=sys.stderr)
            version = "unknown"
        records.append(VersionRecord(name=name, version=version, source=meta["notes"]))
    return records


def collect_action_versions() -> list[ActionRecord]:
    records: list[ActionRecord] = []
    for repo in GITHUB_ACTIONS:
        try:
            version = _latest_action_tag(repo)
        except Exception as exc:  # pylint: disable=broad-except
            print(
                f"Warning: could not fetch {repo} action version: {exc}",
                file=sys.stderr,
            )
            version = "unknown"
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
        > Run `python scripts/update_versions.py` locally to refresh this table immediately.
        """
        ).strip()
    )
    lines.append("")
    return "\n".join(lines)


def write_versions_file(path: str) -> None:
    languages = collect_language_versions()
    actions = collect_action_versions()
    markdown = render_markdown(languages, actions)
    with open(path, "w", encoding="utf-8") as handle:
        handle.write(markdown + "\n")


def _link(label: str, url: Optional[str]) -> str:
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


def main(argv: Optional[list[str]] = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    try:
        write_versions_file(args.output)
    except Exception as exc:  # pylint: disable=broad-except
        print(f"Error: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
