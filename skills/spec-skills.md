# JobCLI Skills Spec (Minimal)

## Purpose
Define the required contract for the JobCLI skill workflow: build persona files, then run strict YES/HIGH job filtering.

## Skills In Scope
- `skills/jobcli-cv-summary/SKILL.md`
- `skills/jobcli-job-search/SKILL.md`
- `skills/jobcli-job-search/scripts/job_discriminator.py`

## Required User-Scoped Files
- `profiles/<user_id>/resume.pdf`
- `profiles/<user_id>/CVSUMMARY.md`
- `profiles/<user_id>/persona_profile.json`
- `profiles/<user_id>/jobs_seen.json`
- `profiles/<user_id>/jobs_filtered_out.json`
- `profiles/<user_id>/jobs_yes_high.json`

## Non-Negotiable Rules
1. No ranking, no score thresholds, no weighted formulas.
2. LLM decision is binary per job: `YES` or `NO`.
3. Keep only `YES` with `HIGH` confidence.
4. If uncertain, return `NO`.
5. Use one isolated LLM context per job.
6. Never mix files across users.

## Minimal End-to-End Flow
1. Build persona artifacts from CV using `jobcli-cv-summary`.
2. Retrieve unseen jobs with `jobcli search` and `--seen-update`.
3. Aggregate results into `profiles/<user_id>/jobs_new_all.json`.
4. Apply deterministic hard rejects (role/domain, seniority, work mode, location).
5. Run discriminator:

```bash
python3 skills/jobcli-job-search/scripts/job_discriminator.py \
  --cvsummary profiles/<user_id>/CVSUMMARY.md \
  --jobs-json profiles/<user_id>/jobs_new_all.json \
  --output profiles/<user_id>/jobs_yes_high.json
```

6. Return only jobs from `jobs_yes_high.json`.
7. Keep persistent artifacts; remove temporary `jobs_new_keyword_*.json` and `jobs_new_all.json`.

## Done Criteria
- `CVSUMMARY.md` and `persona_profile.json` exist per user.
- `jobs_yes_high.json` contains only accepted jobs.
- Rejected jobs do not reappear due to seen tracking.
