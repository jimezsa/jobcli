---
name: jobcli-job-search
description: Search and rank unseen jobs with JobCLI using CVSUMMARY persona.
homepage: https://github.com/jimezsa/jobcli
metadata:
  {
    "openclaw":
      {
        "emoji": "💼",
        "os": ["linux", "darwin"],
        "requires": { "bins": ["jobcli"] },
        "install":
          [
            {
              "id": "source",
              "kind": "shell",
              "script": "git clone https://github.com/jimezsa/jobcli && cd jobcli && make && sudo mv jobcli /usr/local/bin/",
              "bins": ["jobcli"],
              "label": "Build JobCLI from source (Go 1.25)",
            },
            {
              "id": "release",
              "kind": "shell",
              "script": "curl -sL https://github.com/jimezsa/jobcli/releases/latest/download/jobcli_$(uname -s)_$(uname -m).tar.gz | tar xz && sudo mv jobcli /usr/local/bin/",
              "bins": ["jobcli"],
              "label": "Install JobCLI release",
            },
          ],
      },
  }
---

# JobCLI Ranking (Compact)

Goal: rank only unseen jobs. Persist only `jobs_seen.json` between runs.

Prerequisite: `CVSUMMARY.md` exists in cwd.
Trigger: user asks for job search/ranking.

## 1) Persona input

Read from `CVSUMMARY.md`:

- `## Persona Summary`
- `## Search Keywords`
- `## Ranking Criteria`

If missing, stop and ask user to run the CV summary skill first.

## 2) Setup

Run once if needed:

```bash
jobcli config init
```

Collect `location` and `country`. Use `jobs_seen.json` as seen history (only persistent state file).

## 3) Search per keyword (sequential)

For each keyword:

```bash
jobcli search "<keyword>" --location "<location>" --country "<code>" --limit 30 \
  --seen jobs_seen.json --new-only --seen-update \
  --json --output jobs_new_keyword_<n>.json
```

Rules:

- run sequentially
- on 403/429 or empty result: skip

`--seen-update` marks discovered unseen jobs as seen in `jobs_seen.json`.

## 4) Aggregate + dedupe (single command type)

For each `jobs_new_keyword_<n>.json` run:

```bash
jobcli seen update --seen jobs_new_all.json --input jobs_new_keyword_<n>.json \
  --out jobs_new_all.json --stats
```

Notes:

- missing `jobs_new_all.json` is treated as empty on first run
- repeated URLs are deduped by JobCLI seen merge logic
- `jobs_new_all.json` is temporary for this run only

If `jobs_new_all.json` is empty, report "no new jobs found" and stop (no ranking).

## 5) Score (0.0-1.0)

Equal weights (0.2 each):

- title match
- skill overlap
- domain fit
- seniority alignment
- language fit

Final score = average of the 5 dimensions.

## 6) Output

Sort by score descending.

- default threshold: `>= 0.7`
- allow custom threshold or all jobs
- send one job per message

Format:

```text
🥇 job_title_here
📍 Location
🏢 Company_name
⭐ Score: 0.00

full_url_link_here
```

Use `🥈` for rank 2, `🥉` for rank 3, and `N.` for rank 4+.
Do not persist ranked output unless user explicitly asks to keep `ranked_jobs.md`.
After results, send one short funny motivational line.

## 7) Cleanup

Delete:

- `jobs_new_keyword_*.json`
- `jobs_new_all.json`
- `ranked_jobs.md` (unless user asked to keep it)

Keep:

- `CVSUMMARY.md`
- `jobs_seen.json`

## Notes

- privacy first: never expose personal data from `CVSUMMARY.md`
- if rate-limited often, suggest lower `--limit 10` or proxies
- for daily refreshes, consider `--hours 48`
