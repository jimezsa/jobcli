#!/usr/bin/env python3
"""Filter jobs with an LLM using CVSUMMARY.md as persona input."""

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any, Dict, List, Tuple


DEFAULT_API_KEY = os.environ.get(
    "MINIMAX_API_KEY",
    os.environ.get("ANTHROPIC_API_KEY", os.environ.get("OPENAI_API_KEY", "")),
)
DEFAULT_MODEL = "MiniMax-M2.5"
DEFAULT_API_URL = os.environ.get("ANTHROPIC_BASE_URL", "https://api.minimax.io/anthropic").rstrip("/") + "/v1/messages"

TITLE_KEYS = ("title", "job_title", "position", "role")
DESCRIPTION_KEYS = ("description", "snippet", "summary", "details")
LOCATION_KEYS = ("location", "city", "region", "country")
ID_KEYS = ("id", "job_id", "url", "link")
COMPANY_KEYS = ("company", "company_name", "employer")


def read_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def looks_like_job(obj: Any) -> bool:
    if not isinstance(obj, dict):
        return False
    has_title = any(isinstance(obj.get(k), str) and obj.get(k).strip() for k in TITLE_KEYS)
    has_signal = any(isinstance(obj.get(k), str) and obj.get(k).strip() for k in (DESCRIPTION_KEYS + LOCATION_KEYS + ID_KEYS + COMPANY_KEYS))
    return has_title and has_signal


def collect_jobs_recursive(node: Any) -> List[Dict[str, Any]]:
    jobs: List[Dict[str, Any]] = []
    if isinstance(node, list):
        for item in node:
            jobs.extend(collect_jobs_recursive(item))
        return jobs
    if isinstance(node, dict):
        if looks_like_job(node):
            return [node]
        for value in node.values():
            jobs.extend(collect_jobs_recursive(value))
    return jobs


def first_str(job: Dict[str, Any], keys: Tuple[str, ...]) -> str:
    for key in keys:
        value = job.get(key)
        if isinstance(value, str) and value.strip():
            return value.strip()
    return ""


def compact_job(job: Dict[str, Any]) -> Dict[str, Any]:
    description = first_str(job, DESCRIPTION_KEYS)
    if len(description) > 2000:
        description = description[:2000] + "..."
    return {
        "id": first_str(job, ID_KEYS),
        "title": first_str(job, TITLE_KEYS),
        "company": first_str(job, COMPANY_KEYS),
        "location": first_str(job, LOCATION_KEYS),
        "description": description,
    }


def parse_decision(content: str) -> Dict[str, str]:
    content = content.strip()
    try:
        data = json.loads(content)
    except json.JSONDecodeError:
        start = content.find("{")
        end = content.rfind("}")
        if start == -1 or end == -1 or end <= start:
            return {"decision": "NO", "confidence": "LOW", "reason": "invalid_json_response"}
        try:
            data = json.loads(content[start : end + 1])
        except json.JSONDecodeError:
            return {"decision": "NO", "confidence": "LOW", "reason": "invalid_json_response"}

    decision = str(data.get("decision", "NO")).strip().upper()
    confidence = str(data.get("confidence", "LOW")).strip().upper()
    reason = str(data.get("reason", "")).strip()

    if decision not in {"YES", "NO"}:
        decision = "NO"
    if confidence not in {"HIGH", "LOW"}:
        confidence = "LOW"
    if not reason:
        reason = "no_reason_provided"

    return {"decision": decision, "confidence": confidence, "reason": reason}


def llm_compare(cvsummary: str, job: Dict[str, Any], api_key: str, model: str, api_url: str, timeout: int) -> Dict[str, str]:
    system_prompt = (
        "You are a strict job filter. Compare one candidate CV summary and one job. "
        "Return JSON only with keys: decision, confidence, reason. "
        "Rules: use YES only when clearly aligned. If unsure return NO. "
        "Use HIGH only when very certain; else LOW. Focus on title/domain first, then description."
    )
    user_prompt = (
        "CVSUMMARY.md:\n"
        f"{cvsummary}\n\n"
        "JOB:\n"
        f"{json.dumps(compact_job(job), ensure_ascii=True)}\n\n"
        'Return exactly: {"decision":"YES|NO","confidence":"HIGH|LOW","reason":"short reason"}'
    )

    payload = {
        "model": model,
        "max_tokens": 256,
        "temperature": 0.1,
        "system": system_prompt,
        "messages": [{"role": "user", "content": [{"type": "text", "text": user_prompt}]}],
    }
    headers = {
        "x-api-key": api_key,
        "anthropic-version": "2023-06-01",
        "Content-Type": "application/json",
    }

    req = urllib.request.Request(
        api_url,
        data=json.dumps(payload).encode("utf-8"),
        headers=headers,
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=timeout) as response:
        response_data = json.loads(response.read().decode("utf-8"))

    # Anthropic-compatible response shape.
    content_blocks = response_data.get("content")
    if isinstance(content_blocks, list):
        text_parts: List[str] = []
        for block in content_blocks:
            if isinstance(block, dict) and block.get("type") == "text":
                text = block.get("text", "")
                if isinstance(text, str) and text.strip():
                    text_parts.append(text.strip())
        return parse_decision("\n".join(text_parts))

    # Fallback for standard MiniMax/OpenAI-compatible response shape.
    base_resp = response_data.get("base_resp")
    if isinstance(base_resp, dict):
        status_code = int(base_resp.get("status_code", 0))
        if status_code != 0:
            status_msg = str(base_resp.get("status_msg", "unknown_error"))
            raise RuntimeError(f"MiniMax API error {status_code}: {status_msg}")

    content = response_data.get("choices", [{}])[0].get("message", {}).get("content", "")
    if isinstance(content, list):
        text_parts: List[str] = []
        for item in content:
            if isinstance(item, dict):
                text = item.get("text", "")
                if isinstance(text, str) and text.strip():
                    text_parts.append(text.strip())
        content = "\n".join(text_parts)
    return parse_decision(content)


def main() -> int:
    parser = argparse.ArgumentParser(description="Filter jobs with LLM using CVSUMMARY.md.")
    parser.add_argument("--cvsummary", required=True, help="Path to CVSUMMARY.md")
    parser.add_argument("--jobs-json", required=True, help="Path to jobs JSON (list or nested structure)")
    parser.add_argument("--output", required=True, help="Output path for YES jobs JSON list")
    parser.add_argument("--api-key", default=DEFAULT_API_KEY, help="LLM API key (default: MINIMAX_API_KEY, fallback: ANTHROPIC_API_KEY/OPENAI_API_KEY)")
    parser.add_argument("--model", default=DEFAULT_MODEL, help=f"Model name (default: {DEFAULT_MODEL})")
    parser.add_argument("--api-url", default=DEFAULT_API_URL, help=f"API URL (default: {DEFAULT_API_URL})")
    parser.add_argument("--timeout", type=int, default=60, help="HTTP timeout seconds")
    parser.add_argument("--max-jobs", type=int, default=0, help="Optional limit for processed jobs (0 = all)")
    args = parser.parse_args()

    if not args.api_key:
        print("Error: MINIMAX_API_KEY/ANTHROPIC_API_KEY/OPENAI_API_KEY not set and --api-key not provided", file=sys.stderr)
        return 1

    try:
        cvsummary_text = Path(args.cvsummary).read_text(encoding="utf-8")
        jobs_data = read_json(Path(args.jobs_json))
        jobs = collect_jobs_recursive(jobs_data)
    except Exception as exc:  # noqa: BLE001
        print(f"Error reading inputs: {exc}", file=sys.stderr)
        return 1

    if args.max_jobs > 0:
        jobs = jobs[: args.max_jobs]

    yes_jobs: List[Dict[str, Any]] = []
    processed = 0

    for job in jobs:
        processed += 1
        try:
            result = llm_compare(
                cvsummary=cvsummary_text,
                job=job,
                api_key=args.api_key,
                model=args.model,
                api_url=args.api_url,
                timeout=args.timeout,
            )
        except urllib.error.URLError as exc:
            print(f"Network/API error on job {processed}: {exc}", file=sys.stderr)
            continue
        except Exception as exc:  # noqa: BLE001
            print(f"Evaluation error on job {processed}: {exc}", file=sys.stderr)
            continue

        if result["decision"] == "YES" and result["confidence"] == "HIGH":
            yes_jobs.append(job)

    Path(args.output).write_text(json.dumps(yes_jobs, ensure_ascii=True, indent=2), encoding="utf-8")
    print(f"Processed: {processed}")
    print(f"YES_HIGH: {len(yes_jobs)}")
    print(f"Output: {args.output}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
