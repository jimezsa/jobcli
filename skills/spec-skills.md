# JobCLI Skills Spec (Minimal)

## Purpose
Define the required contract for the JobCLI skill workflow: build persona files, then run YES job filtering with a configurable confidence threshold.

## Skills In Scope
- `skills/jobcli-cv-summary/SKILL.md`
- `skills/jobcli-job-search/SKILL.md`
- `skills/jobcli-job-search/scripts/job_discriminator.py`

## Required User-Scoped Files
- `profiles/<user_id>/resume.pdf`
- `profiles/<user_id>/persona_querie.json`
- `profiles/<user_id>/jobs_seen.json`
- `profiles/<user_id>/jobs_filtered_out.json`
- `profiles/<user_id>/jobs_yes_high.json`

## Non-Negotiable Rules
1. No ranking, no score thresholds, no weighted formulas.
2. LLM decision is binary per job: `YES` or `NO`.
3. Keep only `YES` with confidence at or above `--min-confidence` (`LOW` by default).
4. Use `--min-confidence HIGH` for strict-only matches.
5. Use one isolated LLM context per job.
6. Never mix files across users.

## Minimal End-to-End Flow
1. Build persona artifacts from CV using `jobcli-cv-summary`.
2. Retrieve unseen jobs with `jobcli search --query-file profiles/<user_id>/persona_querie.json` and `--seen-update`.
3. Apply deterministic hard rejects (role/domain, seniority, work mode, location).
4. Run discriminator:

```bash
python3 skills/jobcli-job-search/scripts/job_discriminator.py \
  --persona-json profiles/<user_id>/persona_querie.json \
  --jobs-json profiles/<user_id>/jobs_new_all.json \
  --min-confidence LOW \
  --output profiles/<user_id>/jobs_yes_high.json
```

5. Return only jobs from `jobs_yes_high.json`.
6. Keep persistent artifacts.

## Done Criteria
- `persona_querie.json` exists per user.
- `jobs_yes_high.json` contains only accepted jobs.
- Rejected jobs do not reappear due to seen tracking.
