---
name: jobcli-job-search
description: Search unseen jobs and filter them with a strict per-job YES/NO persona gate.
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

# JobCLI Job Search + Binary Persona Filter (v3)

Goal: eliminate ranking and keep only high-confidence persona matches (`YES/HIGH`) using an LLM gate.

Prerequisites:

- `profiles/<user_id>/CVSUMMARY.md`
- script: `skills/jobcli-job-search/scripts/job_discriminator.py`
- `OPENAI_API_KEY` set in environment (or pass `--api-key`)

Trigger: user asks for job search and filtering (no ranking).

## 0) Multi-User Isolation (Required)

Inputs:

- single-user mode: one `user_id`
- batch mode: list of `user_id`

User-scoped files only:

- CV summary: `profiles/<user_id>/CVSUMMARY.md`
- persona JSON (optional): `profiles/<user_id>/persona_profile.json`
- seen state: `profiles/<user_id>/jobs_seen.json` (persistent)
- per-query temp: `profiles/<user_id>/jobs_new_keyword_<n>.json`
- aggregate temp: `profiles/<user_id>/jobs_new_all.json`
- hard rejects: `profiles/<user_id>/jobs_filtered_out.json`
- subagent keepers: `profiles/<user_id>/jobs_yes_high.json`

Never cross user paths.

## 1) Persona Input Priority

Load in this order:

1. `profiles/<user_id>/CVSUMMARY.md` (required)
2. `profiles/<user_id>/persona_profile.json` (optional helper)

If `CVSUMMARY.md` is missing, stop and ask to run `jobcli-cv-summary` first.

Resolve location/country:

1. values from persona files
2. fallback to explicit user input

## 2) Query Plan + Retrieval

Use persona role titles to build up to 12 queries:

1. core role titles: 4-5
2. skill-intent titles: 3-4
3. domain/seniority variants: 3-4

Run sequentially per query:

```bash
jobcli search "<query>" --location "<location>" --country "<code>" --limit 30 \
  --seen profiles/<user_id>/jobs_seen.json --new-only --seen-update \
  --json --output profiles/<user_id>/jobs_new_keyword_<n>.json --hours 72
```

Rules:

1. keep `--seen-update` enabled
2. continue on per-query errors or zero results
3. if very low volume, retry remaining queries with `--hours 96`

## 3) Aggregate + Normalize

Aggregate query files into:

- `profiles/<user_id>/jobs_new_all.json`

Normalize:

1. canonicalize URL
2. trim/lower comparable fields
3. dedupe by URL, fallback `(title, company, location)`

If empty, report no new jobs and stop for that user.

## 4) Deterministic Hard Reject (Before Subagent)

Reject obvious mismatches before running the per-job gate:

1. excluded role/domain keyword in title
2. severe seniority mismatch in title
3. work-mode hard mismatch (for example remote-only persona vs onsite-only job)
4. explicit location mismatch

Persist hard rejects:

- `profiles/<user_id>/jobs_filtered_out.json`

## 5) LLM Gate (Recursive Jobs JSON)

Evaluate all jobs from the aggregate JSON recursively (one LLM context per job):

```bash
python3 skills/jobcli-job-search/scripts/job_discriminator.py \
  --cvsummary profiles/<user_id>/CVSUMMARY.md \
  --jobs-json profiles/<user_id>/jobs_new_all.json \
  --output profiles/<user_id>/jobs_yes_high.json
```

Script behavior:

1. recursively finds job objects in JSON
2. calls LLM once per job (`CVSUMMARY.md` + job context)
3. keeps only `YES` + `HIGH` decisions
4. writes a JSON list of original job objects to `--output`

Rules:

1. one context per job (no multi-job prompt)
2. compare title/domain first, then description
3. if unsure, return `NO`
4. if uncertain, return `NO` (do not allow ambiguous yes)

## 6) Final Decision Policy (No Ranking)

Only keep:

- `decision = YES`
- `confidence = HIGH`

Everything else goes to reject bucket.

Persist:

- `profiles/<user_id>/jobs_yes_high.json`

Output order is retrieval order or newest-first only.  
Do not compute or display score/rank.

## 7) Output to User

Show only `YES/HIGH` jobs, one item per message:

```text
[user_id]
job_title_here
Location
Company_name
full_url_link_here
```

If no `YES/HIGH` jobs, report count = 0.

## 8) Seen-State and Cleanup

Because retrieval uses `--seen-update`, both accepted and rejected jobs are already tracked and should not reappear next run.

Delete temporary artifacts:

- `profiles/<user_id>/jobs_new_keyword_*.json`
- `profiles/<user_id>/jobs_new_all.json`

Keep persistent artifacts:

- `profiles/<user_id>/CVSUMMARY.md`
- `profiles/<user_id>/persona_profile.json` (if present)
- `profiles/<user_id>/jobs_seen.json`
- `profiles/<user_id>/jobs_filtered_out.json`
- `profiles/<user_id>/jobs_yes_high.json`

## Notes

- never reintroduce weighted ranking in this skill
- keep persona hard exclusions explicit to prevent cross-domain false positives
- optional motivation skill can run after filtering only
