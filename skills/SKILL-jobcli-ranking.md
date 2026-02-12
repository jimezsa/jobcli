---
name: jobcli-job-search
description: Search and rank unseen jobs with JobCLI using per-user CVSUMMARY personas.
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

Goal: rank only unseen jobs per user. Persist only each user's `jobs_seen.json`
between runs.

Prerequisite: `profiles/<user_id>/CVSUMMARY.md` exists.
Trigger: user asks for job search/ranking (single user or multiple users).

## 0) Multi-User Mode and Isolation (Required)

Inputs:

- single-user mode: one `user_id`
- batch mode: list of `user_id` values

For each user, all files must be user-scoped:

- summary: `profiles/<user_id>/CVSUMMARY.md`
- seen state (persistent): `profiles/<user_id>/jobs_seen.json`
- per-keyword temp files: `profiles/<user_id>/jobs_new_keyword_<n>.json`
- run aggregate (temporary): `profiles/<user_id>/jobs_new_all.json`
- optional ranked output: `profiles/<user_id>/ranked_jobs.md`

Never share state files between users.

## 1) Persona input

Read from `profiles/<user_id>/CVSUMMARY.md`:

- `## Persona Summary`
- `## Search Keywords`
- `## Ranking Criteria`
- `## User Context` (default location/country if present)

If missing, stop and ask user to run the CV summary skill first.

Determine `location` and `country` for that user:

- first choice: values from `## User Context`
- fallback: ask the user

Use `profiles/<user_id>/jobs_seen.json` as seen history (only persistent state file for that user).

## 2) Search per keyword (sequential)

For each keyword:

```bash
jobcli search "<keyword>" --location "<location>" --country "<code>" --limit 30 \
  --seen profiles/<user_id>/jobs_seen.json --new-only --seen-update \
  --json --output profiles/<user_id>/jobs_new_keyword_<n>.json --hours 72
```

Rules:

- run sequentially
- on 403/429 or empty result: skip

`--seen-update` marks discovered unseen jobs as seen in `jobs_seen.json`.

## 3) Aggregate + dedupe (single command type)

For each `jobs_new_keyword_<n>.json` run:

```bash
jobcli seen update --seen profiles/<user_id>/jobs_new_all.json \
  --input profiles/<user_id>/jobs_new_keyword_<n>.json \
  --out profiles/<user_id>/jobs_new_all.json --stats
```

Notes:

- missing `profiles/<user_id>/jobs_new_all.json` is treated as empty on first run
- repeated URLs are deduped by JobCLI seen merge logic
- `profiles/<user_id>/jobs_new_all.json` is temporary for this run only

If `profiles/<user_id>/jobs_new_all.json` is empty, report "no new jobs found"
for that user and stop ranking for that user.

## 4) Score (0.0-1.0)

Equal weights (0.2 each):

- title match
- skill overlap
- domain fit
- seniority alignment
- language fit

Final score = average of the 5 dimensions.

## 5) Output

Sort by score descending.

- default threshold: `>= 0.7`
- allow custom threshold or all jobs
- send one job per message

Format:

```text
[user_id]
🥇 job_title_here
📍 Location
🏢 Company_name
⭐ Score: 0.00

full_url_link_here
```

Use `🥈` for rank 2, `🥉` for rank 3, and `N.` for rank 4+.
Do not persist ranked output unless user explicitly asks to keep
`profiles/<user_id>/ranked_jobs.md`.
After results, send one short funny motivational line.

Batch mode behavior:

- run users sequentially to reduce rate-limit risk
- complete all steps for one `user_id` before starting next `user_id`
- return grouped output blocks, one block per `user_id`

## 6) Cleanup

Delete:

- `profiles/<user_id>/jobs_new_keyword_*.json`
- `profiles/<user_id>/jobs_new_all.json`
- `profiles/<user_id>/ranked_jobs.md` (unless user asked to keep it)

Keep:

- `profiles/<user_id>/CVSUMMARY.md`
- `profiles/<user_id>/jobs_seen.json`

## Notes

- privacy first: never expose personal data from `profiles/<user_id>/CVSUMMARY.md`
- if rate-limited often, suggest lower `--limit 10` or proxies
- isolation check: if a command references a different `user_id` path than the
  active user, treat it as a bug and fix before execution
