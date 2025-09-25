#!/usr/bin/env python3
"""Update versions.md with latest language and GitHub Actions versions."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
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
        "url": "https://go.dev/VERSION?m=text",
        "type": "text",
        "notes": "https://go.dev/dl/",
    },
    "Python": {
        "url": "https://api.github.com/repos/python/cpython/releases?per_page=10",
        "type": "github_release_list",
        "notes": "https://www.python.org/downloads/",
    },
}

GITHUB_ACTIONS = [
    ("actions/checkout", "https://github.com/actions/checkout"),
    ("actions/setup-go", "https://github.com/actions/setup-go"),
    ("arduino/setup-task", "https://github.com/arduino/setup-task"),
    ("golangci/golangci-lint-action", "https://github.com/golangci/golangci-lint-action"),
    ("codecov/codecov-action", "https://github.com/codecov/codecov-action"),
    ("pnpm/action-setup", "https://github.com/pnpm/action-setup"),
    ("actions/setup-node", "https://github.com/actions/setup-node"),
    ("astral-sh/setup-uv", "https://github.com/astral-sh/setup-uv"),
    ("actions/setup-python", "https://github.com/actions/setup-python"),
    ("stefanzweifel/git-auto-commit-action", "https://github.com/stefanzweifel/git-auto-commit-action"),
    ("peter-evans/create-pull-request", "https://github.com/peter-evans/create-pull-request"),
]

try:  # Python 3.11+
    UTC = dt.UTC
except AttributeError:  # Fallback for older interpreters
    UTC = dt.timezone.utc

GH_PATH = shutil.which("gh")

if GH_PATH is None:
    raise SystemExit("gh CLI is required to run this script")

USER_AGENT = "llm-shared-version-bot/1.0"


def _gh_api(path: str) -> object:
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


def _github_headers() -> dict[str, str]:
    headers = {"User-Agent": USER_AGENT}
    token = os.getenv("GITHUB_TOKEN") or os.getenv("GH_TOKEN")
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


def _fetch_text(url: str) -> str:
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(request, timeout=30) as response:
        return response.read().decode("utf-8").strip()


def _fetch_json(url: str) -> object:
    request = urllib.request.Request(url, headers=_github_headers())
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response)


def _latest_go_version() -> str:
    text = _fetch_text(LANGUAGE_SOURCES["Go"]["url"])
    first_line = text.splitlines()[0] if text else ""
    return first_line.replace("go", "", 1)


def _latest_python_version() -> str:
    try:
        data = _fetch_json(LANGUAGE_SOURCES["Python"]["url"])
        if isinstance(data, list):
            for release in data:
                if release.get("draft") or release.get("prerelease"):
                    continue
                tag = release.get("tag_name") or release.get("name")
                if tag:
                    return tag.lstrip("v")
    except urllib.error.HTTPError as exc:
        if exc.code not in (403, 404):
            raise
        print(f"Warning: GitHub API returned {exc.code} when fetching Python releases", file=sys.stderr)
    except urllib.error.URLError as exc:
        print(f"Warning: could not reach GitHub for Python releases: {exc}", file=sys.stderr)

    fallback = _latest_action_tag("python/cpython")
    if any(marker in fallback.lower() for marker in ("alpha", "beta", "rc", "pre")):
        return "unknown"
    return fallback.lstrip("v") if fallback else "unknown"


def _latest_action_tag(repo: str) -> str:
    try:
        data = _gh_api(f"repos/{repo}/releases/latest")
        tag = data.get("tag_name") or data.get("name")
        if tag:
            return tag
    except RuntimeError:
        pass

    try:
        data = _gh_api(f"repos/{repo}/tags?per_page=1")
        if isinstance(data, list) and data:
            return data[0].get("name", "unknown")
    except RuntimeError as exc:
        print(f"Warning: {exc}", file=sys.stderr)

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
    for repo, url in GITHUB_ACTIONS:
        try:
            version = _latest_action_tag(repo)
        except Exception as exc:  # pylint: disable=broad-except
            print(f"Warning: could not fetch {repo} action version: {exc}", file=sys.stderr)
            version = "unknown"
        records.append(ActionRecord(repo=repo, version=version, url=url))
    return records


def render_markdown(languages: list[VersionRecord], actions: list[ActionRecord]) -> str:
    timestamp = dt.datetime.now(UTC).strftime("%Y-%m-%d %H:%M UTC")
    lines = ["# Toolchain Versions", "", f"_Last updated: {timestamp}_", "", "## Languages", ""]
    lines.append("| Tool | Latest Version | Source |")
    lines.append("| --- | --- | --- |")
    for record in languages:
        tool_link = _link(record.name, record.source)
        source_link = _link(_url_display(record.source), record.source)
        lines.append(f"| {tool_link} | {record.version} | {source_link} |")
    lines.extend(["", "## GitHub Actions", "", "| Action | Latest Tag |", "| --- | --- |"])
    for action in actions:
        action_link = _link(action.url, action.url)
        lines.append(f"| {action_link} | {action.version} |")
    lines.append("")
    lines.append(textwrap.dedent(
        """\
        > Run `python scripts/update_versions.py` locally to refresh this table immediately.
        """
    ).strip())
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


def _url_display(url: Optional[str]) -> str:
    if not url:
        return ""
    display = url.replace("https://", "").replace("http://", "").rstrip("/")
    return display or url


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Update versions.md with current tool versions")
    parser.add_argument("--output", default="versions.md", help="Path to versions.md (default: versions.md)")
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
