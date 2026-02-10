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

This skill only ranks *new/unseen* jobs, and updates the seen-history JSON
after searching so future runs don’t re-rank the same listings.

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

Decide on a seen-history path (default: `jobs_seen.json` in cwd).

## 3) Search per keyword (sequential, new-only)

For each keyword:

```bash
jobcli search "<keyword>" --location "<location>" --country "<code>" --limit 30 \
  --seen jobs_seen.json --new-only \
  --json --output jobs_new_keyword_<n>.json
```

Rules:

- run sequentially (avoid anti-bot triggers)
- if 403/429 or empty, retry once with `--sites linkedin,google`
- if still failing, skip and note failure

## 4) Aggregate + dedupe (new jobs only)

Merge `jobs_new_keyword_*.json` into `jobs_new_all.json`, dedupe by full URL,
keep matched keyword list per job.

If `jobs_new_all.json` is empty, report “no new jobs found” and stop (do not
rank).

## 5) Update seen history (after search)

After aggregating/deduping, update the seen-history JSON so the new jobs won’t
show up again on the next run:

```bash
jobcli seen update --seen jobs_seen.json --input jobs_new_all.json --out jobs_seen.json --stats
```

Notes:

- `--seen` missing file is treated as empty; `--out` writes/updates it.
- This marks all discovered new jobs as “seen” regardless of whether the user
  applies to them.

## 6) Score each job (0.0–1.0)

Use equal-weight dimensions (0.2 each):

- title match
- skill overlap
- domain fit
- seniority alignment
- language fit

Final score = average of 5 dimensions.

## 7) Present ranked output

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

## 8) Cleanup

Delete `jobs_new_keyword_*.json`. Keep `jobs_new_all.json`, `CVSUMMARY.md`, and
`ranked_jobs.md` (if created).

## Notes

- privacy first: never expose personal data from `CVSUMMARY.md`
- if many rate-limit failures, suggest lower `--limit 10` or proxies
- for daily refreshes, consider `--hours 48`
