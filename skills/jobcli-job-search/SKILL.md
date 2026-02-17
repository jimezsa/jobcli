---
name: jobcli-job-search
description: Search unseen jobs and keep only strict YES/HIGH persona matches.
homepage: https://github.com/jimezsa/jobcli
metadata:
  {
    "openclaw":
      {
        "emoji": "💼",
        "os": ["linux", "darwin"],
        "requires": { "bins": ["jobcli", "python3"] },
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

# JobCLI Job Search (Minimal)

Goal: retrieve unseen jobs and keep only jobs classified as `YES` + `HIGH`.

## Prerequisites

- `profiles/<user_id>/CVSUMMARY.md`
- optional: `profiles/<user_id>/persona_profile.json`
- script: `skills/jobcli-job-search/scripts/job_discriminator.py`
- API key: `MINIMAX_API_KEY` (fallbacks supported by script)

## Required User-Scoped Files

- `profiles/<user_id>/CVSUMMARY.md`
- `profiles/<user_id>/jobs_seen.json`
- `profiles/<user_id>/jobs_new_keyword_<n>.json` (temp)
- `profiles/<user_id>/jobs_new_all.json` (temp)
- `profiles/<user_id>/jobs_filtered_out.json`
- `profiles/<user_id>/jobs_yes_high.json`

Never mix files across users.

## Minimal Flow

1. Load `CVSUMMARY.md` (required). If missing, run `jobcli-cv-summary` first.
2. Build search queries from persona role targets.
3. Retrieve unseen jobs per query with seen tracking:

```bash
jobcli search "<query>" --location "<location>" --country "<code>" --limit 20 \
  --seen profiles/<user_id>/jobs_seen.json --new-only --seen-update \
  --json --output profiles/<user_id>/jobs_new_keyword_<n>.json --hours 48
```

4. Aggregate query outputs into `profiles/<user_id>/jobs_new_all.json` using `jobcli` merge/dedupe:

```bash
for f in profiles/<user_id>/jobs_new_keyword_*.json; do
  jobcli seen update \
    --seen profiles/<user_id>/jobs_new_all.json \
    --input "$f" \
    --out profiles/<user_id>/jobs_new_all.json
done
```
5. Apply deterministic hard rejects (role/domain mismatch, seniority mismatch, work-mode/location hard mismatch), write rejects to `profiles/<user_id>/jobs_filtered_out.json`, then run the LLM gate on remaining jobs:

```bash
python3 skills/jobcli-job-search/scripts/job_discriminator.py \
  --cvsummary profiles/<user_id>/CVSUMMARY.md \
  --jobs-json profiles/<user_id>/jobs_new_all.json \
  --output profiles/<user_id>/jobs_yes_high.json
```

6. Return only jobs from `jobs_yes_high.json`.
7. Remove temp files (`jobs_new_keyword_*.json`, `jobs_new_all.json`).

## Non-Negotiable Rules

1. No ranking and no score-based ordering.
2. One LLM context per job.
3. Keep only `decision=YES` and `confidence=HIGH`.
4. If uncertain, return `NO`.
5. Keep `--seen-update` enabled so processed jobs do not repeat.

## Output Format

send only accepted jobs to <user_id> as an enumerated list:

```text
1. 💼 Role: job_title
   🏢 Company: company
   📍 Location: location
   🔗 Apply: url
2. 💼 Role: job_title
   🏢 Company: company
   📍 Location: location
   🔗 Apply: url
```

If no accepted jobs: report count `0`.

## Validation

Per user, confirm:

1. `jobs_yes_high.json` exists.
2. `jobs_yes_high.json` contains only accepted jobs.
3. `jobs_seen.json` was updated.
