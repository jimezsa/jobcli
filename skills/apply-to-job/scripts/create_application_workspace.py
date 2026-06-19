#!/usr/bin/env python3
"""Create an application workspace folder for the apply-to-job skill.

Inputs (one of):
  --from-index PATH INDEX
      Path to jobs_yes_high.json and 0-based index of the entry to use.
  --from-json '{"title":"...", "company":"...", "url":"...", "description":"..."}'
      Inline JSON snippet.
  --url URL --title TITLE --company COMPANY [--description TEXT]
      Discrete args (description optional; can be added later).

Output: a folder under --apps-root (default `applications/`) named
`<YYYY-MM-DD>-<company-slug>-<role-slug>/` containing:
  - job.json
  - job-description.txt (empty if no description provided)
  - STATUS.md (initial state: created)
  - screenshots/ (empty)

Prints the absolute workspace path on stdout.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from datetime import date
from pathlib import Path


def slugify(value: str, max_len: int = 40) -> str:
    value = value.lower()
    value = re.sub(r"[^\w\s-]", "", value, flags=re.UNICODE)
    value = re.sub(r"[\s_-]+", "-", value).strip("-")
    return value[:max_len].strip("-") or "untitled"


def load_from_index(path: Path, index: int) -> dict:
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, list):
        raise SystemExit(f"{path} is not a JSON list")
    if index < 0 or index >= len(data):
        raise SystemExit(f"index {index} out of range (0..{len(data) - 1})")
    return data[index]


def normalize_job(raw: dict) -> dict:
    return {
        "site": raw.get("site", ""),
        "url": raw.get("url", ""),
        "title": raw.get("title", "").strip(),
        "company": raw.get("company", "").strip(),
        "location": raw.get("location", ""),
        "description": raw.get("description", ""),
        "posted_at": raw.get("posted_at", ""),
        "source": raw.get("_source", "manual"),
    }


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    src = p.add_mutually_exclusive_group(required=True)
    src.add_argument("--from-index", nargs=2, metavar=("JOBS_JSON", "INDEX"))
    src.add_argument("--from-json", metavar="JSON")
    src.add_argument("--url", metavar="URL")
    p.add_argument("--title", help="role title (with --url)")
    p.add_argument("--company", help="company (with --url)")
    p.add_argument("--description", default="", help="description text (with --url)")
    p.add_argument("--apps-root", default="applications", help="root folder for application workspaces")
    p.add_argument("--date", default=None, help="override date prefix (YYYY-MM-DD)")
    args = p.parse_args()

    if args.from_index:
        raw = load_from_index(Path(args.from_index[0]), int(args.from_index[1]))
    elif args.from_json:
        raw = json.loads(args.from_json)
    else:
        if not (args.url and args.title and args.company):
            p.error("--url requires --title and --company")
        raw = {
            "url": args.url,
            "title": args.title,
            "company": args.company,
            "description": args.description,
        }

    job = normalize_job(raw)
    if not job["title"] or not job["company"]:
        raise SystemExit("job must have non-empty title and company")

    day = args.date or date.today().isoformat()
    folder_name = f"{day}-{slugify(job['company'])}-{slugify(job['title'])}"
    apps_root = Path(args.apps_root).resolve()
    workspace = apps_root / folder_name
    if workspace.exists():
        print(f"workspace already exists: {workspace}", file=sys.stderr)
    workspace.mkdir(parents=True, exist_ok=True)
    (workspace / "screenshots").mkdir(exist_ok=True)

    (workspace / "job.json").write_text(json.dumps(job, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    (workspace / "job-description.txt").write_text(job["description"] or "", encoding="utf-8")
    (workspace / "STATUS.md").write_text("created\n", encoding="utf-8")

    print(workspace)
    return 0


if __name__ == "__main__":
    sys.exit(main())
