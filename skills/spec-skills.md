# Skill Ranking Pipeline Plan (v2)

## Goal

Improve ranking quality so top jobs strongly match each user's persona profile and reduce false positives.

## Scope Analyzed

- `skills/jobcli-cv-summary/SKILL.md`
- `skills/jobcli-job-search/SKILL.md`
- `skills/pirate-motivator/SKILL.md`

## Current Gaps

1. Persona extraction is mostly narrative and not machine-scored.
2. Ranking uses fixed equal weights (0.2 each) without hard filters.
3. No explicit negative constraints (forbidden domains, seniority mismatch, onsite-only mismatch, etc.).
4. No evidence trace per score dimension, so quality is hard to debug.
5. No feedback loop from user outcomes (`apply`, `skip`, `interview`) to tune weights.
6. Seen-state timing is not explicit enough; it should guarantee non-matching jobs are not re-ranked on the next run.
7. `pirate-motivator` is useful for UX engagement but should not influence ranking logic.

## Target Architecture

Use a 2-stage pipeline:

1. **Retrieval stage**: maximize recall for relevant jobs.
2. **Ranking stage**: maximize precision with persona-aware scoring and hard constraints.

## Pipeline v2 (End-to-End)

## Stage 0: Persona Schema Upgrade

Extend `profiles/<user_id>/CVSUMMARY.md` with a machine-readable section:

```markdown
## Persona Profile v2

- Seniority Target: Mid-Senior
- Role Family: Backend, Platform
- Must-Have Skills: Go, SQL, APIs, Docker
- Preferred Skills: Kubernetes, AWS, Terraform
- Excluded Areas: Frontend-only, QA-only
- Work Mode: Remote or Hybrid
- Location Constraints: US, Canada
- Language Requirements: English (Fluent)
- Compensation Floor (optional): 130000 USD
```

Also write JSON mirror for stable parsing:

- `profiles/<user_id>/persona_profile.json`

## Stage 1: Query Expansion (Recall)

Generate 3 query buckets from persona:

1. Core role titles (8-10).
2. Skill-intent titles (6-8).
3. Domain/seniority variants (6-8).

Rules:

- Keep each query 2-5 words.
- Remove low-intent generic terms.
- De-duplicate semantically similar queries.

Output:

- `profiles/<user_id>/queries_v2.json`

## Stage 2: Retrieval Execution

Run sequentially per query with bounded recall:

```bash
jobcli search "<query>" --location "<location>" --country "<code>" --limit 30 \
  --seen profiles/<user_id>/jobs_seen.json --new-only --seen-update \
  --json --output profiles/<user_id>/jobs_new_keyword_<n>.json --hours 96
```

Changes vs current flow:

- Use `--seen-update` during retrieval so processed unseen jobs are not re-ranked next run.
- Increase freshness window to 96h only if volume is low.
- Keep per-query artifacts for diagnostics until ranking is complete.

## Stage 3: Aggregate and Normalize

Aggregate all per-query files into:

- `profiles/<user_id>/jobs_new_all.json`

Normalize before scoring:

1. Canonicalize URLs.
2. Trim and lowercase comparable text fields.
3. Dedupe by URL, then fallback key `(title, company, location)`.

## Stage 4: Hard Constraint Filter (Pre-Rank)

Reject jobs failing non-negotiables:

1. Seniority outside target band (large mismatch).
2. Work mode mismatch (`onsite-only` when persona is remote-only).
3. Location mismatch outside allowed geographies.
4. Language mismatch when explicit required language is unsupported.
5. Excluded role family/domain.

Output:

- `profiles/<user_id>/jobs_filtered_out.json` with `reject_reason`.

## Stage 5: Weighted Scoring Model

Score only surviving jobs with explicit evidence per dimension.
Ranking must compare persona against both job title and job description.

Scoring text policy:

1. Use `title + description` as primary evidence.
2. Cap description digest to 600 chars for token control.
3. If description is missing, fallback to `title` only.

### Scoring dimensions

1. Title-role fit: weight `0.25`
2. Description-persona fit: weight `0.25`
3. Must-have skill coverage (title + description + snippet): weight `0.20`
4. Preferred skill coverage: weight `0.10`
5. Seniority alignment: weight `0.10`
6. Work mode + location fit: weight `0.05`
7. Freshness signal: weight `0.05`

### Penalties

- Missing must-have cluster: `-0.15`
- Overqualified/underqualified strong mismatch: `-0.10`
- Ambiguous/very short posting data: `-0.05`
- Missing description (title-only fallback): `-0.05`

### Formula

`final_score = clamp(sum(weight_i * subscore_i) - penalties, 0.0, 1.0)`

Output per job:

- `score_breakdown`
- `matched_terms`
- `missing_must_haves`
- `penalties_applied`
- `confidence` (`high|medium|low`)
- `scoring_text_used` (title + description digest, or title-only fallback)

## Stage 6: Decision Bands

Use fixed action bands:

1. `>= 0.80`: strong match, recommended to apply.
2. `0.65 - 0.79`: review manually.
3. `< 0.65`: low fit, hide by default.

Allow user override threshold.

## Stage 7: Output Contract

Return ranked results in a structured markdown table plus compact cards.

Required fields:

1. Rank
2. Title
3. Company
4. Location
5. Final score
6. Top 3 match reasons
7. Top 2 mismatch reasons
8. URL

Persist optional report:

- `profiles/<user_id>/ranked_jobs.md`
- `profiles/<user_id>/ranked_jobs.json`

## Stage 8: Seen-State Update Strategy

Update `jobs_seen.json` during retrieval, not after ranking.

Policy:

1. Mark discovered unseen jobs immediately per query via `--seen-update`.
2. This guarantees jobs that did not match well in this run are not re-ranked in the next run.
3. Keep an optional short re-check list for near-threshold jobs (`0.60 - 0.69`) for 7 days only when explicitly enabled.

Artifact:

- `profiles/<user_id>/jobs_recheck.json` (optional)

## Stage 9: Feedback Learning Loop

Capture user outcomes:

- `applied`
- `interview`
- `rejected`
- `not_relevant`

Persist to:

- `profiles/<user_id>/ranking_feedback.json`

Weekly adaptation:

1. Increase weights on dimensions correlated with `applied/interview`.
2. Increase penalties tied to repeated `not_relevant` patterns.
3. Regenerate query buckets from top-performing titles.

## Stage 10: Motivation Integration (Non-Scoring)

`pirate-motivator` remains optional post-output UX:

1. If user got low results count, send motivational audio/text.
2. Do not use motivation skill outputs as ranking inputs.

## Implementation Plan (Practical)

## Phase 1: Skill Spec Upgrades

1. Update `skills/jobcli-cv-summary/SKILL.md` to emit `Persona Profile v2` + `persona_profile.json`.
2. Update `skills/jobcli-job-search/SKILL.md` for staged retrieval + immediate seen update (`--seen-update`).

## Phase 2: Ranking Engine Contract

1. Define scoring schema in markdown + JSON examples.
2. Add deterministic scoring rules and penalty logic.
3. Add score evidence requirements in output format.

## Phase 3: Feedback and Calibration

1. Add feedback file contract.
2. Add weekly tuning heuristics.
3. Add quality metrics report.

## Success Metrics

Measure per user over rolling 2 weeks:

1. Precision@10 (target: +30% from current baseline).
2. User "relevant" rate on top 20 (target: >= 70%).
3. False-positive rate in top 10 (target: <= 20%).
4. Median manual review time per run (target: -25%).

## Risk Controls

1. If job descriptions are sparse, lower confidence and push to manual band.
2. Keep rule-based fallback if structured persona JSON is missing.
3. Keep all ranking decisions explainable with reasons and penalties.

## Immediate Next Actions

1. Align both skills on the new `Persona Profile v2` contract.
2. Implement immediate `seen` update semantics in the job-search skill.
3. Introduce score breakdown + confidence in ranked output.
4. Start collecting feedback labels for calibration.
