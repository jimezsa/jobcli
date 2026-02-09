---
name: jobcli-job-search
description: Search and rank job listings against a persona using the JobCLI tool.
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
              "label": "Build JobCLI from source (requires Go 1.25)",
            },
            {
              "id": "release",
              "kind": "shell",
              "script": "curl -sL https://github.com/jimezsa/jobcli/releases/latest/download/jobcli_$(uname -s)_$(uname -m).tar.gz | tar xz && sudo mv jobcli /usr/local/bin/",
              "bins": ["jobcli"],
              "label": "Install JobCLI from GitHub release",
            },
          ],
      },
  }
---

# JobCLI — Job Search + Persona Ranking

Daily/on-demand flow: read `CVSUMMARY.md`, search jobs with `jobcli`, dedupe,
score against persona, and return ranked results.

> Prerequisite: `CVSUMMARY.md` exists in cwd.
> Trigger: user asks for job search/ranking.

## 1) Load persona inputs

Read `CVSUMMARY.md` and extract:

- `## Persona Summary`
- all keywords under `## Search Keywords`
- `## Ranking Criteria`

If missing, stop and ask user to run CV summary skill first.

## 2) Prepare JobCLI

Run once if needed:

```bash
jobcli config init
```

Ask for `location` and `country` if unknown.

## 3) Search per keyword (sequential)

For each keyword:

```bash
jobcli search "<keyword>" --location "<location>" --country "<code>" --limit 30 --json --output jobs_keyword_<n>.json
```

Rules:

- run sequentially (avoid anti-bot triggers)
- if 403/429 or empty, retry once with `--sites linkedin,google`
- if still failing, skip and note failure

## 4) Aggregate + dedupe

Merge `jobs_keyword_*.json`, dedupe by full URL, keep matched keyword list per
job.

## 5) Score each job (0.0–1.0)

Use equal-weight dimensions (0.2 each):

- title match
- skill overlap
- domain fit
- seniority alignment
- language fit

Final score = average of 5 dimensions.

## 6) Present ranked output

Sort descending by score.

- default: show only score `>= 0.7`
- user may request all jobs or a custom threshold
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
Offer saving to `ranked_jobs.md`.
After results, send one short funny motivational line.

## 7) Cleanup

Delete `jobs_keyword_*.json`. Keep `CVSUMMARY.md` and `ranked_jobs.md` (if
created).

## Notes

- privacy first: never expose personal data from `CVSUMMARY.md`
- if many rate-limit failures, suggest lower `--limit 10` or proxies
- for daily refreshes, consider `--hours 48`
