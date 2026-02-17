# Persona Binary Filter Plan (v3)

## Goal

Replace ranking with a simple per-job YES/NO filter based on persona summary, with high-confidence decisions only.

## Why Change

Current ranking still lets wrong-domain jobs pass (example: mechanical engineer receiving software roles).  
New design removes score tuning and keeps only jobs that clearly match persona intent.

## Scope

- `skills/spec-skills.md` (this plan)
- `skills/jobcli-cv-summary/SKILL.md` (persona summary contract)
- `skills/jobcli-job-search/SKILL.md` (filter-only flow)
- new script in `skills/jobcli-job-search/scripts/` similar usage style to:
  `skills/pirate-motivator/scripts/generate_motivation.py`

## New Principles

1. No ranking, no weighted scores, no thresholds like 0.65/0.80.
2. One isolated context per job evaluation.
3. Binary output only: `YES` or `NO`.
4. Keep only `YES` decisions with `HIGH` confidence.
5. If confidence is not high, default to `NO` (fail closed).

## Input Contract (Persona Summary First)

Primary persona file:

- `profiles/<user_id>/persona_summary.md`

Optional structured companion:

- `profiles/<user_id>/persona_profile.json`

`persona_summary.md` must include:

1. target roles (allowed role families)
2. excluded roles/domains (explicit negatives)
3. seniority target
4. location/work-mode constraints
5. must-have skills (short list)

## End-to-End Flow (Filter-Only)

## Stage 1: Retrieve Unseen Jobs

Keep existing `jobcli search` collection flow with seen tracking.  
Collect candidates only; do not rank.

## Stage 2: Deterministic Pre-Filter (Hard Reject)

Before any subagent call, reject obvious mismatches:

1. role-family/domain mismatch from title
2. excluded domain keyword hit
3. location/work-mode hard mismatch
4. severe seniority mismatch when explicit in title

Output rejects to:

- `profiles/<user_id>/jobs_filtered_out.json`

## Stage 3: Subagent Binary Decision (One Context Per Job)

For each surviving job, call a Python script with one job at a time.

Proposed script path:

- `skills/jobcli-job-search/scripts/job_discriminator.py`

Invocation style (example):

```bash
python3 skills/jobcli-job-search/scripts/job_discriminator.py \
  --persona profiles/<user_id>/persona_summary.md \
  --job-json /tmp/job_<id>.json
```

Required script output (JSON):

```json
{
  "job_id": "<id>",
  "decision": "YES",
  "confidence": "HIGH",
  "reason": "Title and domain align with mechanical design profile.",
  "mismatch_flags": []
}
```

Rules:

1. Evaluate title first, then description.
2. If title domain conflicts with persona, return `NO` immediately.
3. Use one prompt/context per job (no batch comparison).
4. If uncertain, return `NO`.

## Stage 4: Final Output (No Ranking)

Return only jobs where:

- `decision = YES`
- `confidence = HIGH`

Display order can be retrieval order or newest-first, but never score-based rank.

Persist:

- `profiles/<user_id>/jobs_yes_high.json`
- `profiles/<user_id>/jobs_no_or_low.json`

## Stage 5: Seen-State Policy

Keep `--seen-update` during retrieval so processed jobs do not come back next run, including `NO` jobs.

## Implementation Plan

## Phase 1: Spec and Contracts

1. Update `skills/jobcli-cv-summary/SKILL.md` to produce `persona_summary.md` (plus JSON if needed).
2. Update `skills/jobcli-job-search/SKILL.md` to remove ranking language and enforce binary filter flow.
3. Define exact decision schema: `decision`, `confidence`, `reason`, `mismatch_flags`.

## Phase 2: Python Subagent Script

1. Add `skills/jobcli-job-search/scripts/job_discriminator.py`.
2. Accept persona summary + single job JSON input.
3. Return strict JSON only.
4. Add guardrails for domain mismatch (mechanical vs software false positives).

## Phase 3: Integration in Main Agent Workflow

1. Aggregate retrieved jobs.
2. Run script once per job (isolated contexts).
3. Keep only `YES/HIGH`.
4. Output filtered jobs without ranking.

## Phase 4: Validation

1. Test with mechanical-engineer persona and mixed job set.
2. Confirm software-heavy jobs are rejected with `NO/HIGH`.
3. Confirm matching mechanical roles pass with `YES/HIGH`.

## Success Criteria

1. Zero ranked-score output in job-search skill.
2. Binary decision per job (`YES/NO`) with confidence label.
3. Significant drop in cross-domain false positives (mechanical persona receiving software jobs).
4. Agent workflow stays simple: retrieve -> hard reject -> subagent gate -> output.
