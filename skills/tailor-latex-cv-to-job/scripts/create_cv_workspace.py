#!/usr/bin/env python3
"""Create a derived LaTeX CV workspace without modifying the source file."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import shutil
from pathlib import Path


GRAPHICS_EXTENSIONS = [".pdf", ".png", ".jpg", ".jpeg", ".eps"]


def slugify(value: str) -> str:
    slug = re.sub(r"[^a-zA-Z0-9]+", "-", value.strip().lower()).strip("-")
    return slug or "job"


def unique_dir(root: Path, label: str) -> Path:
    date_prefix = dt.datetime.now().strftime("%Y%m%d")
    base = root / f"{date_prefix}-{slugify(label)}"
    candidate = base
    counter = 2
    while candidate.exists():
        candidate = root / f"{base.name}-{counter}"
        counter += 1
    return candidate


def possible_file(path: Path, kind: str) -> list[Path]:
    if path.suffix:
        return [path]
    if kind == "graphic":
        return [path.with_suffix(ext) for ext in GRAPHICS_EXTENSIONS]
    if kind == "bib":
        return [path.with_suffix(".bib"), path]
    return [path.with_suffix(".tex"), path]


def find_local_refs(tex_text: str) -> list[tuple[str, str]]:
    refs: list[tuple[str, str]] = []
    patterns = [
        (r"\\includegraphics(?:\[[^\]]*\])?\{([^}]+)\}", "graphic"),
        (r"\\input\{([^}]+)\}", "tex"),
        (r"\\include\{([^}]+)\}", "tex"),
        (r"\\bibliography\{([^}]+)\}", "bib"),
        (r"\\addbibresource\{([^}]+)\}", "bib"),
    ]
    for pattern, kind in patterns:
        for match in re.findall(pattern, tex_text):
            for raw_ref in match.split(","):
                ref = raw_ref.strip()
                if ref:
                    refs.append((ref, kind))
    return refs


def copy_ref(source_dir: Path, workspace: Path, ref: str, kind: str) -> dict[str, str]:
    raw = Path(ref)
    if raw.is_absolute():
        return {"ref": ref, "status": "skipped-absolute"}

    for candidate in possible_file(source_dir / raw, kind):
        if candidate.exists() and candidate.is_file():
            rel = candidate.relative_to(source_dir)
            destination = workspace / rel
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(candidate, destination)
            return {"ref": ref, "copied": str(rel), "status": "copied"}

    return {"ref": ref, "status": "missing"}


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source_tex", type=Path, help="Path to the original .tex CV.")
    parser.add_argument("--job-label", default="job", help="Company/role label for the workspace directory.")
    parser.add_argument("--out-root", type=Path, default=Path("tailored-cvs"), help="Root directory for tailored CV workspaces.")
    parser.add_argument("--tex-name", default="cv-tailored.tex", help="Filename for the derived .tex copy.")
    args = parser.parse_args()

    source_tex = args.source_tex.expanduser().resolve()
    if not source_tex.exists():
        raise SystemExit(f"source .tex does not exist: {source_tex}")
    if source_tex.suffix.lower() != ".tex":
        raise SystemExit(f"source file must be .tex: {source_tex}")

    out_root = args.out_root.expanduser()
    out_root.mkdir(parents=True, exist_ok=True)
    workspace = unique_dir(out_root, args.job_label)
    workspace.mkdir(parents=True)

    derived_tex = workspace / args.tex_name
    shutil.copy2(source_tex, derived_tex)

    source_text = source_tex.read_text(encoding="utf-8", errors="replace")
    refs = find_local_refs(source_text)
    copied_refs = []
    for ref, kind in refs:
        copied_refs.append(copy_ref(source_tex.parent, workspace, ref, kind))

    manifest = {
        "source_tex": str(source_tex),
        "workspace": str(workspace.resolve()),
        "derived_tex": str(derived_tex.resolve()),
        "copied_refs": copied_refs,
    }
    manifest_path = workspace / "tailoring-manifest.json"
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")

    print(json.dumps(manifest, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
