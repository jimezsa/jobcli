---
name: jobcli-job-search
description: Search and rank unseen jobs with JobCLI using per-user persona profiles and token-efficient scoring.
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

# JobCLI Job Search + Ranking (v2)

Goal: rank unseen jobs with high precision while keeping run-time and token use bounded.

Prerequisites:

- `profiles/<user_id>/persona_profile.json`

Trigger: user asks for job search/ranking for one or more users.

## 0) Multi-User Isolation (Required)

Inputs:

- single-user mode: one `user_id`
- batch mode: list of `user_id`

User-scoped files only:

- persona JSON: `profiles/<user_id>/persona_profile.json`
- seen state: `profiles/<user_id>/jobs_seen.json` (persistent)
- per-query temp: `profiles/<user_id>/jobs_new_keyword_<n>.json`
- aggregate temp: `profiles/<user_id>/jobs_new_all.json`
- optional filtered-out: `profiles/<user_id>/jobs_filtered_out.json`
- optional ranked: `profiles/<user_id>/ranked_jobs.md`, `profiles/<user_id>/ranked_jobs.json`
- optional feedback: `profiles/<user_id>/ranking_feedback.json`
- optional recheck: `profiles/<user_id>/jobs_recheck.json`

Never cross user paths.

## 1) Persona Input Priority

Load persona from `profiles/<user_id>/persona_profile.json` only.
If missing required persona data, stop and ask to run `jobcli-cv-summary` first.

Resolve location/country:

1. use values in persona files
2. fallback to explicit user input

## 2) Token-Efficiency Guardrails (Mandatory)

1. Parse persona once per user and cache normalized fields.
2. Query budget: max 12 queries per user per run.
3. Retrieval budget: max 30 jobs/query.
4. Deep-scoring budget: max 120 deduped jobs.
5. Evidence text budget per job:
   - always score against `title` + `description` when description exists
   - include description digest capped to 600 chars
   - if description is missing, score against `title` only
6. Output budget:
   - default top 15 ranked jobs
   - reasons limited to 3 match + 2 mismatch, each <= 12 words

## 3) Query Plan (Per User)

Build up to 12 queries from persona/keyword bank:

1. core role titles: 4-5
2. skill-intent titles: 3-4
3. domain/seniority variants: 3-4

De-duplicate semantically similar queries.
Use only realistic job-position titles (no skill-only/tool-only queries).
Write query list to:

- `profiles/<user_id>/queries_v2.json`

## 4) Retrieval (Sequential, Seen Updated Immediately)

For each query run:

```bash
jobcli search "<query>" --location "<location>" --country "<code>" --limit 30 \
  --seen profiles/<user_id>/jobs_seen.json --new-only --seen-update \
  --json --output profiles/<user_id>/jobs_new_keyword_<n>.json --hours 72
```

If total new jobs are very low, rerun remaining queries with `--hours 96`.

Rules:

- run sequentially (lower rate-limit risk)
- on 403/429 or empty output: skip and continue
- `--seen-update` stays enabled so low-fit jobs are not re-ranked next run

## 5) Aggregate + Normalize

For each per-query file:

```bash
jobcli seen update --seen profiles/<user_id>/jobs_new_all.json \
  --input profiles/<user_id>/jobs_new_keyword_<n>.json \
  --out profiles/<user_id>/jobs_new_all.json --stats
```

Then normalize:

1. canonicalize URL
2. normalize comparable text (trim/lowercase for matching keys)
3. dedupe by URL, fallback `(title, company, location)`
4. lightweight lexical pre-rank and keep top 120 for deep scoring

If aggregate is empty, report "no new jobs found" and stop for that user.

## 6) Hard Constraint Filter (Pre-Rank)

Reject jobs that violate non-negotiables:

1. severe seniority mismatch
2. work-mode mismatch (for example onsite-only vs remote-only persona)
3. location outside allowed constraints
4. language mismatch when explicit language requirement exists
5. excluded area/role family

Store rejects with reasons in:

- `profiles/<user_id>/jobs_filtered_out.json`

## 7) Weighted Scoring (0.0-1.0)

Score surviving jobs with explicit breakdown. Ranking must compare persona against
both job title and job description content.

Weights:

1. title-role fit: `0.25`
2. description-persona fit: `0.25`
3. must-have skill coverage (title + description + snippet): `0.20`
4. preferred skill coverage: `0.10`
5. seniority alignment: `0.10`
6. work mode + location fit: `0.05`
7. freshness signal: `0.05`

Penalties:

- missing must-have cluster: `-0.15`
- strong over/under qualification mismatch: `-0.10`
- sparse or ambiguous posting data: `-0.05`
- missing description (title-only fallback): `-0.05`

Formula:

`final_score = clamp(sum(weight_i * subscore_i) - penalties, 0.0, 1.0)`

For each job capture:

- `score_breakdown`
- `matched_terms`
- `missing_must_haves`
- `penalties_applied`
- `confidence` (`high|medium|low`)
- `scoring_text_used` (title + description digest, or title-only fallback)

## 8) Ranking Bands and Output

Bands:

1. `>= 0.80`: strong match
2. `0.8 - 0.79`: manual review
3. `< 0.8`: low fit (hide by default)

Sort by score descending and output top 15 by default.

Output format per job:

- rank
- title
- company
- location
- final score
- top 3 match reasons
- top 2 mismatch reasons
- URL

Persist outputs only if user requests:

- `profiles/<user_id>/ranked_jobs.md`
- `profiles/<user_id>/ranked_jobs.json`

## 9) Feedback Loop (Optional but Recommended)

Capture user labels:

- `applied`
- `interview`
- `rejected`
- `not_relevant`

Persist labels to:

- `profiles/<user_id>/ranking_feedback.json`

Weekly tuning:

1. increase weights correlated with `applied/interview`
2. increase penalties on repeated `not_relevant` patterns
3. refresh high-performing query variants

## 10) Cleanup

Delete temporary artifacts:

- `profiles/<user_id>/jobs_new_keyword_*.json`
- `profiles/<user_id>/jobs_new_all.json`

Keep persistent artifacts:

- `profiles/<user_id>/persona_profile.json`
- `profiles/<user_id>/jobs_seen.json`
- `profiles/<user_id>/ranking_feedback.json` (if used)
- `profiles/<user_id>/jobs_recheck.json` (if used)

In batch mode, process users sequentially end-to-end.

## Notes

- privacy first: never expose personal data
- if rate-limited often, suggest lower query count, lower `--limit`, or proxies
- keep outputs concise; avoid verbose scoring narratives
- if using `skills/pirate-motivator/SKILL.md`, use it only after ranking is complete
