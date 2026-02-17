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

## 7) Two-Stage Scoring (0.0-1.0)

### Stage 1: Title Filter (Fast)
Score keywords against job TITLE only. This is the primary filter.

- Match each persona keyword against job title (case-insensitive)
- **If ANY keyword matches title → title_score = 1.0 (pass)**
- **If NO keyword matches title → title_score = 0.0 (reject)**
- Keep candidates with title_score = 1.0 for Stage 2

### Stage 2: Description + Domain Score (Selective)
For Stage 1 candidates (title matched), score against full job description AND check domain.

**Weighing:**
1. title match: `0.40` (binary: 0 or 1)
2. description relevance: `0.30`
3. domain fit: `0.30`

**Description scoring:**
- Check if description contains persona skills/keywords
- description_score = 1.0 if good match, 0.5 if partial, 0.0 if none

**Domain scoring:**
- Load `preferred_domains` and `excluded_domains` from persona
- If job title/description contains ANY excluded_domain → domain_score = 0.0 (strong penalty)
- If job contains preferred_domain → domain_score += 0.5
- domain_score = clamp(0.0 to 1.0)

**Penalties:**
- excluded domain match: `-0.50` (major penalty)
- missing description: `-0.10`
- sparse posting: `-0.05`

**Threshold: >= 0.80** (only jobs meeting this are sent to user)

Formula:
`final_score = clamp((0.40 * title_score) + (0.30 * desc_score) + (0.30 * domain_score) - penalties, 0.0, 1.0)`

For each job capture:
- `stage1_title_matches` (true/false)
- `title_score` (1.0 or 0.0)
- `description_score`
- `domain_score`
- `domain_penalty_applied`
- `final_score`
- `matched_terms`

## 8) Output

Sort by score descending.

- **threshold: `>= 0.80`** (only jobs meeting this pass)
- If no jobs pass threshold, report count = 0
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

Persist outputs only if user requests:

- `profiles/<user_id>/ranked_jobs.md`
- `profiles/<user_id>/ranked_jobs.json`

## 9) Cleanup

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
